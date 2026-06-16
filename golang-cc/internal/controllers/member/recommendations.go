package member

import (
	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/member"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

type recommendationRequest struct {
	AmountMinor  int64  `json:"amount_minor" binding:"required,gt=0"`
	CategoryCode string `json:"category_code" binding:"required"`
	MerchantCode string `json:"merchant_code"`
	MerchantName string `json:"merchant_name"`
	Date         string `json:"date" binding:"required"`
}

type rewardUnitResponse struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Symbol         string `json:"symbol"`
	SymbolPosition string `json:"symbol_position"`
	TWDRate        string `json:"twd_rate"`
	Precision      int    `json:"precision"`
}

type allocationResponse struct {
	RuleID           string             `json:"rule_id"`
	RuleName         string             `json:"rule_name"`
	ActivityID       string             `json:"activity_id"`
	ActivityName     string             `json:"activity_name"`
	StackGroup       string             `json:"stack_group"`
	ActionRequired   string             `json:"action_required"`
	ActionMessage    string             `json:"action_message"`
	RewardUnit       rewardUnitResponse `json:"reward_unit"`
	Rate             string             `json:"rate"`
	UncappedReward   string             `json:"uncapped_reward"`
	MonthlyCap       *string            `json:"monthly_cap"`
	UsedBefore       string             `json:"used_before"`
	RemainingBefore  *string            `json:"remaining_before"`
	AllocatedReward  string             `json:"allocated_reward"`
	PreferenceWeight string             `json:"preference_weight"`
	Score            string             `json:"score"`
}

type recommendationResponse struct {
	Rank            int                     `json:"rank"`
	CardID          string                  `json:"card_id"`
	CardName        string                  `json:"card_name"`
	TotalScore      string                  `json:"total_score"`
	TotalUnweighted string                  `json:"total_unweighted"`
	Allocations     []allocationResponse    `json:"allocations"`
	Explanation     []string                `json:"explanation"`
	Reminders       []string                `json:"reminders"`
	PaymentOptions  []paymentOptionResponse `json:"payment_options"`
}

type paymentOptionResponse struct {
	PaymentMethodCode string                  `json:"payment_method_code"`
	PaymentMethodName string                  `json:"payment_method_name"`
	PaymentMethods    []paymentMethodResponse `json:"payment_methods"`
	TotalScore        string                  `json:"total_score"`
	TotalUnweighted   string                  `json:"total_unweighted"`
	Allocations       []allocationResponse    `json:"allocations"`
	Explanation       []string                `json:"explanation"`
	Reminders         []string                `json:"reminders"`
}

func (c *Controller) Recommend(ctx *gin.Context) {
	var request recommendationRequest
	if ctx.ShouldBindJSON(&request) != nil {
		failure(ctx, 400, "validation_failed", "輸入格式錯誤")
		return
	}
	date, err := recommendations.ParseLocalDate(request.Date)
	if err != nil {
		failure(ctx, 400, "validation_failed", "日期格式錯誤")
		return
	}
	result, err := c.service.Recommend(ctx, userID(ctx), service.RecommendationInput{
		AmountMinor: request.AmountMinor, CategoryCode: request.CategoryCode,
		MerchantCode: request.MerchantCode, MerchantName: request.MerchantName, Date: date,
	})
	if err != nil {
		failure(ctx, 400, "validation_failed", err.Error())
		return
	}
	items := make([]recommendationResponse, 0, len(result.Recommendations))
	for _, item := range result.Recommendations {
		items = append(items, mapRecommendation(item))
	}
	data(ctx, 200, gin.H{"input": request, "recommendations": items, "empty_reason": result.EmptyReason})
}

func mapRecommendation(item recommendations.CardRecommendation) recommendationResponse {
	allocations := make([]allocationResponse, 0, len(item.Allocations))
	for _, allocation := range item.Allocations {
		allocations = append(allocations, mapAllocation(allocation))
	}
	options := make([]paymentOptionResponse, 0, len(item.PaymentOptions))
	for _, option := range item.PaymentOptions {
		optionAllocations := make([]allocationResponse, 0, len(option.Allocations))
		for _, allocation := range option.Allocations {
			optionAllocations = append(optionAllocations, mapAllocation(allocation))
		}
		methods := make([]paymentMethodResponse, 0, len(option.PaymentMethods))
		for _, method := range option.PaymentMethods {
			methods = append(methods, paymentMethodResponse{Code: method.Code, Name: method.Name})
		}
		options = append(options, paymentOptionResponse{
			PaymentMethodCode: option.PaymentMethodCode, PaymentMethodName: option.PaymentMethodName,
			PaymentMethods: methods,
			TotalScore:     option.TotalScore.String(), TotalUnweighted: option.TotalUnweighted.String(),
			Allocations: optionAllocations, Explanation: option.Explanation, Reminders: option.Reminders,
		})
	}
	return recommendationResponse{
		Rank: item.Rank, CardID: string(item.CardID), CardName: item.CardName,
		TotalScore: item.TotalScore.String(), TotalUnweighted: item.TotalUnweighted.String(),
		Allocations: allocations, Explanation: item.Explanation, Reminders: item.Reminders, PaymentOptions: options,
	}
}

func mapAllocation(item recommendations.RuleEvaluation) allocationResponse {
	return allocationResponse{
		RuleID: string(item.RuleID), RuleName: item.RuleName, ActivityID: string(item.ActivityID), ActivityName: item.ActivityName, StackGroup: item.StackGroup, ActionRequired: item.ActionRequired, ActionMessage: item.ActionMessage,
		RewardUnit: rewardUnitResponse{
			ID: string(item.RewardUnit.ID), Code: item.RewardUnit.Code, Name: item.RewardUnit.Name,
			Symbol: item.RewardUnit.Symbol, SymbolPosition: item.RewardUnit.SymbolPosition,
			TWDRate: item.RewardUnit.TWDRate.String(), Precision: item.RewardUnit.Precision,
		},
		Rate: item.Rate.String(), UncappedReward: item.UncappedReward.String(),
		MonthlyCap: decimalPointer(item.MonthlyCap), UsedBefore: item.UsedBefore.String(),
		RemainingBefore: decimalPointer(item.RemainingBefore), AllocatedReward: item.AllocatedReward.String(),
		PreferenceWeight: item.PreferenceWeight.String(), Score: item.Score.String(),
	}
}

func decimalPointer(value *recommendations.Decimal) *string {
	if value == nil {
		return nil
	}
	result := value.String()
	return &result
}
