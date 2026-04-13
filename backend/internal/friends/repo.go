package friends

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

// List returns all users who are mutual contacts with userID.
func (r *Repo) List(ctx context.Context, userID string) ([]Friend, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.display_name, u.created_at
		FROM users u
		JOIN contacts c_me   ON c_me.phone_hash  = u.phone_hash
		                     AND c_me.owner_id   = $1
		JOIN contacts c_them ON c_them.phone_hash = (SELECT phone_hash FROM users WHERE id = $1)
		                     AND c_them.owner_id  = u.id
		WHERE u.id != $1
		ORDER BY u.display_name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list friends: %w", err)
	}
	defer rows.Close()

	var result []Friend
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.DisplayName, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan friend: %w", err)
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

// AreFriends returns true if userA and userB have mutually added each other.
// This is the canonical implementation used by groups and users domains.
func (r *Repo) AreFriends(ctx context.Context, userAID, userBID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM contacts c_a
		JOIN contacts c_b ON TRUE
		WHERE c_a.owner_id  = $1
		  AND c_a.phone_hash = (SELECT phone_hash FROM users WHERE id = $2)
		  AND c_b.owner_id  = $2
		  AND c_b.phone_hash = (SELECT phone_hash FROM users WHERE id = $1)
	`, userAID, userBID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("are friends: %w", err)
	}
	return count > 0, nil
}
