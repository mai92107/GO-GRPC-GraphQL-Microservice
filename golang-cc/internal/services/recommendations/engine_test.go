package recommendations

import "testing"

var (
	cash   = RewardUnit{ID: "cash", Code: "cash_twd", Name: "現金", Symbol: "NT$", Precision: 2}
	points = RewardUnit{ID: "points", Code: "points", Name: "點數", Symbol: "點", Precision: 2}
)

func TestAcceptanceCaseStacksRulesAndAppliesCap(t *testing.T) {
	input := baseInput()
	input.AmountMinor = 20000
	input.Preferences["points"] = MustDecimal("2")
	input.MonthlyUsage["dining"] = MustDecimal("95")
	input.Rules = []RewardRule{
		rule("general", "card-a", cash, "0.01", nil, "general"),
		rule("dining", "card-a", points, "0.05", decimalPointer("100"), "dining"),
	}

	result := recommendOK(t, input)
	card := result.Recommendations[0]
	assertDecimal(t, card.TotalScore, "12.000000")
	assertDecimal(t, card.Allocations[0].AllocatedReward, "2.00")
	assertDecimal(t, card.Allocations[1].UncappedReward, "10.00")
	assertDecimal(t, card.Allocations[1].AllocatedReward, "5.00")
	assertDecimal(t, *card.Allocations[1].RemainingBefore, "5.00")
}

func TestDateBoundariesAndInactiveEntries(t *testing.T) {
	input := baseInput()
	start, end := MustLocalDate("2026-06-10"), MustLocalDate("2026-06-10")
	active := rule("active", "card-a", cash, "0.01", nil, "general")
	active.StartDate, active.EndDate = &start, &end
	inactive := rule("inactive", "card-a", cash, "10", nil, "general")
	inactive.IsActive = false
	input.Rules = []RewardRule{active, inactive}

	if got := len(recommendOK(t, input).Recommendations[0].Allocations); got != 1 {
		t.Fatalf("allocations = %d, want 1", got)
	}

	input.Date = MustLocalDate("2026-06-11")
	if result := recommendOK(t, input); result.EmptyReason == "" {
		t.Fatal("expected no matching rules after end date")
	}

	input.Date = MustLocalDate("2026-06-09")
	if result := recommendOK(t, input); result.EmptyReason == "" {
		t.Fatal("expected no matching rules before start date")
	}

	input.Date = MustLocalDate("2026-06-10")
	input.Cards[0].IsActive = false
	if result := recommendOK(t, input); result.EmptyReason == "" {
		t.Fatal("expected no recommendation for inactive card")
	}
}

func TestCapStates(t *testing.T) {
	tests := []struct {
		name     string
		cap      *Decimal
		used     string
		expected string
	}{
		{name: "uncapped", cap: nil, used: "0", expected: "10.00"},
		{name: "below cap", cap: decimalPointer("20"), used: "2", expected: "10.00"},
		{name: "partially capped", cap: decimalPointer("10"), used: "4", expected: "6.00"},
		{name: "cap exhausted", cap: decimalPointer("10"), used: "10", expected: "0.00"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := baseInput()
			input.AmountMinor = 10000
			input.Rules = []RewardRule{rule("rule-a", "card-a", cash, "0.1", test.cap, "general")}
			input.MonthlyUsage["rule-a"] = MustDecimal(test.used)
			allocation := recommendOK(t, input).Recommendations[0].Allocations[0]
			assertDecimal(t, allocation.AllocatedReward, test.expected)
		})
	}
}

func TestDifferentMonthlyUsageProducesNewMonthReset(t *testing.T) {
	input := baseInput()
	input.Rules = []RewardRule{rule("limited", "card-a", cash, "0.1", decimalPointer("10"), "general")}
	input.MonthlyUsage["limited"] = MustDecimal("10")
	assertDecimal(t, recommendOK(t, input).Recommendations[0].Allocations[0].AllocatedReward, "0")

	input.MonthlyUsage = map[ID]Decimal{}
	assertDecimal(t, recommendOK(t, input).Recommendations[0].Allocations[0].AllocatedReward, "10")
}

