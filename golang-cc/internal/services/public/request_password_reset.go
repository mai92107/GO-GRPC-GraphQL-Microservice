package public

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	userID, actualEmail, err := s.repository.ActiveUserByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	raw, hash, err := secure.Token(32)
	if err != nil {
		return err
	}
	if err = s.repository.CreatePasswordReset(ctx, secure.UUID(), userID, hash, s.now().Add(time.Hour)); err != nil {
		return err
	}
	return s.email.SendPasswordReset(ctx, actualEmail, s.publicBaseURL+"/reset-password?token="+url.QueryEscape(raw))
}
