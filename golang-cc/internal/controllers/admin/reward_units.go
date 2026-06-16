package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/admin"
)

type rewardUnitRequest struct {
	Code           string `json:"code" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Symbol         string `json:"symbol" binding:"required"`
	SymbolPosition string `json:"symbol_position" binding:"required,oneof=prefix suffix"`
	TWDRate        string `json:"twd_rate" binding:"required"`
	Precision      int    `json:"precision" binding:"min=0,max=6"`
}

type rewardUnitResponse struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Symbol         string `json:"symbol"`
	SymbolPosition string `json:"symbol_position"`
	TWDRate        string `json:"twd_rate"`
	Precision      int    `json:"precision"`
}

func (c *Controller) ListRewardUnits(ctx *gin.Context) {
	items, err := c.service.ListRewardUnits(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	output := make([]rewardUnitResponse, 0, len(items))
	for _, item := range items {
		output = append(output, rewardUnitResponse{
			ID: item.ID, Code: item.Code, Name: item.Name, Symbol: item.Symbol,
			SymbolPosition: item.SymbolPosition, TWDRate: item.TWDRate, Precision: item.Precision,
		})
	}
	data(ctx, http.StatusOK, output)
}

func (r rewardUnitRequest) serviceInput() service.RewardUnitInput {
	return service.RewardUnitInput{Code: r.Code, Name: r.Name, Symbol: r.Symbol, SymbolPosition: r.SymbolPosition, TWDRate: r.TWDRate, Precision: r.Precision}
}

func (c *Controller) CreateRewardUnit(ctx *gin.Context) {
	var input rewardUnitRequest
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	id, err := c.service.CreateRewardUnit(ctx.Request.Context(), input.serviceInput())
	if err != nil {
		failure(ctx, http.StatusConflict, "conflict", "回饋單位重複")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"id": id})
}
func (c *Controller) UpdateRewardUnit(ctx *gin.Context) {
	var input rewardUnitRequest
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	if err := c.service.UpdateRewardUnit(ctx.Request.Context(), ctx.Param("id"), input.serviceInput()); err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到單位")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}
func (c *Controller) DeleteRewardUnit(ctx *gin.Context) {
	err := c.service.DeleteRewardUnit(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到可刪除單位")
		return
	}
	if err != nil {
		failure(ctx, http.StatusConflict, "unit_in_use", "回饋單位使用中")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}