func TestStableTieBreakers(t *testing.T) {
	input := baseInput()
	input.Cards = []Card{
		{ID: "card-z", UserID: "user-a", Name: "同名卡", IsActive: true},
		{ID: "card-b", UserID: "user-a", Name: "乙卡", IsActive: true},
		{ID: "card-a", UserID: "user-a", Name: "同名卡", IsActive: true},
		{ID: "card-high-unweighted", UserID: "user-a", Name: "甲卡", IsActive: true},
	}
	input.Preferences["points"] = MustDecimal("0.5")
	input.Rules = []RewardRule{
		rule("z", "card-z", cash, "0.01", nil, "general"),
		rule("b", "card-b", cash, "0.01", nil, "general"),
		rule("a", "card-a", cash, "0.01", nil, "general"),
		rule("unweighted", "card-high-unweighted", points, "0.02", nil, "general"),
	}

	got := recommendOK(t, input).Recommendations
	want := []ID{"card-high-unweighted", "card-b", "card-a", "card-z"}
	for index := range want {
		if got[index].CardID != want[index] {
			t.Fatalf("rank %d = %s, want %s", index+1, got[index].CardID, want[index])
		}
	}
}

func TestDuplicateCardIDKeepsBestRecommendation(t *testing.T) {
	input := baseInput()
	low := input.Cards[0]
	low.Name = "低回饋顯示名稱"
	high := low
	high.Name = "高回饋顯示名稱"
	input.Cards = []Card{low, high}
	input.Rules = []RewardRule{rule("reward", "card-a", cash, "0.01", nil, "general")}

	got := recommendOK(t, input).Recommendations
	if len(got) != 1 {
		t.Fatalf("recommendations=%d,want 1", len(got))
	}
	if got[0].CardName != "低回饋顯示名稱" {
		t.Fatalf("card name=%q,want stable best name", got[0].CardName)
	}
}

func TestRoundingAtMinimumAmount(t *testing.T) {
	input := baseInput()
	input.AmountMinor = 1
	input.Rules = []RewardRule{rule("tiny", "card-a", cash, "0.5", nil, "general")}
	allocation := recommendOK(t, input).Recommendations[0].Allocations[0]
	assertDecimal(t, allocation.UncappedReward, "0.01")
}

func TestValidationAndNoEffectiveRules(t *testing.T) {
	input := baseInput()
	input.AmountMinor = 0
	if _, err := Recommend(input); err == nil {
		t.Fatal("expected amount validation error")
	}
	input.AmountMinor = 100
	input.CategoryCode = "general"
	if _, err := Recommend(input); err == nil {
		t.Fatal("expected category validation error")
	}
	input.CategoryCode = "dining"
	input.Rules = nil
	result := recommendOK(t, input)
	if len(result.Recommendations) != 0 || result.EmptyReason == "" {
		t.Fatal("expected an explained empty result")
	}
}

func TestActivityRequiresMatchingCategoryAndMerchantKeyword(t *testing.T) {
	input := baseInput()
	activity := rule("pxmart", "card-a", cash, "0.1", nil, "dining")
	activity.MerchantCodes = []string{"px_mart"}
	input.Rules = []RewardRule{activity}

	input.MerchantCode = "px_mart"
	if got := len(recommendOK(t, input).Recommendations); got != 1 {
		t.Fatalf("matching merchant recommendations = %d, want 1", got)
	}
	input.MerchantCode = "px_mart"
	if got := len(recommendOK(t, input).Recommendations); got != 1 {
		t.Fatalf("case-insensitive merchant recommendations = %d, want 1", got)
	}
	input.MerchantCode = "carrefour"
	if got := len(recommendOK(t, input).Recommendations); got != 0 {
		t.Fatalf("non-matching merchant recommendations = %d, want 0", got)
	}
	input.MerchantCode = "px_mart"
	input.CategoryCode = "online"
	if got := len(recommendOK(t, input).Recommendations); got != 0 {
		t.Fatalf("non-matching category recommendations = %d, want 0", got)
	}
}

func TestEmptyMerchantKeywordsApplyToEveryMerchant(t *testing.T) {
	input := baseInput()
	input.MerchantName = "任何店家"
	input.Rules = []RewardRule{rule("all-merchants", "card-a", cash, "0.01", nil, "dining")}
	if got := len(recommendOK(t, input).Recommendations); got != 1 {
		t.Fatalf("recommendations = %d, want 1", got)
	}
}

func TestPaymentMethodRestrictions(t *testing.T) {
	input := baseInput()
	restricted := rule("mobile", "card-a", cash, "0.1", nil, "general")
	restricted.PaymentMethods = []string{"apple_pay", "samsung_pay"}
	input.Rules = []RewardRule{restricted}

	input.PaymentMethods = []PaymentMethod{{Code: "apple_pay", Name: "Apple Pay"}}
	if got := len(recommendOK(t, input).Recommendations); got != 1 {
		t.Fatalf("matching payment recommendations = %d, want 1", got)
	}
	input.PaymentMethods = []PaymentMethod{{Code: "line_pay", Name: "LINE Pay"}}
	if got := len(recommendOK(t, input).Recommendations); got != 0 {
		t.Fatalf("non-matching payment recommendations = %d, want 0", got)
	}

	restricted.PaymentMethods = nil
	input.Rules = []RewardRule{restricted}
	if got := len(recommendOK(t, input).Recommendations); got != 1 {
		t.Fatalf("unrestricted payment recommendations = %d, want 1", got)
	}
}

