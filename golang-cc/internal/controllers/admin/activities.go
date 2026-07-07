package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/domain"
	service "github.com/rafa/golang-cc/internal/services/admin"
)

type activityFlowRequest struct {
	Activity     activitySummaryRequest `json:"activity" binding:"required"`
	RewardGroups []rewardGroupRequest   `json:"reward_groups" binding:"required,min=1"`
}

type activitySummaryRequest struct {
	BankID        string `json:"bank_id" binding:"required"`
	CardProductID string `json:"card_product_id" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	SourceURL     string `json:"source_url"`
	EffectiveFrom string `json:"effective_from" binding:"required"`
	EffectiveTo   string `json:"effective_to" binding:"required"`
	IsActive      bool   `json:"is_active"`
}

type rewardGroupRequest struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name" binding:"required"`
	Description  string                   `json:"description"`
	DisplayOrder int                      `json:"display_order"`
	IsActive     bool                     `json:"is_active"`
	Components   []rewardComponentRequest `json:"components" binding:"required,min=1"`
}

type rewardComponentRequest struct {
	ID             string                     `json:"id"`
	RewardGroupIDs []string                   `json:"reward_group_ids" binding:"required,min=1"`
	Name           string                     `json:"name" binding:"required"`
	Description    string                     `json:"description"`
	Layer          int                        `json:"layer" binding:"required,min=1"`
	StackGroup     string                     `json:"stack_group" binding:"required"`
	StackMode      string                     `json:"stack_mode" binding:"required"`
	Priority       int                        `json:"priority"`
	EffectiveFrom  string                     `json:"effective_from"`
	EffectiveTo    string                     `json:"effective_to"`
	IsActive       bool                       `json:"is_active"`
	Requirements   []rewardRequirementRequest `json:"requirements" binding:"required,min=1"`
	Benefits       []rewardBenefitRequest     `json:"benefits" binding:"required,min=1"`
}

type rewardRequirementRequest struct {
	ID            string          `json:"id"`
	Type          string          `json:"requirement_type" binding:"required"`
	Operator      string          `json:"operator" binding:"required"`
	Configuration json.RawMessage `json:"configuration_json" binding:"required"`
	Description   string          `json:"description"`
	IsActive      bool            `json:"is_active"`
}

type rewardBenefitRequest struct {
	ID           string  `json:"id"`
	BenefitType  string  `json:"benefit_type" binding:"required"`
	Value        string  `json:"value" binding:"required"`
	RewardUnitID string  `json:"reward_unit_id" binding:"required"`
	CapAmount    *string `json:"cap_amount"`
	CapFormula   *string `json:"cap_formula"`
	CapPeriod    *string `json:"cap_period"`
	Description  string  `json:"description"`
	IsActive     bool    `json:"is_active"`
}

type activityStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type activitySummaryResponse struct {
	ID             string `json:"id"`
	BankID         string `json:"bank_id"`
	BankName       string `json:"bank_name"`
	CardProductID  string `json:"card_product_id"`
	CardName       string `json:"card_name"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	SourceURL      string `json:"source_url"`
	EffectiveFrom  string `json:"effective_from"`
	EffectiveTo    string `json:"effective_to"`
	IsActive       bool   `json:"is_active"`
	GroupCount     int64  `json:"group_count"`
	ComponentCount int64  `json:"component_count"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type activityFlowResponse struct {
	Activity     activitySummaryResponse `json:"activity"`
	RewardGroups []rewardGroupResponse   `json:"reward_groups"`
}

type rewardGroupResponse struct {
	ID           string                    `json:"id"`
	ActivityID   string                    `json:"activity_id"`
	Name         string                    `json:"name"`
	Description  string                    `json:"description"`
	DisplayOrder int                       `json:"display_order"`
	IsActive     bool                      `json:"is_active"`
	Components   []rewardComponentResponse `json:"components"`
}

type rewardComponentResponse struct {
	ID             string                      `json:"id"`
	RewardGroupIDs []string                    `json:"reward_group_ids"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	Layer          int                         `json:"layer"`
	StackGroup     string                      `json:"stack_group"`
	StackMode      string                      `json:"stack_mode"`
	Priority       int                         `json:"priority"`
	EffectiveFrom  string                      `json:"effective_from"`
	EffectiveTo    string                      `json:"effective_to"`
	IsActive       bool                        `json:"is_active"`
	Requirements   []rewardRequirementResponse `json:"requirements"`
	Benefits       []rewardBenefitResponse     `json:"benefits"`
}

type rewardRequirementResponse struct {
	ID                string          `json:"id"`
	RewardComponentID string          `json:"reward_component_id"`
	Type              string          `json:"requirement_type"`
	Operator          string          `json:"operator"`
	Configuration     json.RawMessage `json:"configuration_json"`
	Description       string          `json:"description"`
	IsActive          bool            `json:"is_active"`
}

type rewardBenefitResponse struct {
	ID                string  `json:"id"`
	RewardComponentID string  `json:"reward_component_id"`
	BenefitType       string  `json:"benefit_type"`
	Value             string  `json:"value"`
	RewardUnitID      string  `json:"reward_unit_id"`
	CapAmount         *string `json:"cap_amount"`
	CapFormula        *string `json:"cap_formula"`
	CapPeriod         *string `json:"cap_period"`
	Description       string  `json:"description"`
	IsActive          bool    `json:"is_active"`
}

func (r activityFlowRequest) serviceInput() service.ActivityFlowInput {
	groups := make([]service.RewardGroupInput, 0, len(r.RewardGroups))
	for _, group := range r.RewardGroups {
		components := make([]service.RewardComponentInput, 0, len(group.Components))
		for _, component := range group.Components {
			requirements := make([]service.RewardRequirementInput, 0, len(component.Requirements))
			for _, requirement := range component.Requirements {
				requirements = append(requirements, service.RewardRequirementInput{ID: requirement.ID, RequirementType: requirement.Type, Operator: requirement.Operator, Configuration: requirement.Configuration, Description: requirement.Description, IsActive: requirement.IsActive})
			}
			benefits := make([]service.RewardBenefitInput, 0, len(component.Benefits))
			for _, benefit := range component.Benefits {
				benefits = append(benefits, service.RewardBenefitInput{ID: benefit.ID, BenefitType: benefit.BenefitType, Value: benefit.Value, RewardUnitID: benefit.RewardUnitID, CapAmount: benefit.CapAmount, CapFormula: benefit.CapFormula, CapPeriod: benefit.CapPeriod, Description: benefit.Description, IsActive: benefit.IsActive})
			}
			components = append(components, service.RewardComponentInput{ID: component.ID, RewardGroupIDs: component.RewardGroupIDs, Name: component.Name, Description: component.Description, Layer: component.Layer, StackGroup: component.StackGroup, StackMode: component.StackMode, Priority: component.Priority, EffectiveFrom: component.EffectiveFrom, EffectiveTo: component.EffectiveTo, IsActive: component.IsActive, Requirements: requirements, Benefits: benefits})
		}
		groups = append(groups, service.RewardGroupInput{ID: group.ID, Name: group.Name, Description: group.Description, DisplayOrder: group.DisplayOrder, IsActive: group.IsActive, Components: components})
	}
	return service.ActivityFlowInput{Activity: service.ActivitySummaryInput{BankID: r.Activity.BankID, CardProductID: r.Activity.CardProductID, Title: r.Activity.Title, Description: r.Activity.Description, SourceURL: r.Activity.SourceURL, EffectiveFrom: r.Activity.EffectiveFrom, EffectiveTo: r.Activity.EffectiveTo, IsActive: r.Activity.IsActive}, RewardGroups: groups}
}

func (c *Controller) ListActivities(ctx *gin.Context) {
	filters := service.ActivityFilters{BankID: ctx.Query("bank_id"), CardProductID: ctx.Query("card_product_id")}
	if raw := ctx.Query("is_active"); raw != "" {
		active := raw == "true" || raw == "1"
		filters.IsActive = &active
	}
	items, err := c.service.ListActivities(ctx.Request.Context(), filters)
	if err != nil {
		println("ListActivities error:", err.Error())
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	out := make([]activitySummaryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapActivitySummary(item))
	}
	data(ctx, http.StatusOK, out)
}

func (c *Controller) GetActivity(ctx *gin.Context) {
	item, err := c.service.GetActivity(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到活動")
		return
	}
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapActivityFlow(item))
}

func (c *Controller) CreateActivity(ctx *gin.Context) {
	var request activityFlowRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "活動資料無效："+err.Error())
		return
	}
	id, err := c.service.CreateActivity(ctx.Request.Context(), request.serviceInput())
	if err != nil {
		println("CreateActivity error:", err.Error())
		failure(ctx, http.StatusBadRequest, "validation_failed", activityValidationMessage(err))
		return
	}
	data(ctx, http.StatusCreated, gin.H{"id": id})
}

func (c *Controller) UpdateActivity(ctx *gin.Context) {
	var request activityFlowRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "活動資料無效："+err.Error())
		return
	}
	err := c.service.UpdateActivity(ctx.Request.Context(), ctx.Param("id"), request.serviceInput())
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到活動")
		return
	}
	if err != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", activityValidationMessage(err))
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}

func activityValidationMessage(err error) string {
	if err == nil {
		return "活動資料無效"
	}
	if errors.Is(err, domain.ErrInvalidInput) && err.Error() != domain.ErrInvalidInput.Error() {
		return err.Error()
	}
	return "活動資料無效"
}

func (c *Controller) SetActivityStatus(ctx *gin.Context) {
	var request activityStatusRequest
	if ctx.ShouldBindJSON(&request) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "活動狀態無效")
		return
	}
	err := c.service.SetActivityStatus(ctx.Request.Context(), ctx.Param("id"), request.IsActive)
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到活動")
		return
	}
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "更新失敗")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}

func (c *Controller) DeleteActivity(ctx *gin.Context) {
	err := c.service.DeleteActivity(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到活動")
		return
	}
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "刪除失敗")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}

func (c *Controller) ActivityRequirementOptions(ctx *gin.Context) {
	items, err := c.service.ActivityRequirementOptions(ctx.Request.Context(), service.RequirementOptionFilters{RequirementType: ctx.Query("requirement_type"), CardProductID: ctx.Query("card_product_id")})
	if err != nil {
		println("ActivityRequirementOptions error:", err.Error())
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapRequirementOptions(items))
}

func mapActivitySummary(item domain.ActivitySummary) activitySummaryResponse {
	return activitySummaryResponse{ID: item.ID, BankID: item.BankID, BankName: item.BankName, CardProductID: item.CardProductID, CardName: item.CardName, Title: item.Title, Description: item.Description, SourceURL: item.SourceURL, EffectiveFrom: item.EffectiveFrom, EffectiveTo: item.EffectiveTo, IsActive: item.IsActive, GroupCount: item.GroupCount, ComponentCount: item.ComponentCount, CreatedAt: item.CreatedAt.Format("2006-01-02"), UpdatedAt: item.UpdatedAt.Format("2006-01-02")}
}

func mapActivityFlow(item domain.ActivityFlow) activityFlowResponse {
	groups := make([]rewardGroupResponse, 0, len(item.RewardGroups))
	for _, group := range item.RewardGroups {
		components := make([]rewardComponentResponse, 0, len(group.Components))
		for _, component := range group.Components {
			requirements := make([]rewardRequirementResponse, 0, len(component.Requirements))
			for _, requirement := range component.Requirements {
				requirements = append(requirements, rewardRequirementResponse{ID: requirement.ID, RewardComponentID: requirement.RewardComponentID, Type: requirement.RequirementType, Operator: requirement.Operator, Configuration: requirement.Configuration, Description: requirement.Description, IsActive: requirement.IsActive})
			}
			benefits := make([]rewardBenefitResponse, 0, len(component.Benefits))
			for _, benefit := range component.Benefits {
				benefits = append(benefits, rewardBenefitResponse{ID: benefit.ID, RewardComponentID: benefit.RewardComponentID, BenefitType: benefit.BenefitType, Value: benefit.Value, RewardUnitID: benefit.RewardUnitID, CapAmount: benefit.CapAmount, CapFormula: benefit.CapFormula, CapPeriod: benefit.CapPeriod, Description: benefit.Description, IsActive: benefit.IsActive})
			}
			components = append(components, rewardComponentResponse{ID: component.ID, RewardGroupIDs: component.RewardGroupIDs, Name: component.Name, Description: component.Description, Layer: component.Layer, StackGroup: component.StackGroup, StackMode: component.StackMode, Priority: component.Priority, EffectiveFrom: component.EffectiveFrom, EffectiveTo: component.EffectiveTo, IsActive: component.IsActive, Requirements: requirements, Benefits: benefits})
		}
		groups = append(groups, rewardGroupResponse{ID: group.ID, ActivityID: group.ActivityID, Name: group.Name, Description: group.Description, DisplayOrder: group.DisplayOrder, IsActive: group.IsActive, Components: components})
	}
	return activityFlowResponse{Activity: mapActivitySummary(item.Activity), RewardGroups: groups}
}

type codeNameResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type requirementTypeOptionResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	ValueKey    string `json:"value_key"`
	ValueSource string `json:"value_source"`
}

type cardPlanOptionResponse struct {
	ID            string `json:"id"`
	CardProductID string `json:"card_product_id"`
	PlanType      string `json:"plan_type"`
	Name          string `json:"name"`
}

func mapRequirementOptions(items domain.RequirementOptionSet) gin.H {
	return gin.H{
		"operators":           mapCodeNameOptions(items.Operators),
		"payment_methods":     mapCodeNameOptions(items.PaymentMethods),
		"merchants":           mapCodeNameOptions(items.Merchants),
		"categories":          mapCodeNameOptions(items.Categories),
		"card_networks":       mapCodeNameOptions(items.CardNetworks),
		"card_products":       mapCodeNameOptions(items.CardProducts),
		"card_plans":          mapCardPlanOptions(items.CardPlans),
		"account_tiers":       mapCodeNameOptions(items.AccountTiers),
		"user_qualifications": mapCodeNameOptions(items.UserQualifications),
		"channels":            mapCodeNameOptions(items.Channels),
		"regions":             mapCodeNameOptions(items.Regions),
		"currencies":          mapCodeNameOptions(items.Currencies),
	}
}

func mapCodeNameOptions(values []domain.CodeNameOption) []codeNameResponse {
	out := make([]codeNameResponse, 0, len(values))
	for _, value := range values {
		out = append(out, codeNameResponse{Code: value.Code, Name: value.Name})
	}
	return out
}

func mapRequirementTypeOptions(values []domain.RequirementTypeOption) []requirementTypeOptionResponse {
	out := make([]requirementTypeOptionResponse, 0, len(values))
	for _, value := range values {
		out = append(out, requirementTypeOptionResponse{Code: value.Code, Name: value.Name, ValueKey: value.ValueKey, ValueSource: value.ValueSource})
	}
	return out
}

func mapCardPlanOptions(values []domain.CardPlanOption) []cardPlanOptionResponse {
	out := make([]cardPlanOptionResponse, 0, len(values))
	for _, value := range values {
		out = append(out, cardPlanOptionResponse{ID: value.ID, CardProductID: value.CardProductID, PlanType: value.PlanType, Name: value.Name})
	}
	return out
}
