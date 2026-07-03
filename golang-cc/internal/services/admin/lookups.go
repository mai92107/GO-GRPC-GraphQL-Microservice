package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
)

func (s *Service) ListRegions(ctx context.Context) ([]domain.LookupItem, error) {
	return s.repository.ListRegions(ctx)
}

func (s *Service) CreateRegion(ctx context.Context, id, name string) error {
	item, err := prepareLookupItem(id, name, true)
	if err != nil {
		return err
	}
	return s.repository.CreateRegion(ctx, item)
}

func (s *Service) UpdateRegion(ctx context.Context, id, name string, active bool) error {
	item, err := prepareLookupItem(id, name, active)
	if err != nil {
		return err
	}
	return s.repository.UpdateRegion(ctx, item)
}

func (s *Service) DeleteRegion(ctx context.Context, id string) error {
	return s.repository.DeleteRegion(ctx, strings.TrimSpace(id))
}

func (s *Service) ListUserQualifications(ctx context.Context) ([]domain.LookupItem, error) {
	return s.repository.ListUserQualifications(ctx)
}

func (s *Service) CreateUserQualification(ctx context.Context, id, name string) error {
	item, err := prepareLookupItem(id, name, true)
	if err != nil {
		return err
	}
	return s.repository.CreateUserQualification(ctx, item)
}

func (s *Service) UpdateUserQualification(ctx context.Context, id, name string, active bool) error {
	item, err := prepareLookupItem(id, name, active)
	if err != nil {
		return err
	}
	return s.repository.UpdateUserQualification(ctx, item)
}

func (s *Service) DeleteUserQualification(ctx context.Context, id string) error {
	return s.repository.DeleteUserQualification(ctx, strings.TrimSpace(id))
}

func prepareLookupItem(id, name string, active bool) (domain.LookupItem, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" || name == "" {
		return domain.LookupItem{}, domain.ErrInvalidInput
	}
	return domain.LookupItem{ID: id, Name: name, IsActive: active}, nil
}
