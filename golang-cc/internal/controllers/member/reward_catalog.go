package member

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (c *Controller) RewardOverview(ctx *gin.Context) {
	at := time.Now()
	if value := ctx.Query("at"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			failure(ctx, 400, "validation_failed", "at 必須是 RFC3339 時間")
			return
		}
		at = parsed
	}
	item, err := c.service.RewardOverview(ctx, userID(ctx), ctx.Param("id"), at)
	if notFound(err) {
		failure(ctx, 404, "not_found", "找不到卡片")
		return
	}
	if err != nil {
		println("RewardOverview error:", err.Error())
		failure(ctx, 500, "internal_error", "查詢卡片回饋失敗")
		return
	}
	data(ctx, 200, item)
}

func (c *Controller) SetQualificationStatus(ctx *gin.Context) {
	var request struct {
		IsQualified   *bool  `json:"is_qualified"`
		EffectiveFrom string `json:"effective_from"`
	}
	if ctx.ShouldBindJSON(&request) != nil || request.IsQualified == nil {
		failure(ctx, 400, "validation_failed", "輸入格式錯誤")
		return
	}
	effectiveFrom, err := time.Parse(time.RFC3339, request.EffectiveFrom)
	if err != nil {
		failure(ctx, 400, "validation_failed", "effective_from 必須是 RFC3339 時間")
		return
	}
	err = c.service.SetQualificationStatus(ctx, userID(ctx), ctx.Param("id"), ctx.Param("planId"), *request.IsQualified, effectiveFrom)
	if notFound(err) {
		failure(ctx, 404, "not_found", "找不到卡片或資格方案")
		return
	}
	if err != nil {
		failure(ctx, 409, "qualification_conflict", "資格期間與既有資料衝突")
		return
	}
	data(ctx, 200, gin.H{"updated": true})
}

func (c *Controller) UpdatePaymentMethods(ctx *gin.Context) {
	var request struct {
		PaymentMethodIDs []string `json:"payment_method_ids"`
	}
	if ctx.ShouldBindJSON(&request) != nil {
		failure(ctx, 400, "validation_failed", "輸入格式錯誤")
		return
	}
	codes := request.PaymentMethodIDs

	if err := c.service.UpdatePaymentMethods(ctx, userID(ctx), codes); err != nil {
		println("UpdatePaymentMethods error:", err.Error())
		failure(ctx, 400, "validation_failed", "支付方式資料無效")
		return
	}
	data(ctx, 200, gin.H{"updated": true})
}
