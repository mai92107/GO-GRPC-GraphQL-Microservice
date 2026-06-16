package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/admin"
)

type bankRequest struct {
	Name       string `json:"name" binding:"required"`
	Code       string `json:"code"`
	WebsiteURL string `json:"website_url"`
	IsActive   *bool  `json:"is_active"`
}

func (c *Controller) ListBanks(ctx *gin.Context) {
	result, err := c.service.ListBanks(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapBanks(result))
}
func (c *Controller) CreateBank(ctx *gin.Context) {
	var input bankRequest
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "銀行名稱必填")
		return
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	id, err := c.service.CreateBank(ctx.Request.Context(), service.BankInput{Name: input.Name, Code: input.Code, WebsiteURL: input.WebsiteURL, IsActive: active})
	if err != nil {
		failure(ctx, http.StatusConflict, "conflict", "銀行重複或資料無效")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"id": id})
}
func (c *Controller) UpdateBank(ctx *gin.Context) {
	var input bankRequest
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	err := c.service.UpdateBank(ctx.Request.Context(), ctx.Param("id"), service.BankInput{Name: input.Name, Code: input.Code, WebsiteURL: input.WebsiteURL, IsActive: *input.IsActive})
	if err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到銀行")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}
func (c *Controller) DeleteBank(ctx *gin.Context) {
	err := c.service.DeleteBank(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到銀行")
		return
	}
	if err != nil {
		failure(ctx, http.StatusConflict, "bank_in_use", "銀行已有卡片，請停用")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}
