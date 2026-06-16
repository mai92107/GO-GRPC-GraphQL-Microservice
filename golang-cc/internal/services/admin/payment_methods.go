package admin

import (
	"context"
	"regexp"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

var paymentMethodCodePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

func (s *Service) ListPaymentMethods(ctx context.Context) ([]domain.PaymentMethod, error) {
	return s.repository.ListPaymentMethods(ctx)
}

func (s *Service) CreatePaymentMethod(ctx context.Context, code, name string) error {
	code, name = strings.TrimSpace(code), strings.TrimSpace(name)
	if !paymentMethodCodePattern.MatchString(code) || code == recommendations.AnyPaymentCode || name == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.CreatePaymentMethod(ctx, domain.PaymentMethod{Code: code, Name: name, IsActive: true})
}

func (s *Service) UpdatePaymentMethod(ctx context.Context, code, name string, active bool) error {
	name = strings.TrimSpace(name)
	if code == "" || code == recommendations.AnyPaymentCode || name == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdatePaymentMethod(ctx, code, name, active)
}

func (s *Service) DeletePaymentMethod(ctx context.Context, code string) error {
	return s.repository.DeletePaymentMethod(ctx, code)
}
