package groups

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielyakubu/circles/internal/friends"
)

type Service struct {
	repo        *Repo
	friendsSvc  *friends.Service
}

func NewService(repo *Repo, friendsSvc *friends.Service) *Service {
	return &Service{repo: repo, friendsSvc: friendsSvc}
}

func (s *Service) Create(ctx context.Context, name, creatorID string) (Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Group{}, fmt.Errorf("name is required")
	}

	g, err := s.repo.Create(ctx, name, creatorID)
	if err != nil {
		return Group{}, err
	}

	// Auto-add creator as first member.
	if err := s.repo.AddMember(ctx, g.ID, creatorID); err != nil {
		return Group{}, err
	}

	return g, nil
}

func (s *Service) AddMember(ctx context.Context, groupID, requesterID, candidateID string) error {
	// Requester must be a member.
	ok, err := s.repo.IsMember(ctx, groupID, requesterID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	// Candidate must be a friend of the requester.
	areFriends, err := s.friendsSvc.AreFriends(ctx, requesterID, candidateID)
	if err != nil {
		return err
	}
	if !areFriends {
		return ErrNotFriend
	}

	// Circle check: candidate must be friends with every current member.
	fails, err := s.repo.CircleCheckFails(ctx, groupID, candidateID)
	if err != nil {
		return err
	}
	if fails {
		return ErrCircleCheck
	}

	return s.repo.AddMember(ctx, groupID, candidateID)
}

func (s *Service) Leave(ctx context.Context, groupID, userID string) error {
	return s.repo.RemoveMember(ctx, groupID, userID)
}

func (s *Service) ListForUser(ctx context.Context, userID string) ([]Group, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *Service) ListMembers(ctx context.Context, groupID, requesterID string) ([]Member, error) {
	ok, err := s.repo.IsMember(ctx, groupID, requesterID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return s.repo.ListMembers(ctx, groupID)
}
