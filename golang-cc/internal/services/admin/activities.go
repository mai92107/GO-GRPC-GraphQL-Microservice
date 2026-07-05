package admin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/services/recommendations"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

const (
	StackModeAdditive  = "ADDITIVE"
	StackModeBestOnly  = "BEST_ONLY"
	StackModeExclusive = "EXCLUSIVE"
)

func (s *Service) ListActivities(ctx context.Context, filters ActivityFilters) ([]domain.ActivitySummary, error) {
	return s.repository.ListActivities(ctx, filters)
}

func (s *Service) GetActivity(ctx context.Context, id string) (domain.ActivityFlow, error) {
	return s.repository.GetActivity(ctx, id)
}

func (s *Service) CreateActivity(ctx context.Context, input ActivityFlowInput) (string, error) {
	flow, err := prepareActivityFlow(secure.UUID(), input)
	if err != nil {
		return "", err
	}
	return flow.Activity.ID, s.repository.CreateActivity(ctx, flow)
}

func (s *Service) UpdateActivity(ctx context.Context, id string, input ActivityFlowInput) error {
	flow, err := prepareActivityFlow(id, input)
	if err != nil {
		return err
	}
	return s.repository.UpdateActivity(ctx, flow)
}

func (s *Service) SetActivityStatus(ctx context.Context, id string, active bool) error {
	return s.repository.SetActivityStatus(ctx, id, active)
}

func (s *Service) DeleteActivity(ctx context.Context, id string) error {
	return s.repository.DeleteActivity(ctx, id)
}

func (s *Service) ListRequirementTypes(ctx context.Context) ([]domain.RequirementTypeOption, error) {
	return s.repository.ListRequirementTypes(ctx)
}

func (s *Service) ActivityRequirementOptions(ctx context.Context, filters RequirementOptionFilters) (domain.RequirementOptionSet, error) {
	filters.RequirementType = strings.TrimSpace(filters.RequirementType)
	filters.CardProductID = strings.TrimSpace(filters.CardProductID)
	return s.repository.ActivityRequirementOptions(ctx, filters)
}

func prepareActivityFlow(id string, input ActivityFlowInput) (domain.ActivityFlow, error) {
	activity, err := prepareActivitySummary(id, input.Activity)
	if err != nil {
		println("prepareActivityFlow error:", err.Error())
		return domain.ActivityFlow{}, err
	}
	if len(input.RewardGroups) == 0 {
		println("prepareActivityFlow error: no reward groups provided")
		return domain.ActivityFlow{}, domain.ErrInvalidInput
	}
	flow := domain.ActivityFlow{Activity: activity, RewardGroups: make([]domain.RewardGroup, 0, len(input.RewardGroups))}
	for _, groupInput := range input.RewardGroups {
		group, err := prepareRewardGroup(activity.ID, groupInput, activity.EffectiveFrom, activity.EffectiveTo)
		if err != nil {
			return domain.ActivityFlow{}, err
		}
		flow.RewardGroups = append(flow.RewardGroups, group)
	}
	return flow, nil
}

func prepareActivitySummary(id string, input ActivitySummaryInput) (domain.ActivitySummary, error) {
	start, startErr := parseActivityLocalDate(input.EffectiveFrom)
	end, endErr := parseActivityLocalDate(input.EffectiveTo)
	if id == "" || input.BankID == "" || input.CardProductID == "" || strings.TrimSpace(input.Title) == "" ||
		startErr != nil || endErr != nil || end.Before(start.Time) {
		println("prepareActivitySummary error: invalid input - id:", id, "bankID:", input.BankID, "cardProductID:", input.CardProductID, "title:", input.Title, "startErr:", startErr, "endErr:", endErr)
		return domain.ActivitySummary{}, domain.ErrInvalidInput
	}
	return domain.ActivitySummary{
		ID: id, BankID: input.BankID, CardProductID: input.CardProductID,
		Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description),
		SourceURL: strings.TrimSpace(input.SourceURL), EffectiveFrom: input.EffectiveFrom,
		EffectiveTo: input.EffectiveTo, IsActive: input.IsActive,
	}, nil
}

func parseActivityLocalDate(value string) (recommendations.LocalDate, error) {
	if !isYYYYMMDD(value) {
		return recommendations.LocalDate{}, domain.ErrInvalidInput
	}
	return recommendations.ParseLocalDate(value)
}

