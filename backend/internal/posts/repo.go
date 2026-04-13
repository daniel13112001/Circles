package posts

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND user_id = $2
	`, groupID, userID).Scan(&count)
	return count > 0, err
}

func (r *Repo) Create(ctx context.Context, groupID, authorID, imageURL string, caption *string) (Post, error) {
	var p Post
	var authorDisplayName string

	err := r.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO posts (group_id, author_id, image_url, caption)
			VALUES ($1, $2, $3, $4)
			RETURNING id, group_id, author_id, image_url, caption, created_at
		)
		SELECT i.id, i.group_id, i.author_id, u.display_name, i.image_url, i.caption, i.created_at
		FROM inserted i
		JOIN users u ON u.id = i.author_id
	`, groupID, authorID, imageURL, caption).Scan(
		&p.ID, &p.GroupID, &p.Author.ID, &authorDisplayName,
		&p.ImageURL, &p.Caption, &p.CreatedAt,
	)
	if err != nil {
		return Post{}, fmt.Errorf("create post: %w", err)
	}
	p.Author.DisplayName = authorDisplayName
	return p, nil
}

func (r *Repo) GroupFeed(ctx context.Context, groupID string) ([]Post, error) {
	return r.queryFeed(ctx, `
		SELECT p.id, p.group_id, p.author_id, u.display_name, p.image_url, p.caption, p.created_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.group_id = $1
		ORDER BY p.created_at DESC
	`, groupID)
}

func (r *Repo) GlobalFeed(ctx context.Context, userID string) ([]Post, error) {
	return r.queryFeed(ctx, `
		SELECT p.id, p.group_id, p.author_id, u.display_name, p.image_url, p.caption, p.created_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN group_members gm ON gm.group_id = p.group_id AND gm.user_id = $1
		ORDER BY p.created_at DESC
	`, userID)
}

func (r *Repo) queryFeed(ctx context.Context, sql string, arg string) ([]Post, error) {
	rows, err := r.db.Query(ctx, sql, arg)
	if err != nil {
		return nil, fmt.Errorf("query feed: %w", err)
	}
	defer rows.Close()

	var result []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(
			&p.ID, &p.GroupID, &p.Author.ID, &p.Author.DisplayName,
			&p.ImageURL, &p.Caption, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
