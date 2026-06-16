package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) ListCardProducts(ctx context.Context) ([]domain.CardProduct, error) {
	return s.repository.ListCardProducts(ctx)
}

func (s *Service) CreateCardProduct(ctx context.Context, input CardProductInput) (string, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.AccountTiers = normalizeTiers(input.AccountTiers)
	if input.BankID == "" || input.Name == "" {
		return "", domain.ErrInvalidInput
	}
	id := secure.UUID()
	return id, s.repository.CreateCardProduct(ctx, id, domain.CardProductInput(input))
}

func (s *Service) UpdateCardProduct(ctx context.Context, id string, input CardProductInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.AccountTiers = normalizeTiers(input.AccountTiers)
	if id == "" || input.BankID == "" || input.Name == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateCardProduct(ctx, id, domain.CardProductInput(input))
}

func normalizeTiers(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func (s *Service) DeleteCardProduct(ctx context.Context, id string) error {
	return s.repository.DeleteCardProduct(ctx, id)
}
