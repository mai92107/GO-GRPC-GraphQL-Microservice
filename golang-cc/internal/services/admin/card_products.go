package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) ListCards(ctx context.Context) ([]domain.Card, error) {
	return s.repository.ListCards(ctx)
}

func (s *Service) GetCardInfo(ctx context.Context, id string) (domain.CardInfo, error) {
	return s.repository.GetCardInfo(ctx, id)
}

func (s *Service) ListNetworks(ctx context.Context) ([]string, error) {
	return s.repository.ListNetworks(ctx)
}

func (s *Service) CreateCard(ctx context.Context, input CardProductInput) (string, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.QualifiedType = normalizeCardType(input.QualifiedType)
	input.SelectableType = normalizeCardType(input.SelectableType)
	input.Networks = normalizeNetworkNames(input.Networks)
	if input.BankID == "" || input.Name == "" || len(input.Networks) == 0 {
		return "", domain.ErrInvalidInput
	}
	id := secure.UUID()
	return id, s.repository.CreateCard(ctx, id, domain.CardInput(input))
}

func (s *Service) UpdateCard(ctx context.Context, id string, input CardProductInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.QualifiedType = normalizeCardType(input.QualifiedType)
	input.SelectableType = normalizeCardType(input.SelectableType)
	input.Networks = normalizeNetworkNames(input.Networks)
	if id == "" || input.BankID == "" || input.Name == "" || len(input.Networks) == 0 {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateCard(ctx, id, domain.CardInput(input))
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

func normalizeNetworkNames(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func (s *Service) DeleteCard(ctx context.Context, id string) error {
	return s.repository.DeleteCard(ctx, id)
}
