package groups

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

func (r *Repo) Create(ctx context.Context, name, createdBy string) (Group, error) {
	var g Group
	err := r.db.QueryRow(ctx, `
		INSERT INTO groups (name, created_by) VALUES ($1, $2)
		RETURNING id, name, created_by, created_at
	`, name, createdBy).Scan(&g.ID, &g.Name, &g.CreatedBy, &g.CreatedAt)
	if err != nil {
		return Group{}, fmt.Errorf("create group: %w", err)
	}
	return g, nil
}

func (r *Repo) AddMember(ctx context.Context, groupID, userID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)
	`, groupID, userID)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrAlreadyMember
	}
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *Repo) RemoveMember(ctx context.Context, groupID, userID string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM group_members WHERE group_id = $1 AND user_id = $2
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND user_id = $2
	`, groupID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("is member: %w", err)
	}
	return count > 0, nil
}

func (r *Repo) ListForUser(ctx context.Context, userID string) ([]Group, error) {
	rows, err := r.db.Query(ctx, `
		SELECT g.id, g.name, g.created_by, g.created_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = $1
		ORDER BY g.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var result []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.CreatedBy, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

func (r *Repo) ListMembers(ctx context.Context, groupID string) ([]Member, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.display_name, gm.joined_at
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var result []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.DisplayName, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *Repo) GetByID(ctx context.Context, groupID string) (Group, error) {
	var g Group
	err := r.db.QueryRow(ctx, `
		SELECT id, name, created_by, created_at FROM groups WHERE id = $1
	`, groupID).Scan(&g.ID, &g.Name, &g.CreatedBy, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Group{}, ErrNotFound
	}
	if err != nil {
		return Group{}, fmt.Errorf("get group: %w", err)
	}
	return g, nil
}

// CircleCheckFails returns true if candidateID is NOT friends with every
// current member of groupID. A result of true means the add should be rejected.
func (r *Repo) CircleCheckFails(ctx context.Context, groupID, candidateID string) (bool, error) {
	var nonFriendCount int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM group_members gm
		WHERE gm.group_id = $1
		  AND NOT EXISTS (
		    SELECT 1
		    FROM contacts c_candidate
		    JOIN contacts c_member ON TRUE
		    WHERE c_candidate.owner_id  = $2
		      AND c_candidate.phone_hash = (SELECT phone_hash FROM users WHERE id = gm.user_id)
		      AND c_member.owner_id     = gm.user_id
		      AND c_member.phone_hash   = (SELECT phone_hash FROM users WHERE id = $2)
		  )
	`, groupID, candidateID).Scan(&nonFriendCount)
	if err != nil {
		return false, fmt.Errorf("circle check: %w", err)
	}
	return nonFriendCount > 0, nil
}
