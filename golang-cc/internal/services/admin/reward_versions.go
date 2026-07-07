package admin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

func (s *Service) PublishComponentVersion(ctx context.Context, componentID string, input domain.RewardComponentVersionInput) (string, error) {
	rewardValue, err := recommendations.ParseDecimal(input.RewardValue)
	if err != nil || rewardValue.Sign() <= 0 || componentID == "" || input.RewardUnitID == "" ||
		strings.TrimSpace(input.Name) == "" || input.EffectiveFrom.IsZero() ||
		(input.EffectiveTo != nil && !input.EffectiveTo.After(input.EffectiveFrom)) {
		return "", domain.ErrInvalidInput
	}
	if normalizeEffectType(input.EffectType) == "" {
		return "", domain.ErrInvalidInput
	}
	input.EffectType = normalizeEffectType(input.EffectType)
	return s.repository.PublishComponentVersion(ctx, componentID, input)
}

func (s *Service) PublishConditionVersion(ctx context.Context, conditionID string, input domain.RewardConditionVersionInput) (string, error) {
	if conditionID == "" || !json.Valid(input.Configuration) || input.EffectiveFrom.IsZero() ||
		(input.EffectiveTo != nil && !input.EffectiveTo.After(input.EffectiveFrom)) {
		return "", domain.ErrInvalidInput
	}
	switch input.Operator {
	case "equals", "in", "not_in", "gte", "lte", "between":
	default:
		return "", domain.ErrInvalidInput
	}
	return s.repository.PublishConditionVersion(ctx, conditionID, input)
}

func (s *Service) PublishCapVersion(ctx context.Context, capID string, input domain.RewardCapVersionInput) (string, error) {
	input.LimitValue = strings.TrimSpace(input.LimitValue)
	input.LimitFormula = strings.TrimSpace(input.LimitFormula)
	if input.LimitValue == "" && input.LimitFormula == "" {
		return "", domain.ErrInvalidInput
	}
	if input.LimitValue != "" {
		limit, err := recommendations.ParseDecimal(input.LimitValue)
		if err != nil || limit.Sign() <= 0 {
			return "", domain.ErrInvalidInput
		}
	}
	if input.LimitFormula != "" {
		if err := recommendations.ValidateCapFormula(input.LimitFormula); err != nil {
			return "", domain.ErrInvalidInput
		}
	}
	if capID == "" || input.EffectiveFrom.IsZero() ||
		(input.EffectiveTo != nil && !input.EffectiveTo.After(input.EffectiveFrom)) {
		return "", domain.ErrInvalidInput
	}
	if input.CapType != "reward_amount" && input.CapType != "spending_amount" && input.CapType != "transaction_count" {
		return "", domain.ErrInvalidInput
	}
	if input.PeriodType != "calendar_month" && input.PeriodType != "statement_cycle" && input.PeriodType != "campaign" && input.PeriodType != "year" {
		return "", domain.ErrInvalidInput
	}
	return s.repository.PublishCapVersion(ctx, capID, input)
}
