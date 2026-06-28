package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
)

func (s *Service) ListMerchants(ctx context.Context) ([]domain.Merchant, error) {
	return s.repository.ListMerchants(ctx)
}

func (s *Service) CreateMerchant(ctx context.Context, name string, aliases, categories []string) (string, error) {
	item, ok := validMerchant(name, aliases, categories, true)
	if !ok {
		return "", domain.ErrInvalidInput
	}
	return s.repository.CreateMerchant(ctx, item)
}

func (s *Service) UpdateMerchant(ctx context.Context, id, name string, aliases, categories []string, active bool) error {
	item, ok := validMerchant(name, aliases, categories, active)
	if !ok {
		return domain.ErrInvalidInput
	}
	item.Id = strings.TrimSpace(id)
	if item.Id == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateMerchant(ctx, item)
}

func (s *Service) DeleteMerchant(ctx context.Context, id string) error {
	return s.repository.DeleteMerchant(ctx, id)
}

func validMerchant(name string, aliases, categories []string, active bool) (domain.Merchant, bool) {
	name = strings.TrimSpace(name)
	seen, clean := map[string]bool{}, []string{}
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		key := strings.ToLower(alias)
		if alias == "" || seen[key] {
			return domain.Merchant{}, false
		}
		seen[key] = true
		clean = append(clean, alias)
	}
	categorySeen, categoryClean := map[string]bool{}, []string{}
	for _, category := range categories {
		category = strings.TrimSpace(category)
		if category == "" || categorySeen[category] {
			return domain.Merchant{}, false
		}
		categorySeen[category] = true
		categoryClean = append(categoryClean, category)
	}
	if len(categoryClean) == 0 {
		return domain.Merchant{}, false
	}
	return domain.Merchant{Name: name, Aliases: clean, CategoryIDs: categoryClean, IsActive: active}, true
}
