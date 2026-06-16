package admin

import (
	"context"
	"regexp"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
)

var merchantCodePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

func (s *Service) ListMerchants(ctx context.Context) ([]domain.Merchant, error) {
	return s.repository.ListMerchants(ctx)
}

func (s *Service) CreateMerchant(ctx context.Context, code, name string, aliases, categories []string) error {
	item, ok := validMerchant(code, name, aliases, categories, true)
	if !ok {
		return domain.ErrInvalidInput
	}
	return s.repository.CreateMerchant(ctx, item)
}

func (s *Service) UpdateMerchant(ctx context.Context, code, name string, aliases, categories []string, active bool) error {
	item, ok := validMerchant(code, name, aliases, categories, active)
	if !ok {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateMerchant(ctx, item)
}

func (s *Service) DeleteMerchant(ctx context.Context, code string) error {
	return s.repository.DeleteMerchant(ctx, code)
}

func validMerchant(code, name string, aliases, categories []string, active bool) (domain.Merchant, bool) {
	code, name = strings.TrimSpace(code), strings.TrimSpace(name)
	if !merchantCodePattern.MatchString(code) || name == "" {
		return domain.Merchant{}, false
	}
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
	return domain.Merchant{Code: code, Name: name, Aliases: clean, CategoryCodes: categoryClean, IsActive: active}, true
}
