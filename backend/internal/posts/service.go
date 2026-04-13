package posts

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

func (s *Service) Create(ctx context.Context, groupID, authorID, imageURL string, caption *string) (Post, error) {
	ok, err := s.repo.IsMember(ctx, groupID, authorID)
	if err != nil {
		return Post{}, err
	}
	if !ok {
		return Post{}, ErrForbidden
	}

	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return Post{}, fmt.Errorf("image_url is required")
	}

	return s.repo.Create(ctx, groupID, authorID, imageURL, caption)
}

func (s *Service) GroupFeed(ctx context.Context, groupID, userID string) ([]Post, error) {
	ok, err := s.repo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return s.repo.GroupFeed(ctx, groupID)
}

func (s *Service) GlobalFeed(ctx context.Context, userID string) ([]Post, error) {
	return s.repo.GlobalFeed(ctx, userID)
}
