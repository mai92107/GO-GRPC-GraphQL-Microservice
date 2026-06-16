package recommendations

import (
	"fmt"
	"sort"
	"strings"
)

const GeneralCategory = "general"
const AnyPaymentCode = "any_payment"
const AnyPaymentName = "不限支付方式"
const ExclusiveStackGroupPrefix = "exclusive:"

func Recommend(input RecommendationInput) (Result, error) {
	if input.UserID == "" {
		return Result{}, fmt.Errorf("user_id is required")
	}
	if input.AmountMinor <= 0 {
		return Result{}, fmt.Errorf("amount_minor must be greater than zero")
	}
	if input.CategoryCode == "" || input.CategoryCode == GeneralCategory {
		return Result{}, fmt.Errorf("unsupported category_code %q", input.CategoryCode)
	}
	if len(input.PaymentMethods) == 0 {
		return Result{}, fmt.Errorf("payment_methods are required")
	}
	if input.Date.Time.IsZero() {
		return Result{}, fmt.Errorf("date is required")
	}

	rulesByCard := make(map[ID][]RewardRule)
	for _, rule := range input.Rules {
		if rule.UserID != input.UserID {
			continue
		}
		rulesByCard[rule.CardID] = append(rulesByCard[rule.CardID], rule)
	}

	recommendationsByGroup := make(map[string]CardRecommendation)
	for _, card := range input.Cards {
		if card.UserID != input.UserID || !card.IsActive {
			continue
		}

		methods := make([]PaymentMethod, 0, len(input.PaymentMethods)+1)
		if len(input.PaymentMethods) != 1 {
			methods = append(methods, PaymentMethod{Code: AnyPaymentCode, Name: AnyPaymentName})
		}
		for _, method := range input.PaymentMethods {
			if method.Code != AnyPaymentCode {
				methods = append(methods, method)
			}
		}
		var anyOption PaymentOption
		for index, method := range methods {
			option := evaluateCard(card, rulesByCard[card.ID], input, method)
			if index == 0 {
				anyOption = option
			} else if sameOption(option, anyOption) {
				continue
			}
			if len(option.Allocations) > 0 {
				key := string(card.ID) + ":" + option.TotalScore.String()
				current, exists := recommendationsByGroup[key]
				if !exists {
					recommendationsByGroup[key] = CardRecommendation{
						CardID: card.ID, CardName: card.Name, TotalScore: option.TotalScore, TotalUnweighted: option.TotalUnweighted,
						Allocations: option.Allocations, Explanation: option.Explanation, Reminders: option.Reminders,
						PaymentOptions: []PaymentOption{option},
					}
					continue
				}
				current.PaymentOptions = append(current.PaymentOptions, option)
				if option.TotalUnweighted.Cmp(current.TotalUnweighted) > 0 {
					current.TotalUnweighted = option.TotalUnweighted
				}
				recommendationsByGroup[key] = current
			}
		}
	}

	recommendations := make([]CardRecommendation, 0, len(recommendationsByGroup))
	for _, recommendation := range recommendationsByGroup {
		recommendation.PaymentOptions = mergeEquivalentPaymentOptions(recommendation.PaymentOptions)
		sort.Slice(recommendation.PaymentOptions, func(i, j int) bool {
			return recommendation.PaymentOptions[i].PaymentMethodCode < recommendation.PaymentOptions[j].PaymentMethodCode
		})
		recommendations = append(recommendations, recommendation)
	}
	sort.Slice(recommendations, func(i, j int) bool {
		left, right := recommendations[i], recommendations[j]
		return betterRecommendation(left, right)
	})

	for index := range recommendations {
		recommendations[index].Rank = index + 1
	}

	if len(recommendations) == 0 {
		return Result{
			Recommendations: []CardRecommendation{},
			EmptyReason:     "沒有任何啟用中且符合日期、消費類別與店家的回饋規則",
		}, nil
	}
	return Result{Recommendations: recommendations}, nil
}

func betterRecommendation(left, right CardRecommendation) bool {
	if comparison := left.TotalScore.Cmp(right.TotalScore); comparison != 0 {
		return comparison > 0
	}
	if comparison := left.TotalUnweighted.Cmp(right.TotalUnweighted); comparison != 0 {
		return comparison > 0
	}
	if left.CardName != right.CardName {
		return left.CardName < right.CardName
	}
	return left.CardID < right.CardID
}

