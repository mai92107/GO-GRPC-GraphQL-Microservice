package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repository.ListCategories(ctx)
}
func (s *Service) CreateCategory(ctx context.Context, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", domain.ErrInvalidInput
	}
	id := secure.UUID()
	return id, s.repository.CreateCategory(ctx, id, strings.TrimSpace(name))
}
func (s *Service) UpdateCategory(ctx context.Context, id, name string, active bool) error {
	if id == "" || strings.TrimSpace(name) == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateCategory(ctx, id, strings.TrimSpace(name), active)
}
func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	return s.repository.DeleteCategory(ctx, id)
}
