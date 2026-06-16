package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/admin"
)

type cardProductRequest struct {
	BankID       string   `json:"bank_id" binding:"required"`
	Name         string   `json:"name" binding:"required"`
	IsActive     *bool    `json:"is_active"`
	AccountTiers []string `json:"account_tiers"`
}

func (c *Controller) ListCardProducts(ctx *gin.Context) {
	result, err := c.service.ListCardProducts(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapCardProducts(result))
}
func (c *Controller) CreateCardProduct(ctx *gin.Context) {
	var input cardProductRequest
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	id, err := c.service.CreateCardProduct(ctx.Request.Context(), service.CardProductInput{BankID: input.BankID, Name: input.Name, IsActive: active, AccountTiers: input.AccountTiers})
	if err != nil {
		failure(ctx, http.StatusConflict, "conflict", "卡片重複或資料無效")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"id": id})
}
func (c *Controller) UpdateCardProduct(ctx *gin.Context) {
	var input cardProductRequest
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	err := c.service.UpdateCardProduct(ctx.Request.Context(), ctx.Param("id"), service.CardProductInput{BankID: input.BankID, Name: input.Name, IsActive: *input.IsActive, AccountTiers: input.AccountTiers})
	if err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到卡片")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}
func (c *Controller) DeleteCardProduct(ctx *gin.Context) {
	err := c.service.DeleteCardProduct(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到卡片")
		return
	}
	if err != nil {
		failure(ctx, http.StatusConflict, "card_in_use", "卡片已有會員或活動，請停用")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}
