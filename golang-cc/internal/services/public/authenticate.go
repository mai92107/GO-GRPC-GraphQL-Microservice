package public

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) Authenticate(ctx context.Context, token string) (domain.User, error) {
	user, err := s.repository.Authenticate(ctx, secure.Hash(token), s.now())
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}
