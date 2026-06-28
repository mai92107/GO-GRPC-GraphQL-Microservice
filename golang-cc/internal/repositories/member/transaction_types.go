package member

import (
	"fmt"

	"github.com/rafa/golang-cc/internal/services/recommendations"
)

type Transaction struct {
	ID                recommendations.ID
	UserID            recommendations.ID
	CardID            recommendations.ID
	AmountMinor       int64
	CategoryID        string
	MerchantName      string
	MerchantID        string
	PaymentMethodID   string
	PaymentMethodName string
	TransactionDate   recommendations.LocalDate
	Note              string
	Allocations       []recommendations.RuleEvaluation
}

type WriteInput struct {
	CardID          recommendations.ID
	AmountMinor     int64
	CategoryID      string
	MerchantName    string
	MerchantID      string
	PaymentMethodID string
	TransactionDate recommendations.LocalDate
	Note            string
}

var ErrNotFound = fmt.Errorf("transaction not found")
