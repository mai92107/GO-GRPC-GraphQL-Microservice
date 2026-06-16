package public

import (
	"context"

	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) ResetPassword(ctx context.Context, token, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.repository.ResetPassword(ctx, secure.Hash(token), hash, s.now())
}
