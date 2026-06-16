package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type merchantRequest struct {
	Code          string   `json:"code"`
	Name          string   `json:"name" binding:"required"`
	Aliases       []string `json:"aliases"`
	CategoryCodes []string `json:"category_codes" binding:"required,min=1"`
	IsActive      *bool    `json:"is_active"`
}

func (c *Controller) ListMerchants(ctx *gin.Context) {
	items, err := c.service.ListMerchants(ctx)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	data(ctx, 200, items)
}

func (c *Controller) CreateMerchant(ctx *gin.Context) {
	var input merchantRequest
	if ctx.ShouldBindJSON(&input) != nil || c.service.CreateMerchant(ctx, input.Code, input.Name, input.Aliases, input.CategoryCodes) != nil {
		failure(ctx, 400, "validation_failed", "店家資料無效")
		return
	}
	data(ctx, 201, gin.H{"code": input.Code})
}

func (c *Controller) UpdateMerchant(ctx *gin.Context) {
	var input merchantRequest
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, 400, "validation_failed", "店家資料無效")
		return
	}
	if err := c.service.UpdateMerchant(ctx, ctx.Param("code"), input.Name, input.Aliases, input.CategoryCodes, *input.IsActive); err != nil {
		failure(ctx, 400, "validation_failed", "店家資料無效")
		return
	}
	data(ctx, 200, gin.H{"updated": true})
}

func (c *Controller) DeleteMerchant(ctx *gin.Context) {
	if err := c.service.DeleteMerchant(ctx, ctx.Param("code")); err != nil {
		failure(ctx, http.StatusConflict, "merchant_in_use", "店家使用中，請停用")
		return
	}
	data(ctx, 200, gin.H{"deleted": true})
}
