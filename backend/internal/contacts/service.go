package contacts

import "context"

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, ownerID, phoneHash string) (Contact, error) {
	// Prevent adding own hash.
	ownHash, err := s.repo.OwnerPhoneHash(ctx, ownerID)
	if err != nil {
		return Contact{}, err
	}
	if phoneHash == ownHash {
		return Contact{}, ErrSelfAdd
	}

	return s.repo.Add(ctx, ownerID, phoneHash)
}

func (s *Service) List(ctx context.Context, ownerID string) ([]Contact, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Delete(ctx context.Context, ownerID, contactID string) error {
	return s.repo.Delete(ctx, ownerID, contactID)
}
