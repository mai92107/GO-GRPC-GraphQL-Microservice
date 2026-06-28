package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/services/recommendations"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) ListRewardUnits(ctx context.Context) ([]domain.RewardUnit, error) {
	return s.repository.ListRewardUnits(ctx)
}

func (s *Service) CreateRewardUnit(ctx context.Context, input RewardUnitInput) (string, error) {
	if !validRewardUnit(input) {
		return "", domain.ErrInvalidInput
	}
	id := secure.UUID()
	return id, s.repository.CreateRewardUnit(ctx, id, domain.RewardUnitInput(input))
}
func (s *Service) UpdateRewardUnit(ctx context.Context, id string, input RewardUnitInput) error {
	if id == "" || !validRewardUnit(input) {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdateRewardUnit(ctx, id, domain.RewardUnitInput(input))
}
func (s *Service) DeleteRewardUnit(ctx context.Context, id string) error {
	return s.repository.DeleteRewardUnit(ctx, id)
}
func validRewardUnit(input RewardUnitInput) bool {
	rate, err := recommendations.ParseDecimal(input.TWDRate)
	return strings.TrimSpace(input.Name) != "" &&
		strings.TrimSpace(input.Symbol) != "" && (input.SymbolPosition == "prefix" || input.SymbolPosition == "suffix") &&
		err == nil && rate.Sign() > 0 && input.Precision >= 0 && input.Precision <= 6
}
