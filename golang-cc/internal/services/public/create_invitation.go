package public

import (
	"context"
	"net/url"
	"time"

	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) CreateInvitation(ctx context.Context, email, invitedBy string) error {
	raw, hash, err := secure.Token(32)
	if err != nil {
		return err
	}
	if err = s.repository.CreateInvitation(ctx, secure.UUID(), email, hash, invitedBy, s.now().Add(7*24*time.Hour)); err != nil {
		return err
	}
	return s.email.SendInvitation(ctx, email, s.publicBaseURL+"/accept-invitation?token="+url.QueryEscape(raw))
}