func TestPaymentMethodsGroupSameScoreAndSplitDifferentScore(t *testing.T) {
	input := baseInput()
	input.PaymentMethods = []PaymentMethod{
		{Code: "physical_card", Name: "實體信用卡"},
		{Code: "apple_pay", Name: "Apple Pay"},
		{Code: "line_pay", Name: "LINE Pay"},
	}
	base := rule("base", "card-a", cash, "0.01", nil, "general")
	base.StackGroup = "base"
	bonus := rule("bonus", "card-a", cash, "0.02", nil, "general")
	bonus.StackGroup = "bonus"
	bonus.PaymentMethods = []string{"apple_pay", "line_pay"}
	input.Rules = []RewardRule{base, bonus}

	got := recommendOK(t, input).Recommendations
	if len(got) != 2 {
		t.Fatalf("recommendations=%d,want 2", len(got))
	}
	if got[0].Rank != 1 || len(got[0].PaymentOptions) != 1 || len(got[0].PaymentOptions[0].PaymentMethods) != 2 {
		t.Fatalf("top group=%+v", got[0])
	}
	if got[0].PaymentOptions[0].PaymentMethodName != "Apple Pay、LINE Pay" {
		t.Fatalf("merged payment name=%q", got[0].PaymentOptions[0].PaymentMethodName)
	}
	if got[1].Rank != 2 || len(got[1].PaymentOptions) != 1 || got[1].PaymentOptions[0].PaymentMethodCode != AnyPaymentCode {
		t.Fatalf("second group=%+v", got[1])
	}
}

func TestUnrestrictedRuleProducesOnlyAnyPayment(t *testing.T) {
	input := baseInput()
	input.PaymentMethods = []PaymentMethod{{Code: "physical_card", Name: "實體信用卡"}, {Code: "apple_pay", Name: "Apple Pay"}}
	input.Rules = []RewardRule{rule("base", "card-a", cash, "0.01", nil, "general")}

	got := recommendOK(t, input).Recommendations
	if len(got) != 1 || len(got[0].PaymentOptions) != 1 || got[0].PaymentOptions[0].PaymentMethodCode != AnyPaymentCode {
		t.Fatalf("unexpected unrestricted options: %+v", got)
	}
}

func TestExclusivePaymentRuleSuppressesGeneralStacks(t *testing.T) {
	input := baseInput()
	input.PaymentMethods = []PaymentMethod{{Code: "easycard", Name: "悠遊卡功能"}}
	base := rule("base", "card-a", cash, "0.01", nil, "general")
	base.StackGroup = "base"
	bonus := rule("bonus", "card-a", cash, "0.025", nil, "general")
	bonus.StackGroup = "tier_bonus"
	easycard := rule("easycard", "card-a", cash, "0.03", decimalPointer("100"), "general")
	easycard.PaymentMethods = []string{"easycard"}
	easycard.StackGroup = ExclusiveStackGroupPrefix + "easycard_autoload"
	input.Rules = []RewardRule{base, bonus, easycard}

	got := recommendOK(t, input).Recommendations[0]
	if len(got.Allocations) != 1 || got.Allocations[0].RuleID != "easycard" {
		t.Fatalf("allocations=%+v,want only easycard exclusive rule", got.Allocations)
	}
	assertDecimal(t, got.TotalUnweighted, "30.00")
}

func TestOtherMerchantNameDoesNotMatchRestrictedMerchant(t *testing.T) {
	input := baseInput()
	input.MerchantCode = ""
	input.MerchantName = "全聯福利中心"
	restricted := rule("px", "card-a", cash, "0.1", nil, "general")
	restricted.MerchantCodes = []string{"px_mart"}
	input.Rules = []RewardRule{restricted}

	if got := len(recommendOK(t, input).Recommendations); got != 0 {
		t.Fatalf("recommendations=%d,want 0", got)
	}
}

