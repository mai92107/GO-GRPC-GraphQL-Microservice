package member

import (
	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/member"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

type recommendationRequest struct {
	AmountMinor  int64  `json:"amount_minor" binding:"required,gt=0"`
	CategoryID   string `json:"category_id" binding:"required"`
	MerchantID   string `json:"merchant_id"`
	MerchantName string `json:"merchant_name"`
	Date         string `json:"date" binding:"required"`
}

type rewardUnitResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Symbol         string `json:"symbol"`
	SymbolPosition string `json:"symbol_position"`
	TWDRate        string `json:"twd_rate"`
	Precision      int    `json:"precision"`
}

type allocationResponse struct {
	RuleID              string             `json:"rule_id"`
	RuleName            string             `json:"rule_name"`
	ActivityID          string             `json:"activity_id"`
	ActivityName        string             `json:"activity_name"`
	StackGroup          string             `json:"stack_group"`
	Layer               string             `json:"layer"`
	DisplayOrder        int                `json:"display_order"`
	EffectType          string             `json:"effect_type"`
	RewardValue         string             `json:"reward_value"`
	ActionRequired      string             `json:"action_required"`
	ActionMessage       string             `json:"action_message"`
	RewardUnit          rewardUnitResponse `json:"reward_unit"`
	RewardRate          string             `json:"reward_rate"`
	UncappedReward      string             `json:"uncapped_reward"`
	MonthlyCap          *string            `json:"monthly_cap"`
	UsedBefore          string             `json:"used_before"`
	RemainingBefore     *string            `json:"remaining_before"`
	AllocatedReward     string             `json:"allocated_reward"`
	PreferenceWeight    string             `json:"preference_weight"`
	Score               string             `json:"score"`
	SuggestedCardPlanID string             `json:"suggested_card_plan_id,omitempty"`
	SuggestedPlanName   string             `json:"suggested_plan_name,omitempty"`
}

type benefitResponse struct {
	BenefitID       string  `json:"benefit_id"`
	Name            string  `json:"name"`
	ActivityID      string  `json:"activity_id"`
	ActivityName    string  `json:"activity_name"`
	EffectType      string  `json:"effect_type"`
	RewardValue     string  `json:"reward_value"`
	RewardRate      string  `json:"reward_rate"`
	RewardAmount    string  `json:"reward_amount"`
	StackGroup      string  `json:"stack_group"`
	MonthlyCap      *string `json:"monthly_cap"`
	RemainingBefore *string `json:"remaining_before"`
	IsApplied       bool    `json:"is_applied"`
	ExcludeReason   string  `json:"exclude_reason,omitempty"`
}

type layerResponse struct {
	Layer        string            `json:"layer"`
	RewardRate   string            `json:"reward_rate"`
	RewardAmount string            `json:"reward_amount"`
	Benefits     []benefitResponse `json:"benefits"`
}

type excludedBenefitResponse struct {
	BenefitID    string `json:"benefit_id"`
	Name         string `json:"name"`
	ActivityID   string `json:"activity_id"`
	ActivityName string `json:"activity_name"`
	Reason       string `json:"reason"`
	Layer        string `json:"layer"`
	StackGroup   string `json:"stack_group"`
}

type recommendationResponse struct {
	Rank             int                       `json:"rank"`
	CardID           string                    `json:"card_id"`
	CardName         string                    `json:"card_name"`
	TotalScore       string                    `json:"total_score"`
	TotalUnweighted  string                    `json:"total_unweighted"`
	Layers           []layerResponse           `json:"layers"`
	ExcludedBenefits []excludedBenefitResponse `json:"excluded_benefits"`
	Allocations      []allocationResponse      `json:"allocations"`
	Explanation      []string                  `json:"explanation"`
	Reminders        []string                  `json:"reminders"`
	PaymentOptions   []paymentOptionResponse   `json:"payment_options"`
}

