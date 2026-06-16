package admin

import (
	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/admin"
	"net/http"
)

type benefitRequest struct {
	ID                   string   `json:"id"`
	RewardUnitID         string   `json:"reward_unit_id" binding:"required"`
	Name                 string   `json:"name" binding:"required"`
	Rate                 string   `json:"rate" binding:"required"`
	MonthlyCap           *string  `json:"monthly_cap"`
	StackGroup           string   `json:"stack_group" binding:"required"`
	Priority             int      `json:"priority"`
	RequiredAccountTiers []string `json:"required_account_tiers"`
	ActionRequired       string   `json:"action_required"`
	ActionMessage        string   `json:"action_message"`
	PaymentMethods       []string `json:"payment_methods"`
	CategoryCodes        []string `json:"category_codes" binding:"required,min=1"`
	MerchantCodes        []string `json:"merchant_codes"`
}
type activityRequest struct {
	CardProductID     string            `json:"card_product_id" binding:"required"`
	Name              string            `json:"name" binding:"required"`
	StartDate         string            `json:"start_date" binding:"required"`
	EndDate           string            `json:"end_date" binding:"required"`
	IsActive          *bool             `json:"is_active"`
	SourceURL         string            `json:"source_url"`
	VerifiedAt        *string           `json:"verified_at"`
	SharedMonthlyCaps map[string]string `json:"shared_monthly_caps"`
	Benefits          []benefitRequest  `json:"benefits" binding:"required,min=1"`
}

func (r activityRequest) serviceInput() service.ActivityInput {
	bs := make([]service.ActivityBenefitInput, 0, len(r.Benefits))
	for _, b := range r.Benefits {
		bs = append(bs, service.ActivityBenefitInput{ID: b.ID, RewardUnitID: b.RewardUnitID, Name: b.Name, Rate: b.Rate, MonthlyCap: b.MonthlyCap, StackGroup: b.StackGroup, Priority: b.Priority, RequiredAccountTiers: b.RequiredAccountTiers, ActionRequired: b.ActionRequired, ActionMessage: b.ActionMessage, PaymentMethods: b.PaymentMethods, CategoryCodes: b.CategoryCodes, MerchantCodes: b.MerchantCodes})
	}
	return service.ActivityInput{CardProductID: r.CardProductID, Name: r.Name, StartDate: r.StartDate, EndDate: r.EndDate, IsActive: r.IsActive, SourceURL: r.SourceURL, VerifiedAt: r.VerifiedAt, SharedMonthlyCaps: r.SharedMonthlyCaps, Benefits: bs}
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
		failure(ctx, 400, "validation_failed", "活動資料無效")
		return
	}
	id, e := c.service.CreateActivity(ctx.Request.Context(), r.serviceInput())
	if e != nil {
		failure(ctx, 400, "validation_failed", "活動資料無效或期間重疊")
		return
	}
	data(ctx, 201, gin.H{"id": id})
}
func (c *Controller) UpdateActivity(ctx *gin.Context) {
	var r activityRequest
	if ctx.ShouldBindJSON(&r) != nil || r.IsActive == nil {
		failure(ctx, 400, "validation_failed", "活動資料無效")
		return
	}
	e := c.service.UpdateActivity(ctx.Request.Context(), ctx.Param("id"), r.serviceInput())
	if isNotFound(e) {
		failure(ctx, 404, "not_found", "找不到活動")
		return
	}
	if e != nil {
		failure(ctx, 400, "validation_failed", "活動資料無效或期間重疊")
		return
	}
	data(ctx, 200, gin.H{"updated": true})
}
func (c *Controller) DeleteActivity(ctx *gin.Context) {
	e := c.service.DeleteActivity(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(e) {
		failure(ctx, 404, "not_found", "找不到活動")
		return
	}
	if e != nil {
		failure(ctx, http.StatusConflict, "activity_in_use", "活動已有歷史交易，請停用")
		return
	}
	data(ctx, 200, gin.H{"deleted": true})
}
