package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/services/recommendations"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) ListPaymentMethods(ctx context.Context) ([]domain.PaymentMethod, error) {
	return s.repository.ListPaymentMethods(ctx)
}

func (s *Service) CreatePaymentMethod(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", domain.ErrInvalidInput
	}
	id := secure.UUID()
	return id, s.repository.CreatePaymentMethod(ctx, domain.PaymentMethod{ID: id, Name: name, IsActive: true})
}

func (s *Service) UpdatePaymentMethod(ctx context.Context, id, name string, active bool) error {
	name = strings.TrimSpace(name)
	if id == "" || id == recommendations.AnyPaymentID || name == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdatePaymentMethod(ctx, id, name, active)
}

func (s *Service) DeletePaymentMethod(ctx context.Context, id string) error {
	return s.repository.DeletePaymentMethod(ctx, id)
}
