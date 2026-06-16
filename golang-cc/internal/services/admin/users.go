package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (s *Service) ListUsers(ctx context.Context) ([]domain.AdminUser, error) {
	return s.repository.ListUsers(ctx)
}

func (s *Service) UpdateUserStatus(ctx context.Context, id, currentUserID, status string) error {
	if status != "active" && status != "disabled" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateUserStatus(ctx, id, currentUserID, status)
}

func (s *Service) SendUserPasswordReset(ctx context.Context, id string) error {
	email, err := s.repository.UserEmail(ctx, id)
	if err != nil {
		return err
	}
	return s.auth.RequestPasswordReset(ctx, email)
}
