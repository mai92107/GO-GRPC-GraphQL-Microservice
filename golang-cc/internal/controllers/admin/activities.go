package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/admin"
)

type benefitRequest struct {
	ID             string   `json:"id"`
	RewardUnitID   string   `json:"reward_unit_id" binding:"required"`
	Name           string   `json:"name" binding:"required"`
	DisplayOrder   int      `json:"display_order"`
	EffectType     string   `json:"effect_type"`
	RewardValue    string   `json:"reward_value"`
	MonthlyCap     *string  `json:"monthly_cap"`
	Layer          string   `json:"layer" binding:"required"`
	StackGroup     string   `json:"stack_group" binding:"required"`
	Priority       int      `json:"priority"`
	QualifiedType  string   `json:"qualified_type"`
	SelectableType string   `json:"selectable_type"`
	ActionRequired string   `json:"action_required"`
	ActionMessage  string   `json:"action_message"`
	PaymentMethods []string `json:"payment_methods"`
	CategoryIDs    []string `json:"category_ids"`
	MerchantIds    []string `json:"merchant_ids"`
}
type activityRequest struct {
	CardProductID     string            `json:"card_product_id" binding:"required"`
	Name              string            `json:"name" binding:"required"`
	StartDate         string            `json:"start_date" binding:"required"`
	EndDate           string            `json:"end_date" binding:"required"`
	IsActive          *bool             `json:"is_active"`
	SourceURL         string            `json:"source_url"`
	VerifiedAt        *string           `json:"verified_at"`
	NetworkIDs        []string          `json:"network_ids" binding:"required,min=1"`
	SharedMonthlyCaps map[string]string `json:"shared_monthly_caps"`
	Benefits          []benefitRequest  `json:"benefits" binding:"required,min=1"`
}

func (r activityRequest) serviceInput() service.ActivityInput {
	bs := make([]service.ActivityBenefitInput, 0, len(r.Benefits))
	for _, b := range r.Benefits {
		bs = append(bs, service.ActivityBenefitInput{ID: b.ID, RewardUnitID: b.RewardUnitID, Name: b.Name, Layer: b.Layer, DisplayOrder: b.DisplayOrder, EffectType: b.EffectType, RewardValue: b.RewardValue, MonthlyCap: b.MonthlyCap, StackGroup: b.StackGroup, Priority: b.Priority, QualifiedType: b.QualifiedType, SelectableType: b.SelectableType, ActionRequired: b.ActionRequired, ActionMessage: b.ActionMessage, PaymentMethods: b.PaymentMethods, CategoryIDs: b.CategoryIDs, MerchantIds: b.MerchantIds})
	}
	return service.ActivityInput{CardProductID: r.CardProductID, Name: r.Name, StartDate: r.StartDate, EndDate: r.EndDate, IsActive: r.IsActive, SourceURL: r.SourceURL, VerifiedAt: r.VerifiedAt, NetworkIDs: r.NetworkIDs, SharedMonthlyCaps: r.SharedMonthlyCaps, Benefits: bs}
}
func (c *Controller) ListActivities(ctx *gin.Context) {
	x, e := c.service.ListActivities(ctx.Request.Context())
	if e != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	data(ctx, 200, mapActivities(x))
}
func (c *Controller) CreateActivity(ctx *gin.Context) {
	var r activityRequest
	if ctx.ShouldBindJSON(&r) != nil {
		failure(ctx, 400, "validation_failed", "回饋方案資料無效")
		return
	}
	id, e := c.service.CreateActivity(ctx.Request.Context(), r.serviceInput())
	if e != nil {
		failure(ctx, 400, "validation_failed", "回饋方案資料無效或期間重疊")
		return
	}
	data(ctx, 201, gin.H{"id": id})
}
func (c *Controller) UpdateActivity(ctx *gin.Context) {
	var r activityRequest
	if ctx.ShouldBindJSON(&r) != nil || r.IsActive == nil {
		failure(ctx, 400, "validation_failed", "回饋方案資料無效")
		return
	}
	e := c.service.UpdateActivity(ctx.Request.Context(), ctx.Param("id"), r.serviceInput())
	if isNotFound(e) {
		failure(ctx, 404, "not_found", "找不到回饋方案")
		return
	}
	if e != nil {
		failure(ctx, 400, "validation_failed", "回饋方案資料無效或期間重疊")
		return
	}
	data(ctx, 200, gin.H{"updated": true})
}
func (c *Controller) DeleteActivity(ctx *gin.Context) {
	e := c.service.DeleteActivity(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(e) {
		failure(ctx, 404, "not_found", "找不到回饋方案")
		return
	}
	if e != nil {
		failure(ctx, http.StatusConflict, "activity_in_use", "回饋方案已有歷史交易，請停用")
		return
	}
	data(ctx, 200, gin.H{"deleted": true})
}
