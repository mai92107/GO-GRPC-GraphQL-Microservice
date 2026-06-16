package public

import (
	"context"

	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.repository.DeleteSession(ctx, secure.Hash(token))
}
