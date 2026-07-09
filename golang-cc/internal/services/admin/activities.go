package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
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

type activityValidationError struct {
	message string
}

func (e activityValidationError) Error() string {
	return e.message
}

func (e activityValidationError) Unwrap() error {
	return domain.ErrInvalidInput
}

func invalidActivityInput(format string, args ...any) error {
	return activityValidationError{message: fmt.Sprintf(format, args...)}
}

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

func (s *Service) PublishActivity(ctx context.Context, id, userID string) error {
	flow, err := s.repository.GetActivity(ctx, id)
	if err != nil {
		return err
	}
	if err := validatePublishableActivity(flow); err != nil {
		return err
	}
	payload, err := json.Marshal(flow)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(payload)
	return s.repository.PublishActivity(ctx, id, userID, hex.EncodeToString(sum[:]))
}

func (s *Service) SetActivityStatus(ctx context.Context, id string, active bool) error {
	return s.repository.SetActivityStatus(ctx, id, active)
}

func (s *Service) DeleteActivity(ctx context.Context, id string) error {
	return s.repository.DeleteActivity(ctx, id)
}

func validatePublishableActivity(flow domain.ActivityFlow) error {
	if !flow.Activity.IsActive {
		return invalidActivityInput("Activity 必須啟用後才能發布")
	}
	if len(flow.RewardGroups) == 0 {
		return invalidActivityInput("Activity 至少需要 1 個 Group 才能發布")
	}
	publishableRules := 0
	for groupIndex, group := range flow.RewardGroups {
		if !group.IsActive {
			continue
		}
		if len(group.Components) == 0 {
			return invalidActivityInput("Group[%d].components 至少需要 1 個 Component 才能發布", groupIndex+1)
		}
		for componentIndex, component := range group.Components {
			if !component.IsActive {
				continue
			}
			path := fmt.Sprintf("Group[%d].Component[%d]", groupIndex+1, componentIndex+1)
			if component.EffectiveFrom < flow.Activity.EffectiveFrom || component.EffectiveTo > flow.Activity.EffectiveTo {
				return invalidActivityInput("%s 日期必須落在 Activity 日期內", path)
			}
			activeRequirements := 0
			for _, requirement := range component.Requirements {
				if requirement.IsActive {
					activeRequirements++
				}
			}
			if activeRequirements == 0 {
				return invalidActivityInput("%s 至少需要 1 個啟用 Requirement 才能發布", path)
			}
			activeBenefits := 0
			for _, benefit := range component.Benefits {
				if benefit.IsActive {
					activeBenefits++
				}
			}
			if activeBenefits == 0 {
				return invalidActivityInput("%s 至少需要 1 個啟用 Benefit 才能發布", path)
			}
			publishableRules += activeBenefits
		}
	}
	if publishableRules == 0 {
		return invalidActivityInput("Activity 至少需要 1 個可發布的啟用 Benefit")
	}
	return nil
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
		return domain.ActivityFlow{}, err
	}
	if len(input.RewardGroups) == 0 {
		return domain.ActivityFlow{}, invalidActivityInput("reward_groups 至少需要 1 個 Group")
	}
	flow := domain.ActivityFlow{Activity: activity, RewardGroups: make([]domain.RewardGroup, 0, len(input.RewardGroups))}
	groupIDs := map[string]string{}
	for groupIndex, groupInput := range input.RewardGroups {
		group, err := prepareRewardGroup(activity.ID, groupInput, groupIndex)
		if err != nil {
			return domain.ActivityFlow{}, err
		}
		if groupInput.ID != "" {
			groupIDs[groupInput.ID] = group.ID
		}
		groupIDs[group.ID] = group.ID
		flow.RewardGroups = append(flow.RewardGroups, group)
	}
	componentsByID := map[string]domain.RewardComponent{}
	componentOrder := []string{}
	componentIDs := map[string]string{}
	requirementIDs := map[string]string{}
	benefitIDs := map[string]string{}
	for groupIndex, groupInput := range input.RewardGroups {
		groupID := groupIDs[groupInput.ID]
		for componentIndex, componentInput := range groupInput.Components {
			component, err := prepareRewardComponent(componentInput, groupID, groupIDs, componentIDs, requirementIDs, benefitIDs, activity.EffectiveFrom, activity.EffectiveTo, groupIndex, componentIndex)
			if err != nil {
				return domain.ActivityFlow{}, err
			}
			if existing, ok := componentsByID[component.ID]; ok {
				existing.RewardGroupIDs = uniqueStrings(append(existing.RewardGroupIDs, component.RewardGroupIDs...))
				componentsByID[component.ID] = existing
				continue
			}
			componentsByID[component.ID] = component
			componentOrder = append(componentOrder, component.ID)
		}
	}
	for index := range flow.RewardGroups {
		for _, componentID := range componentOrder {
			component := componentsByID[componentID]
			if containsString(component.RewardGroupIDs, flow.RewardGroups[index].ID) {
				flow.RewardGroups[index].Components = append(flow.RewardGroups[index].Components, component)
			}
		}
		if len(flow.RewardGroups[index].Components) == 0 {
			return domain.ActivityFlow{}, invalidActivityInput("Group[%d].components 至少需要 1 個 Component", index+1)
		}
	}
	return flow, nil
}

