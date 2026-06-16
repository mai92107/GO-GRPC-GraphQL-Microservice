package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) ListBanks(ctx context.Context) ([]domain.Bank, error) {
	return s.repository.ListBanks(ctx)
}

func (s *Service) CreateBank(ctx context.Context, input BankInput) (string, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return "", domain.ErrInvalidInput
	}
	id := secure.UUID()
	return id, s.repository.CreateBank(ctx, id, domain.BankInput(input))
}

func (s *Service) UpdateBank(ctx context.Context, id string, input BankInput) error {
	input.Name = strings.TrimSpace(input.Name)
	if id == "" || input.Name == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateBank(ctx, id, domain.BankInput(input))
}

func (s *Service) DeleteBank(ctx context.Context, id string) error {
	return s.repository.DeleteBank(ctx, id)
}
