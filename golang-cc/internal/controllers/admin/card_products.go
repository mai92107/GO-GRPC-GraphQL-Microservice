package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/admin"
)

type cardProductRequest struct {
	BankID         string   `json:"bank_id" binding:"required"`
	Name           string   `json:"name" binding:"required"`
	IsActive       *bool    `json:"is_active"`
	QualifiedType  string   `json:"qualified_type"`
	SelectableType string   `json:"selectable_type"`
	Networks       []string `json:"networks" binding:"required,min=1"`
}

func (c *Controller) ListCards(ctx *gin.Context) {
	result, err := c.service.ListCards(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapCards(result))
}

func (c *Controller) GetCardProduct(ctx *gin.Context) {
	result, err := c.service.GetCardInfo(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		println("取得卡片資訊失敗:", err.Error())
		failure(ctx, http.StatusNotFound, "not_found", "找不到卡片")
		return
	}
	data(ctx, http.StatusOK, mapCardProduct(result))
}

func (c *Controller) ListNetworks(ctx *gin.Context) {
	result, err := c.service.ListNetworks(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, result)
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
	id, err := c.service.CreateCard(ctx.Request.Context(), service.CardProductInput{BankID: input.BankID, Name: input.Name, IsActive: active, QualifiedType: input.QualifiedType, SelectableType: input.SelectableType, Networks: input.Networks})
	if err != nil {
		println("error creating card, error: " + err.Error())
		failure(ctx, http.StatusConflict, "conflict", "卡片重複或資料無效")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"id": id})
}
func (c *Controller) UpdateCard(ctx *gin.Context) {
	var input cardProductRequest
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	err := c.service.UpdateCard(ctx.Request.Context(), ctx.Param("id"), service.CardProductInput{BankID: input.BankID, Name: input.Name, IsActive: *input.IsActive, QualifiedType: input.QualifiedType, SelectableType: input.SelectableType, Networks: input.Networks})
	if err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到卡片")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}
func (c *Controller) DeleteCard(ctx *gin.Context) {
	err := c.service.DeleteCard(ctx.Request.Context(), ctx.Param("id"))
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