func evaluateCard(card Card, rules []RewardRule, input RecommendationInput, method PaymentMethod) PaymentOption {
	result := PaymentOption{
		PaymentMethodCode: method.Code,
		PaymentMethodName: method.Name,
		PaymentMethods:    []PaymentMethod{method},
		Allocations:       []RuleEvaluation{},
		Explanation:       []string{},
		Reminders:         []string{},
	}

	amountTWD := DecimalFromInt(input.AmountMinor).Div(DecimalFromInt(100))
	rules = selectActivity(rules, input.Date)
	candidates := make([]RuleEvaluation, 0, len(rules))
	for _, rule := range rules {
		if !ruleMatches(rule, card.AccountTier, input.CategoryCode, input.MerchantCode, method.Code, input.Date) {
			continue
		}

		uncapped := amountTWD.Mul(rule.Rate).Round(rule.RewardUnit.Precision)
		used := input.MonthlyUsage[rule.ID].Round(rule.RewardUnit.Precision)
		allocated := uncapped
		var remaining *Decimal
		if rule.MonthlyCap != nil {
			value := rule.MonthlyCap.Round(rule.RewardUnit.Precision).Sub(used).Max(Decimal{})
			remaining = &value
			allocated = uncapped.Min(value)
		}

		weight, exists := input.Preferences[rule.RewardUnit.ID]
		if !exists {
			weight = MustDecimal("1")
		}
		twdRate := rule.RewardUnit.TWDRate
		if twdRate.Sign() <= 0 {
			twdRate = MustDecimal("1")
		}
		score := allocated.Mul(twdRate).Mul(weight).Round(6)
		evaluation := RuleEvaluation{
			RuleID:           rule.ID,
			RuleName:         rule.Name,
			ActivityID:       rule.ActivityID,
			ActivityName:     rule.ActivityName,
			StackGroup:       rule.StackGroup,
			ActionRequired:   rule.ActionRequired,
			ActionMessage:    rule.ActionMessage,
			RewardUnit:       rule.RewardUnit,
			Rate:             rule.Rate,
			UncappedReward:   uncapped,
			MonthlyCap:       rule.MonthlyCap,
			UsedBefore:       used,
			RemainingBefore:  remaining,
			AllocatedReward:  allocated,
			PreferenceWeight: weight,
			Score:            score,
		}
		candidates = append(candidates, evaluation)
	}

	selected := bestPerStackGroup(applyExclusiveStackGroups(candidates))
	sort.SliceStable(selected, func(i, j int) bool {
		left, right := ruleByID(rules, selected[i].RuleID), ruleByID(rules, selected[j].RuleID)
		if left.Priority != right.Priority {
			return left.Priority < right.Priority
		}
		return false
	})
	sharedUsed := map[string]Decimal{}
	for _, evaluation := range selected {
		rule := ruleByID(rules, evaluation.RuleID)
		if rule.SharedMonthlyCap != nil {
			key := sharedUsageKey(rule.ActivityID, rule.RewardUnit.ID)
			used := input.ActivityMonthlyUsage[key].Add(sharedUsed[key])
			remaining := rule.SharedMonthlyCap.Round(rule.RewardUnit.Precision).Sub(used).Max(Decimal{})
			evaluation.AllocatedReward = evaluation.AllocatedReward.Min(remaining)
			twdRate := evaluation.RewardUnit.TWDRate
			if twdRate.Sign() <= 0 {
				twdRate = MustDecimal("1")
			}
			evaluation.Score = evaluation.AllocatedReward.Mul(twdRate).Mul(evaluation.PreferenceWeight).Round(6)
			sharedUsed[key] = sharedUsed[key].Add(evaluation.AllocatedReward)
		}
		result.Allocations = append(result.Allocations, evaluation)
		score := evaluation.Score
		allocated := evaluation.AllocatedReward
		twdRate := evaluation.RewardUnit.TWDRate
		if twdRate.Sign() <= 0 {
			twdRate = MustDecimal("1")
		}
		result.TotalScore = result.TotalScore.Add(score)
		result.TotalUnweighted = result.TotalUnweighted.Add(allocated.Mul(twdRate).Round(6))
		result.Explanation = append(result.Explanation,
			fmt.Sprintf("命中%s，本次預估取得 %s", evaluation.RuleName, FormatReward(allocated, evaluation.RewardUnit)))
		if evaluation.ActionRequired != "" && evaluation.ActionRequired != "none" && evaluation.ActionMessage != "" {
			result.Reminders = appendUnique(result.Reminders, evaluation.ActionMessage)
		}
	}
	return result
}

func mergeEquivalentPaymentOptions(options []PaymentOption) []PaymentOption {
	merged := make([]PaymentOption, 0, len(options))
	for _, option := range options {
		found := -1
		for i := range merged {
			if sameOption(option, merged[i]) {
				found = i
				break
			}
		}
		if found < 0 {
			merged = append(merged, option)
			continue
		}
		merged[found].PaymentMethods = append(merged[found].PaymentMethods, option.PaymentMethods...)
		sort.Slice(merged[found].PaymentMethods, func(i, j int) bool {
			return merged[found].PaymentMethods[i].Name < merged[found].PaymentMethods[j].Name
		})
		names := make([]string, 0, len(merged[found].PaymentMethods))
		for _, method := range merged[found].PaymentMethods {
			names = append(names, method.Name)
		}
		merged[found].PaymentMethodName = strings.Join(names, "、")
	}
	return merged
}

