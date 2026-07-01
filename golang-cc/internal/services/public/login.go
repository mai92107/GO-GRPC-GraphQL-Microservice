package public

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	result, err := s.repository.FindLoginUser(ctx, email)
	if err != nil || result.Status != "active" || !VerifyPassword(result.PasswordHash, password) {
		println(err.Error())
		return domain.User{}, "", domain.ErrInvalidCredentials
	}
	raw, hash, err := secure.Token(32)
	if err != nil {
		return domain.User{}, "", err
	}
	if err = s.repository.CreateSession(ctx, secure.UUID(), result.ID, hash, s.now().Add(SessionDuration)); err != nil {
		return domain.User{}, "", err
	}
	return result.User, raw, nil
}