type paymentOptionResponse struct {
	PaymentMethodID   string                    `json:"payment_method_id"`
	PaymentMethodName string                    `json:"payment_method_name"`
	PaymentMethods    []paymentMethodResponse   `json:"payment_methods"`
	TotalScore        string                    `json:"total_score"`
	TotalUnweighted   string                    `json:"total_unweighted"`
	Allocations       []allocationResponse      `json:"allocations"`
	Layers            []layerResponse           `json:"layers"`
	ExcludedBenefits  []excludedBenefitResponse `json:"excluded_benefits"`
	Explanation       []string                  `json:"explanation"`
	Reminders         []string                  `json:"reminders"`
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
		AmountMinor: request.AmountMinor, CategoryID: request.CategoryID,
		MerchantID: request.MerchantID, MerchantName: request.MerchantName, Date: date,
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
			methods = append(methods, paymentMethodResponse{ID: method.ID, Name: method.Name})
		}
		options = append(options, paymentOptionResponse{
			PaymentMethodID: option.PaymentMethodID, PaymentMethodName: option.PaymentMethodName,
			PaymentMethods: methods,
			TotalScore:     option.TotalScore.String(), TotalUnweighted: option.TotalUnweighted.String(),
			Allocations: optionAllocations,
			Layers:      mapLayers(option.Layers), ExcludedBenefits: mapExcludedBenefits(option.ExcludedBenefits),
			Explanation: option.Explanation, Reminders: option.Reminders,
		})
	}
	return recommendationResponse{
		Rank: item.Rank, CardID: string(item.CardID), CardName: item.CardName,
		TotalScore: item.TotalScore.String(), TotalUnweighted: item.TotalUnweighted.String(),
		Layers: mapLayers(item.Layers), ExcludedBenefits: mapExcludedBenefits(item.ExcludedBenefits),
		Allocations: allocations, Explanation: item.Explanation, Reminders: item.Reminders, PaymentOptions: options,
	}
}

func mapAllocation(item recommendations.RuleEvaluation) allocationResponse {
	return allocationResponse{
		RuleID: string(item.RuleID), RuleName: item.RuleName, ActivityID: string(item.ActivityID), ActivityName: item.ActivityName,
		StackGroup: item.StackGroup, Layer: item.Layer, DisplayOrder: item.DisplayOrder,
		EffectType: string(item.EffectType), RewardValue: item.RewardValue.String(),
		ActionRequired: item.ActionRequired, ActionMessage: item.ActionMessage,
		RewardUnit: rewardUnitResponse{
			ID: string(item.RewardUnit.ID), Name: item.RewardUnit.Name,
			Symbol: item.RewardUnit.Symbol, SymbolPosition: item.RewardUnit.SymbolPosition,
			TWDRate: item.RewardUnit.TWDRate.String(), Precision: item.RewardUnit.Precision,
		},
		RewardRate: item.RewardRate.String(), UncappedReward: item.UncappedReward.String(),
		MonthlyCap: decimalPointer(item.MonthlyCap), UsedBefore: item.UsedBefore.String(),
		RemainingBefore: decimalPointer(item.RemainingBefore), AllocatedReward: item.AllocatedReward.String(),
		PreferenceWeight: item.PreferenceWeight.String(), Score: item.Score.String(),
		SuggestedCardPlanID: string(item.SuggestedCardPlanID), SuggestedPlanName: item.SuggestedPlanName,
	}
}

func mapLayers(items []recommendations.LayerResult) []layerResponse {
	layers := make([]layerResponse, 0, len(items))
	for _, item := range items {
		benefits := make([]benefitResponse, 0, len(item.Benefits))
		for _, benefit := range item.Benefits {
			benefits = append(benefits, benefitResponse{
				BenefitID: string(benefit.BenefitID), Name: benefit.Name,
				ActivityID: string(benefit.ActivityID), ActivityName: benefit.ActivityName,
				EffectType: string(benefit.EffectType), RewardValue: benefit.RewardValue.String(),
				RewardRate: benefit.RewardRate.String(), RewardAmount: benefit.RewardAmount.String(),
				StackGroup: benefit.StackGroup, MonthlyCap: decimalPointer(benefit.MonthlyCap),
				RemainingBefore: decimalPointer(benefit.RemainingBefore),
				IsApplied:       benefit.IsApplied, ExcludeReason: string(benefit.ExcludeReason),
			})
		}
		layers = append(layers, layerResponse{
			Layer: item.Layer, RewardRate: item.RewardRate.String(),
			RewardAmount: item.RewardAmount.String(), Benefits: benefits,
		})
	}
	return layers
}

func mapExcludedBenefits(items []recommendations.ExcludedBenefit) []excludedBenefitResponse {
	excluded := make([]excludedBenefitResponse, 0, len(items))
	for _, item := range items {
		excluded = append(excluded, excludedBenefitResponse{
			BenefitID: string(item.BenefitID), Name: item.Name,
			ActivityID: string(item.ActivityID), ActivityName: item.ActivityName,
			Reason: string(item.Reason), StackGroup: item.StackGroup,
			Layer: item.Layer,
		})
	}
	return excluded
}

func decimalPointer(value *recommendations.Decimal) *string {
	if value == nil {
		return nil
	}
	result := value.String()
	return &result
}
