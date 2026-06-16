package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
)

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repository.ListCategories(ctx)
}
func (s *Service) CreateCategory(ctx context.Context, code, name string) error {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.CreateCategory(ctx, strings.TrimSpace(code), strings.TrimSpace(name))
}
func (s *Service) UpdateCategory(ctx context.Context, code, name string, active bool) error {
	if code == "" || strings.TrimSpace(name) == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateCategory(ctx, code, strings.TrimSpace(name), active)
}
func (s *Service) DeleteCategory(ctx context.Context, code string) error {
	return s.repository.DeleteCategory(ctx, code)
}
