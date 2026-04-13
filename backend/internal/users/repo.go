package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// Upsert creates the user if they don't exist, or returns the existing record.
func (r *Repo) Upsert(ctx context.Context, firebaseUID, displayName, phoneHash string) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (firebase_uid, display_name, phone_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (firebase_uid) DO UPDATE SET firebase_uid = EXCLUDED.firebase_uid
		RETURNING id, firebase_uid, display_name, phone_hash, created_at
	`, firebaseUID, displayName, phoneHash).Scan(
		&u.ID, &u.FirebaseUID, &u.DisplayName, &u.PhoneHash, &u.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("upsert user: %w", err)
	}
	return u, nil
}

// GetByFirebaseUID looks up a user by their Firebase UID.
func (r *Repo) GetByFirebaseUID(ctx context.Context, firebaseUID string) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT id, firebase_uid, display_name, phone_hash, created_at
		FROM users WHERE firebase_uid = $1
	`, firebaseUID).Scan(&u.ID, &u.FirebaseUID, &u.DisplayName, &u.PhoneHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by firebase uid: %w", err)
	}
	return u, nil
}

// GetByID looks up a user by their internal UUID.
func (r *Repo) GetByID(ctx context.Context, id string) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT id, firebase_uid, display_name, phone_hash, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.FirebaseUID, &u.DisplayName, &u.PhoneHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// AreFriends returns true if userA and userB have mutually added each other's phone hash.
func (r *Repo) AreFriends(ctx context.Context, userAID, userBID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM contacts c_a
		JOIN contacts c_b ON TRUE
		WHERE c_a.owner_id = $1
		  AND c_a.phone_hash = (SELECT phone_hash FROM users WHERE id = $2)
		  AND c_b.owner_id = $2
		  AND c_b.phone_hash = (SELECT phone_hash FROM users WHERE id = $1)
	`, userAID, userBID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("are friends: %w", err)
	}
	return count > 0, nil
}
