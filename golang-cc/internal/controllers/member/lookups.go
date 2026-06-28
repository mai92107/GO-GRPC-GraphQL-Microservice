package member

import (
	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/domain"
)

type catalogResponse struct {
	ID             string                       `json:"id"`
	BankID         string                       `json:"bank_id"`
	BankName       string                       `json:"bank_name"`
	Name           string                       `json:"name"`
	CardImageURL   string                       `json:"card_image_url"`
	PrimaryColor   string                       `json:"primary_color"`
	IsActive       bool                         `json:"is_active"`
	AccountTiers   []string                     `json:"account_tiers"`
	QualifiedType  string                       `json:"qualified_type"`
	SelectableType string                       `json:"selectable_type"`
	Activities     []domain.CardProductActivity `json:"activities"`
	Networks       []domain.CardNetwork         `json:"networks"`
}
type unitResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Symbol         string `json:"symbol"`
	SymbolPosition string `json:"symbol_position"`
	TWDRate        string `json:"twd_rate"`
	Precision      int    `json:"precision"`
}
type categoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type paymentMethodResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	IsAvailable bool   `json:"is_available"`
}

func (c *Controller) Catalog(ctx *gin.Context) {
	items, err := c.service.Catalog(ctx)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	out := []catalogResponse{}
	for _, x := range items {
		out = append(out, catalogResponse{ID: x.ID, BankID: x.BankID, BankName: x.BankName, Name: x.Name, CardImageURL: x.CardImageURL, PrimaryColor: x.PrimaryColor, IsActive: x.IsActive, AccountTiers: x.AccountTiers, QualifiedType: x.QualifiedType, SelectableType: x.SelectableType, Activities: x.Activities, Networks: x.Networks})
	}
	data(ctx, 200, out)
}

func (c *Controller) CatalogCard(ctx *gin.Context) {
	x, err := c.service.CatalogCard(ctx, ctx.Param("id"))
	if err != nil {
		failure(ctx, 404, "not_found", "找不到卡片")
		return
	}
	data(ctx, 200, catalogResponse{ID: x.ID, BankID: x.BankID, BankName: x.BankName, Name: x.Name, CardImageURL: x.CardImageURL, PrimaryColor: x.PrimaryColor, IsActive: x.IsActive, AccountTiers: x.AccountTiers, QualifiedType: x.QualifiedType, SelectableType: x.SelectableType, Activities: x.Activities, Networks: x.Networks})
}

func (c *Controller) RewardUnits(ctx *gin.Context) {
	items, err := c.service.RewardUnits(ctx)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	out := []unitResponse{}
	for _, x := range items {
		out = append(out, unitResponse{ID: x.ID, Name: x.Name, Symbol: x.Symbol, SymbolPosition: x.SymbolPosition, TWDRate: x.TWDRate, Precision: x.Precision})
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
		out = append(out, categoryResponse{ID: x.ID, Name: x.Name})
	}
	data(ctx, 200, out)
}
func (c *Controller) PaymentMethods(ctx *gin.Context) {
	items, err := c.service.UserPaymentMethods(ctx, userID(ctx))
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	out := make([]paymentMethodResponse, 0, len(items))
	for _, x := range items {
		out = append(out, paymentMethodResponse{ID: x.ID, Name: x.Name, Type: x.Type, IsAvailable: x.IsAvailable})
	}
	data(ctx, 200, out)
}

func (c *Controller) Merchants(ctx *gin.Context) {
	category := ctx.Query("category_id")
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
