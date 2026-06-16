package member

import (
	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/member"
)

type preferenceResponse struct {
	RewardUnitID string `json:"reward_unit_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Weight       string `json:"weight"`
}

func (c *Controller) Preferences(ctx *gin.Context) {
	items, err := c.service.Preferences(ctx, userID(ctx))
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	out := []preferenceResponse{}
	for _, x := range items {
		out = append(out, preferenceResponse{RewardUnitID: x.RewardUnitID, Code: x.Code, Name: x.Name, Symbol: x.Symbol, Weight: x.Weight})
	}
	data(ctx, 200, out)
}
func (c *Controller) UpdatePreferences(ctx *gin.Context) {
	var r struct {
		Preferences []struct {
			RewardUnitID string `json:"reward_unit_id" binding:"required"`
			Weight       string `json:"weight" binding:"required"`
		} `json:"preferences" binding:"required"`
	}
	if ctx.ShouldBindJSON(&r) != nil {
		failure(ctx, 400, "validation_failed", "輸入格式錯誤")
		return
	}
	items := []service.PreferenceInput{}
	for _, x := range r.Preferences {
		items = append(items, service.PreferenceInput{RewardUnitID: x.RewardUnitID, Weight: x.Weight})
	}
	if err := c.service.UpdatePreferences(ctx, userID(ctx), items); err != nil {
		failure(ctx, 400, "validation_failed", "權重或回饋單位無效")
		return
	}
	data(ctx, 200, gin.H{"updated": true})
}
