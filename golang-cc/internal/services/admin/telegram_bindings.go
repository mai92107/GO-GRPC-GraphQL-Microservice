package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (s *Service) ListTelegramBindings(ctx context.Context) ([]domain.TelegramBinding, error) {
	return s.repository.ListTelegramBindings(ctx)
}

func (s *Service) CreateTelegramBinding(ctx context.Context, chatID int64, userID string) error {
	if chatID == 0 || userID == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.CreateTelegramBinding(ctx, chatID, userID)
}

func (s *Service) DeleteTelegramBinding(ctx context.Context, chatID int64) error {
	if chatID == 0 {
		return domain.ErrInvalidInput
	}
	return s.repository.DeleteTelegramBinding(ctx, chatID)
}
