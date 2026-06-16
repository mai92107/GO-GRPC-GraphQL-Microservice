package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
)

func (s *Service) CreateInvitation(ctx context.Context, email, invitedBy string) error {
	if !strings.Contains(email, "@") {
		return domain.ErrInvalidInput
	}
	return s.auth.CreateInvitation(ctx, strings.TrimSpace(email), invitedBy)
}

func (s *Service) ListInvitations(ctx context.Context) ([]domain.Invitation, error) {
	return s.repository.ListInvitations(ctx)
}

func (s *Service) DeleteInvitation(ctx context.Context, id string) error {
	return s.repository.DeleteInvitation(ctx, id)
}