func TestIgnoresOtherUsersCardsAndRules(t *testing.T) {
	input := baseInput()
	otherCard := Card{ID: "other-card", UserID: "user-b", Name: "Other", IsActive: true}
	otherRule := rule("other-rule", otherCard.ID, cash, "99", nil, "general")
	otherRule.UserID = "user-b"
	input.Cards = append(input.Cards, otherCard)
	input.Rules = []RewardRule{
		rule("owned-rule", "card-a", cash, "0.01", nil, "general"),
		otherRule,
	}

	result := recommendOK(t, input)
	if len(result.Recommendations) != 1 || result.Recommendations[0].CardID != "card-a" {
		t.Fatalf("recommendations included another user's data: %+v", result.Recommendations)
	}
}

func TestSelectsActivityByDateAndStacksAcrossGroups(t *testing.T) {
	input := baseInput()
	old := rule("old", "card-a", cash, "0.99", nil, "general")
	old.ActivityID = "old-act"
	old.StartDate = localDatePointer("2025-01-01")
	old.EndDate = localDatePointer("2025-12-31")
	base := rule("base", "card-a", cash, "0.01", nil, "general")
	base.ActivityID = "current"
	base.StackGroup = "base"
	base.StartDate = localDatePointer("2026-01-01")
	base.EndDate = localDatePointer("2026-06-30")
	low := rule("low", "card-a", cash, "0.02", nil, "general")
	low.ActivityID = "current"
	low.StackGroup = "bonus"
	low.StartDate = base.StartDate
	low.EndDate = base.EndDate
	high := rule("high", "card-a", cash, "0.03", nil, "general")
	high.ActivityID = "current"
	high.StackGroup = "bonus"
	high.StartDate = base.StartDate
	high.EndDate = base.EndDate
	input.Rules = []RewardRule{old, base, low, high}
	got := recommendOK(t, input).Recommendations[0]
	if len(got.Allocations) != 2 {
		t.Fatalf("allocations=%d,want 2", len(got.Allocations))
	}
	assertDecimal(t, got.TotalUnweighted, "40.00")
}

func TestAccountTierSharedCapAndReminder(t *testing.T) {
	input := baseInput()
	input.Cards[0].AccountTier = "大戶Plus"
	input.AmountMinor = 100000
	base := rule("base", "card-a", cash, "0.01", nil, "general")
	base.ActivityID = "act"
	base.StackGroup = "base"
	base.RequiredAccountTiers = []string{"大戶"}
	bonus := rule("bonus", "card-a", cash, "0.05", nil, "general")
	bonus.ActivityID = "act"
	bonus.StackGroup = "bonus"
	bonus.RequiredAccountTiers = []string{"大戶Plus"}
	bonus.ActionRequired = "registration"
	bonus.ActionMessage = "請先登錄"
	bonus.SharedMonthlyCap = decimalPointer("30")
	input.ActivityMonthlyUsage = map[string]Decimal{"act:cash": MustDecimal("10")}
	input.Rules = []RewardRule{base, bonus}
	got := recommendOK(t, input).Recommendations[0]
	if len(got.Allocations) != 1 || len(got.Reminders) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
	assertDecimal(t, got.Allocations[0].AllocatedReward, "20.00")
}

func baseInput() RecommendationInput {
	return RecommendationInput{
		UserID:         "user-a",
		AmountMinor:    100000,
		CategoryCode:   "dining",
		MerchantName:   "測試餐廳",
		PaymentMethods: []PaymentMethod{{Code: "physical_card", Name: "實體信用卡"}},
		Date:           MustLocalDate("2026-06-10"),
		Cards:          []Card{{ID: "card-a", UserID: "user-a", Name: "A Card", IsActive: true}},
		Preferences: map[ID]Decimal{
			"cash":   MustDecimal("1"),
			"points": MustDecimal("1"),
		},
		MonthlyUsage: map[ID]Decimal{},
	}
}

func rule(id, cardID ID, unit RewardUnit, rate string, cap *Decimal, categories ...string) RewardRule {
	return RewardRule{
		ID: id, UserID: "user-a", CardID: cardID, Name: string(id), RewardUnit: unit,
		Rate: MustDecimal(rate), MonthlyCap: cap, IsActive: true, CategoryCode: categories,
	}
}

func decimalPointer(value string) *Decimal {
	decimal := MustDecimal(value)
	return &decimal
}

func localDatePointer(value string) *LocalDate { v := MustLocalDate(value); return &v }

func recommendOK(t *testing.T, input RecommendationInput) Result {
	t.Helper()
	result, err := Recommend(input)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertDecimal(t *testing.T, actual Decimal, expected string) {
	t.Helper()
	if actual.Cmp(MustDecimal(expected)) != 0 {
		t.Fatalf("decimal = %s, want %s", actual.String(), expected)
	}
}
