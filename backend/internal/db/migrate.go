package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrationsFS holds all SQL migration files, embedded at compile time.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// Migrate runs all embedded SQL migration files in order.
// It creates a simple schema_migrations table to track applied migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	// Ensure tracking table exists.
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		var count int
		err := pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM schema_migrations WHERE filename = $1`, name,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if count > 0 {
			continue
		}

		content, err := fs.ReadFile(MigrationsFS, "migrations/"+name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		sql := upSection(string(content))

		if _, err := pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("run migration %s: %w", name, err)
		}

		if _, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations (filename) VALUES ($1)`, name,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}

	return nil
}

// upSection returns the SQL between "-- +migrate Up" and "-- +migrate Down".
func upSection(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	inUp := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "-- +migrate Up" {
			inUp = true
			continue
		}
		if trimmed == "-- +migrate Down" {
			break
		}
		if inUp {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}
