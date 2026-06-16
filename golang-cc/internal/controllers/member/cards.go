package member

import (
	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/member"
)

type cardRequest struct {
	CardProductID string `json:"card_product_id"`
	Nickname      string `json:"nickname"`
	LastFour      string `json:"last_four"`
	IsActive      *bool  `json:"is_active"`
	StatementDay  *int   `json:"statement_day"`
	PaymentDueDay *int   `json:"payment_due_day"`
	AccountTier   string `json:"account_tier"`
}
type cardResponse struct {
	ID            string `json:"id"`
	CardProductID string `json:"card_product_id"`
	Name          string `json:"name"`
	Issuer        string `json:"issuer"`
	LastFour      string `json:"last_four"`
	IsActive      bool   `json:"is_active"`
	StatementDay  *int   `json:"statement_day"`
	PaymentDueDay *int   `json:"payment_due_day"`
	AccountTier   string `json:"account_tier"`
}

func (c *Controller) ListCards(ctx *gin.Context) {
	items, err := c.service.ListCards(ctx, userID(ctx))
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢卡片失敗")
		return
	}
	out := make([]cardResponse, 0, len(items))
	for _, x := range items {
		out = append(out, cardResponse{ID: x.ID, CardProductID: x.CardProductID, Name: x.Name, Issuer: x.Issuer, LastFour: x.LastFour, IsActive: x.IsActive, StatementDay: x.StatementDay, PaymentDueDay: x.PaymentDueDay, AccountTier: x.AccountTier})
	}
	data(ctx, 200, out)
}
func (c *Controller) GetCard(ctx *gin.Context) {
	x, err := c.service.GetCard(ctx, userID(ctx), ctx.Param("id"))
	if err != nil {
		failure(ctx, 404, "not_found", "找不到卡片")
		return
	}
	data(ctx, 200, cardResponse{ID: x.ID, CardProductID: x.CardProductID, Name: x.Name, Issuer: x.Issuer, LastFour: x.LastFour, IsActive: x.IsActive, StatementDay: x.StatementDay, PaymentDueDay: x.PaymentDueDay, AccountTier: x.AccountTier})
}
func (c *Controller) CreateCard(ctx *gin.Context) {
	var r cardRequest
	if ctx.ShouldBindJSON(&r) != nil || r.CardProductID == "" {
		failure(ctx, 400, "validation_failed", "請選擇卡片目錄中的卡片")
		return
	}
	active := true
	if r.IsActive != nil {
		active = *r.IsActive
	}
	id, err := c.service.CreateCard(ctx, userID(ctx), service.CardInput{CardProductID: r.CardProductID, Nickname: r.Nickname, LastFour: r.LastFour, IsActive: active, StatementDay: r.StatementDay, PaymentDueDay: r.PaymentDueDay, AccountTier: r.AccountTier})
	if err != nil {
		failure(ctx, 409, "card_conflict", "卡片已在卡片夾中或資料無效")
		return
	}
	data(ctx, 201, gin.H{"id": id, "card_product_id": r.CardProductID})
}
func (c *Controller) UpdateCard(ctx *gin.Context) {
	var r cardRequest
	if ctx.ShouldBindJSON(&r) != nil || r.IsActive == nil {
		failure(ctx, 400, "validation_failed", "PATCH 需提供完整卡片資料")
		return
	}
	err := c.service.UpdateCard(ctx, userID(ctx), ctx.Param("id"), service.CardInput{Nickname: r.Nickname, LastFour: r.LastFour, IsActive: *r.IsActive, StatementDay: r.StatementDay, PaymentDueDay: r.PaymentDueDay, AccountTier: r.AccountTier})
	if err != nil {
		failure(ctx, 404, "not_found", "找不到卡片或資料無效")
		return
	}
	data(ctx, 200, gin.H{"updated": true})
}
func (c *Controller) DeleteCard(ctx *gin.Context) {
	err := c.service.DeleteCard(ctx, userID(ctx), ctx.Param("id"))
	if notFound(err) {
		failure(ctx, 404, "not_found", "找不到卡片")
		return
	}
	if err != nil {
		failure(ctx, 409, "card_in_use", "卡片已有交易，請改為停用")
		return
	}
	data(ctx, 200, gin.H{"deleted": true})
}
