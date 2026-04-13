package friends

import "context"

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID string) ([]Friend, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) AreFriends(ctx context.Context, userAID, userBID string) (bool, error) {
	return s.repo.AreFriends(ctx, userAID, userBID)
}
