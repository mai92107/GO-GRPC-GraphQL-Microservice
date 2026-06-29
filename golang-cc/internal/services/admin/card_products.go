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

func (s *Service) GetCardProduct(ctx context.Context, id string, includeActivities bool) (domain.CardProduct, error) {
	return s.repository.GetCardProduct(ctx, id, includeActivities)
}

func (s *Service) ListCardNetworks(ctx context.Context) ([]string, error) {
	return s.repository.ListCardNetworks(ctx)
}

func (s *Service) CreateCardProduct(ctx context.Context, input CardProductInput) (string, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.AccountTiers = normalizeTiers(input.AccountTiers)
	input.QualifiedType = normalizeCardType(input.QualifiedType)
	input.SelectableType = normalizeCardType(input.SelectableType)
	input.NetworkIDs = normalizeIDs(input.NetworkIDs)
	if input.BankID == "" || input.Name == "" || len(input.NetworkIDs) == 0 {
		return "", domain.ErrInvalidInput
	}
	id := secure.UUID()
	return id, s.repository.CreateCardProduct(ctx, id, domain.CardProductInput(input))
}

func (s *Service) UpdateCardProduct(ctx context.Context, id string, input CardProductInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.AccountTiers = normalizeTiers(input.AccountTiers)
	input.QualifiedType = normalizeCardType(input.QualifiedType)
	input.SelectableType = normalizeCardType(input.SelectableType)
	input.NetworkIDs = normalizeIDs(input.NetworkIDs)
	if id == "" || input.BankID == "" || input.Name == "" || len(input.NetworkIDs) == 0 {
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

func normalizeCardType(value string) string {
	return strings.Join(domain.SplitCardType(value), ",")
}

func normalizeIDs(values []string) []string {
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
