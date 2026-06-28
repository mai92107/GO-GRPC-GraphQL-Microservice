package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type merchantRequest struct {
	Name        string   `json:"name" binding:"required"`
	Aliases     []string `json:"aliases"`
	CategoryIDs []string `json:"category_ids" binding:"required,min=1"`
	IsActive    *bool    `json:"is_active"`
}

func (c *Controller) ListMerchants(ctx *gin.Context) {
	items, err := c.service.ListMerchants(ctx)
	if err != nil {
		println("Error listing merchants:", err.Error())
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	data(ctx, 200, items)
}

func (c *Controller) CreateMerchant(ctx *gin.Context) {
	var input merchantRequest
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, 400, "validation_failed", "店家資料無效")
		return
	}
	id, err := c.service.CreateMerchant(ctx, input.Name, input.Aliases, input.CategoryIDs)
	if err != nil {
		println("Error creating merchant:", err.Error())
		failure(ctx, 400, "validation_failed", "店家資料無效")
		return
	}
	data(ctx, 201, gin.H{"id": id})
}

func (c *Controller) UpdateMerchant(ctx *gin.Context) {
	var input merchantRequest
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, 400, "validation_failed", "店家資料無效")
		return
	}
	if err := c.service.UpdateMerchant(ctx, ctx.Param("id"), input.Name, input.Aliases, input.CategoryIDs, *input.IsActive); err != nil {
		failure(ctx, 400, "validation_failed", "店家資料無效")
		return
	}
	data(ctx, 200, gin.H{"updated": true})
}

func (c *Controller) DeleteMerchant(ctx *gin.Context) {
	if err := c.service.DeleteMerchant(ctx, ctx.Param("id")); err != nil {
		failure(ctx, http.StatusConflict, "merchant_in_use", "店家使用中，請停用")
		return
	}
	data(ctx, 200, gin.H{"deleted": true})
}