func isYYYYMMDD(value string) bool {
	if len(value) != len("2006-01-02") || value[4] != '-' || value[7] != '-' {
		return false
	}
	for index, char := range value {
		if index == 4 || index == 7 {
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func prepareRewardGroup(activityID string, input RewardGroupInput, activityFrom, activityTo string) (domain.RewardGroup, error) {
	id := input.ID
	if id == "" {
		id = secure.UUID()
	}
	if strings.TrimSpace(input.Name) == "" || len(input.Components) == 0 {
		println("prepareRewardGroup error: invalid input - name:", input.Name, "components count:", len(input.Components))
		return domain.RewardGroup{}, domain.ErrInvalidInput
	}
	group := domain.RewardGroup{
		ID: id, ActivityID: activityID, Name: strings.TrimSpace(input.Name), Description: strings.TrimSpace(input.Description),
		DisplayOrder: input.DisplayOrder, IsActive: input.IsActive, Components: make([]domain.RewardComponent, 0, len(input.Components)),
	}
	for _, componentInput := range input.Components {
		component, err := prepareRewardComponent(id, componentInput, activityFrom, activityTo)
		if err != nil {
			println("prepareRewardGroup error:", err.Error())
			return domain.RewardGroup{}, err
		}
		group.Components = append(group.Components, component)
	}
	return group, nil
}

func prepareRewardComponent(groupID string, input RewardComponentInput, activityFrom, activityTo string) (domain.RewardComponent, error) {
	id := input.ID
	if id == "" {
		id = secure.UUID()
	}
	from := input.EffectiveFrom
	if from == "" {
		from = activityFrom
	}
	to := input.EffectiveTo
	if to == "" {
		to = activityTo
	}
	start, startErr := parseActivityLocalDate(from)
	end, endErr := parseActivityLocalDate(to)
	stackMode := normalizeStackMode(input.StackMode)
	if strings.TrimSpace(input.Name) == "" || input.Layer <= 0 || strings.TrimSpace(input.StackGroup) == "" || stackMode == "" ||
		startErr != nil || endErr != nil || end.Before(start.Time) || len(input.Requirements) == 0 || len(input.Benefits) == 0 {
		println("prepareRewardComponent error: invalid input - name:", input.Name, "layer:", input.Layer, "stackGroup:", input.StackGroup, "stackMode:", stackMode, "startErr:", startErr, "endErr:", endErr)
		return domain.RewardComponent{}, domain.ErrInvalidInput
	}
	component := domain.RewardComponent{
		ID: id, RewardGroupID: groupID, Name: strings.TrimSpace(input.Name), Description: strings.TrimSpace(input.Description),
		Layer: input.Layer, StackGroup: strings.TrimSpace(input.StackGroup), StackMode: stackMode, Priority: input.Priority,
		EffectiveFrom: from, EffectiveTo: to, IsActive: input.IsActive,
		Requirements: make([]domain.RewardRequirement, 0, len(input.Requirements)), Benefits: make([]domain.RewardBenefit, 0, len(input.Benefits)),
	}
	for _, requirementInput := range input.Requirements {
		requirement, err := prepareRewardRequirement(id, requirementInput)
		if err != nil {
			println("prepareRewardComponent error:", err.Error())
			return domain.RewardComponent{}, err
		}
		component.Requirements = append(component.Requirements, requirement)
	}
	for _, benefitInput := range input.Benefits {
		benefit, err := prepareRewardBenefit(id, benefitInput)
		if err != nil {
			println("prepareRewardComponent error:", err.Error())
			return domain.RewardComponent{}, err
		}
		component.Benefits = append(component.Benefits, benefit)
	}
	return component, nil
}

func prepareRewardRequirement(componentID string, input RewardRequirementInput) (domain.RewardRequirement, error) {
	id := input.ID
	if id == "" {
		id = secure.UUID()
	}
	configuration := json.RawMessage(input.Configuration)
	if len(configuration) == 0 {
		configuration = json.RawMessage(`{}`)
	}
	if id == "" || strings.TrimSpace(input.RequirementType) == "" || normalizeRequirementOperator(input.Operator) == "" ||
		!json.Valid(configuration) || !validRequirementConfiguration(input.RequirementType, configuration) {
		bytes, _ := configuration.MarshalJSON()
		println("prepareRewardRequirement error: invalid input - id:", id, "requirementType:", input.RequirementType, "operator:", input.Operator, "configuration:", string(bytes))
		println("id == \"\""+id == "")
		println("strings.TrimSpace(input.RequirementType) == \"\""+strings.TrimSpace(input.RequirementType) == "")
		println("normalizeRequirementOperator(input.Operator) == \"\""+normalizeRequirementOperator(input.Operator) == "")
		println("!json.Valid(configuration)", !json.Valid(configuration))
		println("!validRequirementConfiguration(input.RequirementType, configuration))", !validRequirementConfiguration(input.RequirementType, configuration))
		return domain.RewardRequirement{}, domain.ErrInvalidInput
	}
	return domain.RewardRequirement{
		ID: id, RewardComponentID: componentID, RequirementType: strings.TrimSpace(input.RequirementType),
		Operator: normalizeRequirementOperator(input.Operator), Configuration: configuration, Description: strings.TrimSpace(input.Description), IsActive: input.IsActive,
	}, nil
}

func prepareRewardBenefit(componentID string, input RewardBenefitInput) (domain.RewardBenefit, error) {
	id := input.ID
	if id == "" {
		id = secure.UUID()
	}
	value := strings.TrimSpace(input.Value)
	parsedValue, valueErr := recommendations.ParseDecimal(value)
	if id == "" || strings.TrimSpace(input.BenefitType) == "" || input.RewardUnitID == "" || valueErr != nil || parsedValue.Sign() <= 0 {
		println("prepareRewardBenefit error: invalid input - id:", id, "benefitType:", input.BenefitType, "rewardUnitID:", input.RewardUnitID, "value:", value, "valueErr:", valueErr.Error(), "parsedValue:", parsedValue.String())
		return domain.RewardBenefit{}, domain.ErrInvalidInput
	}
	if input.CapAmount != nil {
		cap := strings.TrimSpace(*input.CapAmount)
		if cap == "" {
			input.CapAmount = nil
		} else {
			parsedCap, capErr := recommendations.ParseDecimal(cap)
			if capErr != nil || parsedCap.Sign() <= 0 {
				println("prepareRewardBenefit error: invalid input - id:", id, "capAmount:", cap, "capErr:", capErr.Error(), "parsedCap:", parsedCap.String())
				return domain.RewardBenefit{}, domain.ErrInvalidInput
			}
			input.CapAmount = &cap
		}
	}
	if input.CapPeriod != nil {
		period := strings.TrimSpace(*input.CapPeriod)
		if period == "" {
			input.CapPeriod = nil
		} else {
			input.CapPeriod = &period
		}
	}
	return domain.RewardBenefit{
		ID: id, RewardComponentID: componentID, BenefitType: strings.TrimSpace(input.BenefitType), Value: value,
		RewardUnitID: input.RewardUnitID, CapAmount: input.CapAmount, CapPeriod: input.CapPeriod,
		Description: strings.TrimSpace(input.Description), IsActive: input.IsActive,
	}, nil
}

func normalizeStackMode(value string) string {
	switch strings.TrimSpace(value) {
	case StackModeAdditive, "":
		return StackModeAdditive
	case StackModeBestOnly:
		return StackModeBestOnly
	case StackModeExclusive:
		return StackModeExclusive
	default:
		return ""
	}
}

func normalizeRequirementOperator(value string) string {
	switch strings.TrimSpace(value) {
	case "IN", "NOT_IN", "EQ", "GTE", "LTE", "BETWEEN":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func validRequirementConfiguration(requirementType string, raw json.RawMessage) bool {
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		println("validRequirementConfiguration error: failed to unmarshal configuration, error:", err.Error())
		return false
	}
	key := requiredRequirementKey(requirementType)
	if key == "" {
		println("validRequirementConfiguration error: unknown requirement type:", requirementType)
		return false
	}
	_, ok := config[key]
	if !ok {
		println("validRequirementConfiguration error: missing required key in configuration for requirement type:", requirementType, "required key:", key)
	}
	return ok
}

func requiredRequirementKey(requirementType string) string {
	switch strings.TrimSpace(requirementType) {
	case "PAYMENT_METHOD":
		return "payment_method_codes"
	case "CARD_NETWORK":
		return "network_codes"
	case "CARD_PLAN":
		return "card_plan_ids"
	case "CARD_PRODUCT":
		return "card_product_ids"
	case "MERCHANT":
		return "merchant_ids"
	case "MERCHANT_CATEGORY", "CONSUMPTION_CATEGORY":
		return "category_ids"
	case "CHANNEL":
		return "channels"
	case "REGION":
		return "regions"
	case "CURRENCY":
		return "currency_codes"
	case "AMOUNT":
		return "amount"
	case "ACCOUNT_TIER":
		return "tiers"
	case "USER_QUALIFICATION":
		return "qualification_codes"
	case "DATE_RANGE":
		return "date_range"
	case "WEEKDAY":
		return "weekdays"
	case "TIME_RANGE":
		return "time_range"
	case "ACTION_REQUIRED":
		return "action_codes"
	default:
		return ""
	}
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
