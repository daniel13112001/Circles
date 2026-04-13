package users

import (
	"context"
	"fmt"
	"strings"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// Register creates or retrieves the user. Idempotent.
func (s *Service) Register(ctx context.Context, firebaseUID, displayName, phoneHash string) (User, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return User{}, fmt.Errorf("display_name is required")
	}
	if phoneHash == "" {
		return User{}, fmt.Errorf("phone_hash is required")
	}
	return s.repo.Upsert(ctx, firebaseUID, displayName, phoneHash)
}

// GetMe returns the authenticated user's own profile.
func (s *Service) GetMe(ctx context.Context, userID string) (User, error) {
	return s.repo.GetByID(ctx, userID)
}

// GetFriend returns the profile of targetID, but only if they are friends with requesterID.
func (s *Service) GetFriend(ctx context.Context, requesterID, targetID string) (User, error) {
	if requesterID == targetID {
		return s.repo.GetByID(ctx, requesterID)
	}

	ok, err := s.repo.AreFriends(ctx, requesterID, targetID)
	if err != nil {
		return User{}, err
	}
	if !ok {
		return User{}, ErrForbidden
	}

	return s.repo.GetByID(ctx, targetID)
}

// GetByFirebaseUID is used by middleware to resolve a Firebase UID to an internal user.
func (s *Service) GetByFirebaseUID(ctx context.Context, firebaseUID string) (User, error) {
	return s.repo.GetByFirebaseUID(ctx, firebaseUID)
}
