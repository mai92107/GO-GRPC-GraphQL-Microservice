package member

import (
	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/domain"
)

type catalogResponse struct {
	ID           string                       `json:"id"`
	BankID       string                       `json:"bank_id"`
	BankName     string                       `json:"bank_name"`
	Name         string                       `json:"name"`
	IsActive     bool                         `json:"is_active"`
	AccountTiers []string                     `json:"account_tiers"`
	Activities   []domain.CardProductActivity `json:"activities"`
}
type unitResponse struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Symbol         string `json:"symbol"`
	SymbolPosition string `json:"symbol_position"`
	TWDRate        string `json:"twd_rate"`
	Precision      int    `json:"precision"`
}
type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
type paymentMethodResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (c *Controller) Catalog(ctx *gin.Context) {
	items, err := c.service.Catalog(ctx)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	out := []catalogResponse{}
	for _, x := range items {
		out = append(out, catalogResponse{ID: x.ID, BankID: x.BankID, BankName: x.BankName, Name: x.Name, IsActive: x.IsActive, AccountTiers: x.AccountTiers, Activities: x.Activities})
	}
	data(ctx, 200, out)
}
func (c *Controller) RewardUnits(ctx *gin.Context) {
	items, err := c.service.RewardUnits(ctx)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	out := []unitResponse{}
	for _, x := range items {
		out = append(out, unitResponse{ID: x.ID, Code: x.Code, Name: x.Name, Symbol: x.Symbol, SymbolPosition: x.SymbolPosition, TWDRate: x.TWDRate, Precision: x.Precision})
	}
	data(ctx, 200, out)
}
func (c *Controller) Categories(ctx *gin.Context) {
	items, err := c.service.Categories(ctx)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	out := []categoryResponse{}
	for _, x := range items {
		out = append(out, categoryResponse{Code: x.Code, Name: x.Name})
	}
	data(ctx, 200, out)
}
func (c *Controller) PaymentMethods(ctx *gin.Context) {
	items, err := c.service.PaymentMethods(ctx)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	out := make([]paymentMethodResponse, 0, len(items))
	for _, x := range items {
		out = append(out, paymentMethodResponse{Code: x.Code, Name: x.Name})
	}
	data(ctx, 200, out)
}

func (c *Controller) Merchants(ctx *gin.Context) {
	category := ctx.Query("category_code")
	if category == "" {
		failure(ctx, 400, "validation_failed", "消費類別為必填")
		return
	}
	items, err := c.service.Merchants(ctx, category)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	data(ctx, 200, items)
}
