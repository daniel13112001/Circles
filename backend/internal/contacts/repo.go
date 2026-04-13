package contacts

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// Add inserts a new contact entry for ownerID with the given phone hash.
func (r *Repo) Add(ctx context.Context, ownerID, phoneHash string) (Contact, error) {
	var c Contact
	err := r.db.QueryRow(ctx, `
		INSERT INTO contacts (owner_id, phone_hash)
		VALUES ($1, $2)
		RETURNING id, phone_hash, created_at
	`, ownerID, phoneHash).Scan(&c.ID, &c.PhoneHash, &c.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Contact{}, ErrDuplicate
	}
	if err != nil {
		return Contact{}, fmt.Errorf("add contact: %w", err)
	}
	return c, nil
}

// List returns all contacts for ownerID, joining against users to resolve display names.
func (r *Repo) List(ctx context.Context, ownerID string) ([]Contact, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			c.id,
			c.phone_hash,
			u.display_name,
			(u.id IS NOT NULL) AS matched,
			c.created_at
		FROM contacts c
		LEFT JOIN users u ON u.phone_hash = c.phone_hash
		WHERE c.owner_id = $1
		ORDER BY c.created_at DESC
	`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	defer rows.Close()

	var result []Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.PhoneHash, &c.DisplayName, &c.Matched, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// Delete removes a contact by ID, but only if it belongs to ownerID.
func (r *Repo) Delete(ctx context.Context, ownerID, contactID string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM contacts WHERE id = $1 AND owner_id = $2
	`, contactID, ownerID)
	if err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// OwnerPhoneHash returns the phone_hash of the given user (used for self-add check).
func (r *Repo) OwnerPhoneHash(ctx context.Context, ownerID string) (string, error) {
	var hash string
	err := r.db.QueryRow(ctx, `SELECT phone_hash FROM users WHERE id = $1`, ownerID).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("user not found")
	}
	if err != nil {
		return "", fmt.Errorf("get phone hash: %w", err)
	}
	return hash, nil
}