func prepareActivitySummary(id string, input ActivitySummaryInput) (domain.ActivitySummary, error) {
	start, startErr := parseActivityLocalDate(input.EffectiveFrom)
	end, endErr := parseActivityLocalDate(input.EffectiveTo)
	if id == "" {
		return domain.ActivitySummary{}, invalidActivityInput("Activity.id 必填")
	}
	if input.BankID == "" {
		return domain.ActivitySummary{}, invalidActivityInput("Activity.bank_id 必填")
	}
	if input.CardProductID == "" {
		return domain.ActivitySummary{}, invalidActivityInput("Activity.card_product_id 必填")
	}
	if strings.TrimSpace(input.Title) == "" {
		return domain.ActivitySummary{}, invalidActivityInput("Activity.title 必填")
	}
	if startErr != nil {
		return domain.ActivitySummary{}, invalidActivityInput("Activity.effective_from 必須是 YYYY-MM-DD")
	}
	if endErr != nil {
		return domain.ActivitySummary{}, invalidActivityInput("Activity.effective_to 必須是 YYYY-MM-DD")
	}
	if end.Before(start.Time) {
		return domain.ActivitySummary{}, invalidActivityInput("Activity.effective_to 必須晚於或等於 effective_from")
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

func prepareRewardGroup(activityID string, input RewardGroupInput, index int) (domain.RewardGroup, error) {
	id := stableUUID(input.ID, nil)
	if strings.TrimSpace(input.Name) == "" {
		return domain.RewardGroup{}, invalidActivityInput("Group[%d].name 必填", index+1)
	}
	if len(input.Components) == 0 {
		return domain.RewardGroup{}, invalidActivityInput("Group[%d].components 至少需要 1 個 Component", index+1)
	}
	group := domain.RewardGroup{
		ID: id, ActivityID: activityID, Name: strings.TrimSpace(input.Name), Description: strings.TrimSpace(input.Description),
		DisplayOrder: input.DisplayOrder, IsActive: input.IsActive, Components: make([]domain.RewardComponent, 0, len(input.Components)),
	}
	return group, nil
}

func prepareRewardComponent(input RewardComponentInput, currentGroupID string, groupIDs, componentIDs, requirementIDs, benefitIDs map[string]string, activityFrom, activityTo string, groupIndex, componentIndex int) (domain.RewardComponent, error) {
	id := stableUUID(input.ID, componentIDs)
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
	rewardGroupIDs := resolveRewardGroupIDs(input.RewardGroupIDs, currentGroupID, groupIDs)
	path := fmt.Sprintf("Group[%d].Component[%d]", groupIndex+1, componentIndex+1)
	if strings.TrimSpace(input.Name) == "" {
		return domain.RewardComponent{}, invalidActivityInput("%s.name 必填", path)
	}
	if input.Layer <= 0 {
		return domain.RewardComponent{}, invalidActivityInput("%s.layer 必須大於 0", path)
	}
	if strings.TrimSpace(input.StackGroup) == "" {
		return domain.RewardComponent{}, invalidActivityInput("%s.stack_group 必填", path)
	}
	if stackMode == "" {
		return domain.RewardComponent{}, invalidActivityInput("%s.stack_mode 無效", path)
	}
	if startErr != nil {
		return domain.RewardComponent{}, invalidActivityInput("%s.effective_from 必須是 YYYY-MM-DD", path)
	}
	if endErr != nil {
		return domain.RewardComponent{}, invalidActivityInput("%s.effective_to 必須是 YYYY-MM-DD", path)
	}
	if end.Before(start.Time) {
		return domain.RewardComponent{}, invalidActivityInput("%s.effective_to 必須晚於或等於 effective_from", path)
	}
	if len(input.Requirements) == 0 {
		return domain.RewardComponent{}, invalidActivityInput("%s.requirements 至少需要 1 個 Requirement", path)
	}
	if len(input.Benefits) == 0 {
		return domain.RewardComponent{}, invalidActivityInput("%s.benefits 至少需要 1 個 Benefit", path)
	}
	if len(rewardGroupIDs) == 0 {
		return domain.RewardComponent{}, invalidActivityInput("%s.reward_group_ids 至少需要 1 個有效 Group", path)
	}
	component := domain.RewardComponent{
		ID: id, RewardGroupIDs: rewardGroupIDs, Name: strings.TrimSpace(input.Name), Description: strings.TrimSpace(input.Description),
		Layer: input.Layer, StackGroup: strings.TrimSpace(input.StackGroup), StackMode: stackMode, Priority: input.Priority,
		EffectiveFrom: from, EffectiveTo: to, IsActive: input.IsActive,
		Requirements: make([]domain.RewardRequirement, 0, len(input.Requirements)), Benefits: make([]domain.RewardBenefit, 0, len(input.Benefits)),
	}
	for requirementIndex, requirementInput := range input.Requirements {
		requirement, err := prepareRewardRequirement(id, requirementInput, requirementIDs, path, requirementIndex)
		if err != nil {
			return domain.RewardComponent{}, err
		}
		component.Requirements = append(component.Requirements, requirement)
	}
	for benefitIndex, benefitInput := range input.Benefits {
		benefit, err := prepareRewardBenefit(id, benefitInput, benefitIDs, path, benefitIndex)
		if err != nil {
			return domain.RewardComponent{}, err
		}
		component.Benefits = append(component.Benefits, benefit)
	}
	return component, nil
}

func prepareRewardRequirement(componentID string, input RewardRequirementInput, requirementIDs map[string]string, componentPath string, index int) (domain.RewardRequirement, error) {
	id := stableUUID(input.ID, requirementIDs)
	configuration := json.RawMessage(input.Configuration)
	if len(configuration) == 0 {
		configuration = json.RawMessage(`{}`)
	}
	path := fmt.Sprintf("%s.Requirement[%d]", componentPath, index+1)
	if id == "" {
		return domain.RewardRequirement{}, invalidActivityInput("%s.id 無效", path)
	}
	if strings.TrimSpace(input.RequirementType) == "" {
		return domain.RewardRequirement{}, invalidActivityInput("%s.requirement_type 必填", path)
	}
	if normalizeRequirementOperator(input.Operator) == "" {
		return domain.RewardRequirement{}, invalidActivityInput("%s.operator 無效", path)
	}
	if !json.Valid(configuration) {
		return domain.RewardRequirement{}, invalidActivityInput("%s.configuration_json 必須是合法 JSON", path)
	}
	if !validRequirementConfiguration(input.RequirementType, configuration) {
		return domain.RewardRequirement{}, invalidActivityInput("%s.configuration_json 與 requirement_type 不相符", path)
	}
	return domain.RewardRequirement{
		ID: id, RewardComponentID: componentID, RequirementType: strings.TrimSpace(input.RequirementType),
		Operator: normalizeRequirementOperator(input.Operator), Configuration: configuration, Description: strings.TrimSpace(input.Description), IsActive: input.IsActive,
	}, nil
}

func prepareRewardBenefit(componentID string, input RewardBenefitInput, benefitIDs map[string]string, componentPath string, index int) (domain.RewardBenefit, error) {
	id := stableUUID(input.ID, benefitIDs)
	value := strings.TrimSpace(input.Value)
	parsedValue, valueErr := recommendations.ParseDecimal(value)
	path := fmt.Sprintf("%s.Benefit[%d]", componentPath, index+1)
	if id == "" {
		return domain.RewardBenefit{}, invalidActivityInput("%s.id 無效", path)
	}
	if strings.TrimSpace(input.BenefitType) == "" {
		return domain.RewardBenefit{}, invalidActivityInput("%s.benefit_type 必填", path)
	}
	if input.RewardUnitID == "" {
		return domain.RewardBenefit{}, invalidActivityInput("%s.reward_unit_id 必填", path)
	}
	if valueErr != nil {
		return domain.RewardBenefit{}, invalidActivityInput("%s.value 必須是數字", path)
	}
	if parsedValue.Sign() <= 0 {
		return domain.RewardBenefit{}, invalidActivityInput("%s.value 必須大於 0", path)
	}
	if input.CapAmount != nil {
		cap := strings.TrimSpace(*input.CapAmount)
		if cap == "" {
			input.CapAmount = nil
		} else {
			parsedCap, capErr := recommendations.ParseDecimal(cap)
			if capErr != nil || parsedCap.Sign() <= 0 {
				return domain.RewardBenefit{}, invalidActivityInput("%s.cap_amount 必須是大於 0 的數字", path)
			}
			input.CapAmount = &cap
		}
	}
	if input.CapFormula != nil {
		formula := strings.TrimSpace(*input.CapFormula)
		if formula == "" {
			input.CapFormula = nil
		} else if err := recommendations.ValidateCapFormula(formula); err != nil {
			return domain.RewardBenefit{}, invalidActivityInput("%s.cap_formula 無效", path)
		} else {
			input.CapFormula = &formula
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
	if input.CapPeriod != nil && input.CapAmount == nil && input.CapFormula == nil {
		return domain.RewardBenefit{}, invalidActivityInput("%s.cap_period 需搭配 cap_amount 或 cap_formula", path)
	}
	return domain.RewardBenefit{
		ID: id, RewardComponentID: componentID, BenefitType: strings.TrimSpace(input.BenefitType), Value: value,
		RewardUnitID: input.RewardUnitID, CapAmount: input.CapAmount, CapFormula: input.CapFormula, CapPeriod: input.CapPeriod,
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

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func isUUID(value string) bool {
	return uuidPattern.MatchString(strings.TrimSpace(value))
}

func stableUUID(raw string, seen map[string]string) string {
	raw = strings.TrimSpace(raw)
	if isUUID(raw) {
		return raw
	}
	if seen != nil && raw != "" {
		if id := seen[raw]; id != "" {
			return id
		}
	}
	id := secure.UUID()
	if seen != nil && raw != "" {
		seen[raw] = id
	}
	return id
}

func resolveRewardGroupIDs(rawIDs []string, currentGroupID string, groupIDs map[string]string) []string {
	out := []string{}
	for _, rawID := range rawIDs {
		rawID = strings.TrimSpace(rawID)
		id := groupIDs[rawID]
		if id != "" {
			out = append(out, id)
		}
	}
	return uniqueStrings(out)
}

func uniqueStrings(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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
	case "INSTALLMENT":
		return "is_installment"
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