func sameOption(left, right PaymentOption) bool {
	if len(left.Allocations) == 0 || len(left.Allocations) != len(right.Allocations) || left.TotalScore.Cmp(right.TotalScore) != 0 {
		return false
	}
	for i := range left.Allocations {
		if left.Allocations[i].RuleID != right.Allocations[i].RuleID ||
			left.Allocations[i].AllocatedReward.Cmp(right.Allocations[i].AllocatedReward) != 0 {
			return false
		}
	}
	return true
}

func FormatDecimal(value Decimal, precision int) string {
	return strings.TrimRight(strings.TrimRight(value.StringFixed(precision), "0"), ".")
}

func FormatReward(value Decimal, unit RewardUnit) string {
	number := FormatDecimal(value, unit.Precision)
	if unit.SymbolPosition == "suffix" {
		return number + unit.Symbol
	}
	return unit.Symbol + number
}

func SumRates(option PaymentOption) Decimal {
	total := Decimal{}
	for _, allocation := range option.Allocations {
		total = total.Add(allocation.Rate)
	}
	return total
}

func ruleMatches(rule RewardRule, accountTier, category, merchantCode, paymentMethod string, date LocalDate) bool {
	if !rule.IsActive || rule.Rate.Sign() <= 0 {
		return false
	}
	if rule.StartDate != nil && date.Before(rule.StartDate.Time) {
		return false
	}
	if rule.EndDate != nil && date.After(rule.EndDate.Time) {
		return false
	}
	if len(rule.RequiredAccountTiers) > 0 {
		matched := false
		for _, tier := range rule.RequiredAccountTiers {
			if tier == accountTier {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(rule.PaymentMethods) > 0 {
		if paymentMethod == AnyPaymentCode {
			return false
		}
		matched := false
		for _, method := range rule.PaymentMethods {
			if method == paymentMethod {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	categoryMatches := false
	for _, candidate := range rule.CategoryCode {
		if candidate == GeneralCategory || candidate == category {
			categoryMatches = true
			break
		}
	}
	if !categoryMatches {
		return false
	}
	if len(rule.MerchantCodes) == 0 {
		return true
	}
	for _, code := range rule.MerchantCodes {
		if code == merchantCode {
			return true
		}
	}
	return false
}

func selectActivity(rules []RewardRule, date LocalDate) []RewardRule {
	var selected ID
	var selectedStart *LocalDate
	for _, rule := range rules {
		if !rule.IsActive || (rule.StartDate != nil && date.Before(rule.StartDate.Time)) ||
			(rule.EndDate != nil && date.After(rule.EndDate.Time)) {
			continue
		}
		if selected == "" || (rule.StartDate != nil && (selectedStart == nil || rule.StartDate.After(selectedStart.Time))) {
			selected, selectedStart = rule.ActivityID, rule.StartDate
		}
	}
	out := make([]RewardRule, 0, len(rules))
	for _, rule := range rules {
		if rule.ActivityID == selected {
			out = append(out, rule)
		}
	}
	return out
}

func bestPerStackGroup(values []RuleEvaluation) []RuleEvaluation {
	best := map[string]RuleEvaluation{}
	order := []string{}
	for _, value := range values {
		group := value.StackGroup
		if group == "" {
			group = string(value.RuleID)
		}
		current, exists := best[group]
		if !exists {
			order = append(order, group)
		}
		if !exists || value.Score.Cmp(current.Score) > 0 ||
			(value.Score.Cmp(current.Score) == 0 && value.RuleID < current.RuleID) {
			best[group] = value
		}
	}
	out := make([]RuleEvaluation, 0, len(best))
	for _, group := range order {
		out = append(out, best[group])
	}
	return out
}

func applyExclusiveStackGroups(values []RuleEvaluation) []RuleEvaluation {
	exclusive := make([]RuleEvaluation, 0)
	for _, value := range values {
		if strings.HasPrefix(value.StackGroup, ExclusiveStackGroupPrefix) {
			exclusive = append(exclusive, value)
		}
	}
	if len(exclusive) > 0 {
		return exclusive
	}
	return values
}

func ruleByID(rules []RewardRule, id ID) RewardRule {
	for _, rule := range rules {
		if rule.ID == id {
			return rule
		}
	}
	return RewardRule{}
}

func sharedUsageKey(activityID, unitID ID) string {
	return string(activityID) + ":" + string(unitID)
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
