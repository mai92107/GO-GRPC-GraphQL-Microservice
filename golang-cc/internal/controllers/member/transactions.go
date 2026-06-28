package member

import (
	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/member"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

type transactionRequest struct {
	CardID          string `json:"card_id" binding:"required"`
	AmountMinor     int64  `json:"amount_minor" binding:"required,gt=0"`
	CategoryID      string `json:"category_id" binding:"required"`
	MerchantID      string `json:"merchant_id"`
	MerchantName    string `json:"merchant_name"`
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
	TransactionDate string `json:"transaction_date" binding:"required"`
	Note            string `json:"note"`
	Recommendation  []struct {
		RuleID          string `json:"rule_id"`
		AllocatedReward string `json:"allocated_reward"`
	} `json:"recommendation_summary"`
}

type transactionResponse struct {
	ID                string               `json:"id"`
	UserID            string               `json:"user_id"`
	CardID            string               `json:"card_id"`
	AmountMinor       int64                `json:"amount_minor"`
	CategoryID        string               `json:"category_id"`
	MerchantName      string               `json:"merchant_name"`
	MerchantID        string               `json:"merchant_id"`
	PaymentMethodID   string               `json:"payment_method_id"`
	PaymentMethodName string               `json:"payment_method_name"`
	TransactionDate   string               `json:"transaction_date"`
	Note              string               `json:"note"`
	Allocations       []allocationResponse `json:"allocations"`
}

type transactionSummaryResponse struct {
	ID                string `json:"id"`
	CardID            string `json:"card_id"`
	AmountMinor       int64  `json:"amount_minor"`
	CategoryID        string `json:"category_id"`
	MerchantName      string `json:"merchant_name"`
	MerchantID        string `json:"merchant_id"`
	PaymentMethodID   string `json:"payment_method_id"`
	PaymentMethodName string `json:"payment_method_name"`
	TransactionDate   string `json:"transaction_date"`
	Note              string `json:"note"`
}

func (c *Controller) CreateTransaction(ctx *gin.Context) {
	request, input, ok := parseTransaction(ctx)
	if !ok {
		return
	}
	result, err := c.service.CreateTransaction(ctx, userID(ctx), input)
	if err != nil {
		failure(ctx, 400, "validation_failed", err.Error())
		return
	}
	data(ctx, 201, gin.H{"transaction": mapTransaction(result), "recommendation_changed": recommendationChanged(request, result)})
}

func (c *Controller) UpdateTransaction(ctx *gin.Context) {
	_, input, ok := parseTransaction(ctx)
	if !ok {
		return
	}
	result, err := c.service.UpdateTransaction(ctx, userID(ctx), ctx.Param("id"), input)
	if service.IsTransactionNotFound(err) {
		failure(ctx, 404, "not_found", "找不到交易")
		return
	}
	if err != nil {
		failure(ctx, 400, "validation_failed", err.Error())
		return
	}
	data(ctx, 200, mapTransaction(result))
}

func (c *Controller) DeleteTransaction(ctx *gin.Context) {
	err := c.service.DeleteTransaction(ctx, userID(ctx), ctx.Param("id"))
	if service.IsTransactionNotFound(err) {
		failure(ctx, 404, "not_found", "找不到交易")
		return
	}
	if err != nil {
		failure(ctx, 500, "internal_error", "刪除失敗")
		return
	}
	data(ctx, 200, gin.H{"deleted": true})
}

func (c *Controller) ListTransactions(ctx *gin.Context) { c.transactions(ctx, "") }
func (c *Controller) GetTransaction(ctx *gin.Context)   { c.transactions(ctx, ctx.Param("id")) }

func (c *Controller) RewardCalculations(ctx *gin.Context) {
	items, err := c.service.RewardCalculations(ctx, userID(ctx), ctx.Param("id"))
	if notFound(err) {
		failure(ctx, 404, "not_found", "找不到交易")
		return
	}
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢回饋計算歷史失敗")
		return
	}
	data(ctx, 200, items)
}

func (c *Controller) transactions(ctx *gin.Context, id string) {
	items, err := c.service.ListTransactions(ctx, userID(ctx), id)
	if err != nil {
		failure(ctx, 500, "internal_error", "查詢失敗")
		return
	}
	output := make([]transactionSummaryResponse, 0, len(items))
	for _, item := range items {
		output = append(output, transactionSummaryResponse{
			ID: item.ID, CardID: item.CardID, AmountMinor: item.AmountMinor, CategoryID: item.CategoryID,
			MerchantID: item.MerchantID, MerchantName: item.MerchantName, PaymentMethodID: item.PaymentMethodID, PaymentMethodName: item.PaymentMethodName, TransactionDate: item.TransactionDate, Note: item.Note,
		})
	}
	if id != "" {
		if len(output) == 0 {
			failure(ctx, 404, "not_found", "找不到交易")
			return
		}
		data(ctx, 200, output[0])
		return
	}
	data(ctx, 200, output)
}

func parseTransaction(ctx *gin.Context) (transactionRequest, service.TransactionInput, bool) {
	var request transactionRequest
	if ctx.ShouldBindJSON(&request) != nil {
		failure(ctx, 400, "validation_failed", "輸入格式錯誤")
		return request, service.TransactionInput{}, false
	}
	date, err := recommendations.ParseLocalDate(request.TransactionDate)
	if err != nil {
		failure(ctx, 400, "validation_failed", "日期格式錯誤")
		return request, service.TransactionInput{}, false
	}
	return request, service.TransactionInput{
		CardID: request.CardID, AmountMinor: request.AmountMinor, CategoryID: request.CategoryID,
		MerchantID: request.MerchantID, MerchantName: request.MerchantName, PaymentMethodID: request.PaymentMethodID, TransactionDate: date, Note: request.Note,
	}, true
}

func mapTransaction(item service.Transaction) transactionResponse {
	allocations := make([]allocationResponse, 0, len(item.Allocations))
	for _, allocation := range item.Allocations {
		allocations = append(allocations, mapAllocation(allocation))
	}
	return transactionResponse{
		ID: item.ID, UserID: item.UserID, CardID: item.CardID, AmountMinor: item.AmountMinor,
		CategoryID: item.CategoryID, MerchantID: item.MerchantID, MerchantName: item.MerchantName, PaymentMethodID: item.PaymentMethodID, PaymentMethodName: item.PaymentMethodName,
		TransactionDate: item.TransactionDate.String(), Note: item.Note, Allocations: allocations,
	}
}

func recommendationChanged(request transactionRequest, transaction service.Transaction) bool {
	if len(request.Recommendation) == 0 {
		return false
	}
	actual := make(map[string]recommendations.Decimal, len(transaction.Allocations))
	for _, allocation := range transaction.Allocations {
		actual[string(allocation.RuleID)] = allocation.AllocatedReward
	}
	if len(actual) != len(request.Recommendation) {
		return true
	}
	for _, expected := range request.Recommendation {
		value, err := recommendations.ParseDecimal(expected.AllocatedReward)
		allocated, exists := actual[expected.RuleID]
		if err != nil || !exists || allocated.Cmp(value) != 0 {
			return true
		}
	}
	return false
}
