package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ListPaymentMethods(ctx *gin.Context) {
	items, err := c.service.ListPaymentMethods(ctx)
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, items)
}

func (c *Controller) CreatePaymentMethod(ctx *gin.Context) {
	var input struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	if ctx.ShouldBindJSON(&input) != nil || c.service.CreatePaymentMethod(ctx, input.Code, input.Name) != nil {
		failure(ctx, http.StatusConflict, "conflict", "支付方式資料無效或重複")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"code": input.Code})
}

func (c *Controller) UpdatePaymentMethod(ctx *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		IsActive *bool  `json:"is_active" binding:"required"`
	}
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	if err := c.service.UpdatePaymentMethod(ctx, ctx.Param("code"), input.Name, *input.IsActive); err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到支付方式")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}

func (c *Controller) DeletePaymentMethod(ctx *gin.Context) {
	err := c.service.DeletePaymentMethod(ctx, ctx.Param("code"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到可刪除支付方式")
		return
	}
	if err != nil {
		failure(ctx, http.StatusConflict, "payment_method_in_use", "支付方式使用中，請停用")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}
