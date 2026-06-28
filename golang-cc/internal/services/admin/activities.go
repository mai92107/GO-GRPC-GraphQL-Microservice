package admin

import (
	"context"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/services/recommendations"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) ListActivities(ctx context.Context) ([]domain.Activity, error) {
	return s.repository.ListActivities(ctx)
}

func (s *Service) CreateActivity(ctx context.Context, input ActivityInput) (string, error) {
	item, err := prepareActivity(secure.UUID(), input, true)
	if err != nil {
		return "", err
	}
	return item.ID, s.repository.CreateActivity(ctx, item)
}

func (s *Service) UpdateActivity(ctx context.Context, id string, input ActivityInput) error {
	item, err := prepareActivity(id, input, false)
	if err != nil {
		return err
	}
	return s.repository.UpdateActivity(ctx, item)
}

func (s *Service) DeleteActivity(ctx context.Context, id string) error {
	return s.repository.DeleteActivity(ctx, id)
}

func prepareActivity(id string, input ActivityInput, defaultActive bool) (domain.Activity, error) {
	start, startErr := recommendations.ParseLocalDate(input.StartDate)
	end, endErr := recommendations.ParseLocalDate(input.EndDate)
	input.NetworkIDs = normalizeIDs(input.NetworkIDs)
	if startErr != nil || endErr != nil || end.Before(start.Time) || id == "" || input.CardProductID == "" ||
		strings.TrimSpace(input.Name) == "" || len(input.NetworkIDs) == 0 || len(input.Benefits) == 0 {
		return domain.Activity{}, domain.ErrInvalidInput
	}
	active := defaultActive
	if input.IsActive != nil {
		active = *input.IsActive
	}
	benefits := make([]domain.ActivityBenefit, 0, len(input.Benefits))
	if input.SharedMonthlyCaps == nil {
		input.SharedMonthlyCaps = map[string]string{}
	}
	for _, value := range input.Benefits {
		rewardValue := strings.TrimSpace(value.RewardValue)
		parsedRewardValue, rewardErr := recommendations.ParseDecimal(rewardValue)
		effectType := normalizeEffectType(value.EffectType)
		if rewardErr != nil || parsedRewardValue.Sign() <= 0 || effectType == "" ||
			value.RewardUnitID == "" || strings.TrimSpace(value.Name) == "" ||
			strings.TrimSpace(value.StackGroup) == "" || len(value.CategoryIDs) == 0 {
			return domain.Activity{}, domain.ErrInvalidInput
		}
		value.Layer = strings.TrimSpace(value.Layer)
		if value.Layer == "" {
			return domain.Activity{}, domain.ErrInvalidInput
		}
		if value.ActionRequired == "" {
			value.ActionRequired = "none"
		}
		if value.ActionRequired != "none" && value.ActionRequired != "registration" && value.ActionRequired != "app_switch" && value.ActionRequired != "account_setup" {
			return domain.Activity{}, domain.ErrInvalidInput
		}
		value.SelectableType = strings.TrimSpace(value.SelectableType)
		if value.ActionRequired == "app_switch" && (strings.TrimSpace(value.ActionMessage) == "" || value.SelectableType == "" || value.QualifiedType == "") {
			return domain.Activity{}, domain.ErrInvalidInput
		}
		if value.ActionRequired != "app_switch" && value.SelectableType != "" {
			return domain.Activity{}, domain.ErrInvalidInput
		}
		stackGroup := strings.TrimSpace(value.StackGroup)
		stackPolicy := normalizeStackPolicy(value.StackPolicy, stackGroup)

		if stackPolicy == "" {
			return domain.Activity{}, domain.ErrInvalidInput
		}
		seen := map[string]bool{}

		merchants := make([]string, 0, len(value.MerchantIds))
		for _, code := range value.MerchantIds {
			code = strings.TrimSpace(code)
			normalized := "merchant:" + code
			if code == "" || seen[normalized] {
				return domain.Activity{}, domain.ErrInvalidInput
			}
			seen[normalized] = true
			merchants = append(merchants, code)
		}
		paymentMethods := make([]string, 0, len(value.PaymentMethods))
		for _, method := range value.PaymentMethods {
			method = strings.TrimSpace(method)
			if method == "" || method == recommendations.AnyPaymentID || seen["payment:"+method] {
				return domain.Activity{}, domain.ErrInvalidInput
			}
			seen["payment:"+method] = true
			paymentMethods = append(paymentMethods, method)
		}
		benefitID := value.ID
		if benefitID == "" {
			benefitID = secure.UUID()
		}
		benefits = append(benefits, domain.ActivityBenefit{
			ID: benefitID, RewardUnitID: value.RewardUnitID, Name: strings.TrimSpace(value.Name),
			Layer: value.Layer, DisplayOrder: value.DisplayOrder, EffectType: effectType, RewardValue: rewardValue,
			MonthlyCap: value.MonthlyCap, StackGroup: stackGroup, Priority: value.Priority,
			QualifiedType: value.QualifiedType, SelectableType: value.SelectableType, ActionRequired: value.ActionRequired,
			ActionMessage: strings.TrimSpace(value.ActionMessage), PaymentMethods: paymentMethods,
			CategoryIDs: value.CategoryIDs, MerchantIDs: merchants,
		})
	}
	return domain.Activity{
		ID: id, CardProductID: input.CardProductID, Name: strings.TrimSpace(input.Name),
		StartDate: input.StartDate, EndDate: input.EndDate, IsActive: active, SourceURL: strings.TrimSpace(input.SourceURL),
		VerifiedAt: input.VerifiedAt, NetworkIDs: input.NetworkIDs, SharedMonthlyCaps: input.SharedMonthlyCaps, Benefits: benefits,
	}, nil
}

func normalizeEffectType(value string) string {
	switch strings.TrimSpace(value) {
	case "", string(recommendations.EffectAddRate):
		return string(recommendations.EffectAddRate)
	case string(recommendations.EffectSetRate), string(recommendations.EffectMultiplyRate), string(recommendations.EffectAddCash), string(recommendations.EffectDiscount):
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeStackPolicy(value, stackGroup string) string {
	switch strings.TrimSpace(value) {
	case recommendations.StackPolicyStack, recommendations.StackPolicyBestOfGroup, recommendations.StackPolicyExclusive:
		return strings.TrimSpace(value)
	case "":
		if strings.HasPrefix(stackGroup, recommendations.ExclusiveStackGroupPrefix) {
			return recommendations.StackPolicyExclusive
		}
		if stackGroup != "" {
			return recommendations.StackPolicyBestOfGroup
		}
		return recommendations.StackPolicyStack
	default:
		return ""
	}
}
