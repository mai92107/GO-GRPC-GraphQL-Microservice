package domain

import (
	"time"
)

type Dashboard struct {
	Members          int
	Banks            int
	Cards            int
	ActiveActivities int
}

type AdminUser struct {
	ID          string
	Email       string
	DisplayName string
	Role        string
	Status      string
	CreatedAt   time.Time
}

type TelegramBinding struct {
	ChatID      int64
	UserID      string
	Email       string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Invitation struct {
	ID         string
	Email      string
	ExpiresAt  time.Time
	AcceptedAt *time.Time
	CreatedAt  time.Time
}

type Bank struct {
	ID         string
	Name       string
	Code       string
	WebsiteURL string
	IsActive   bool
}

type BankInput struct {
	Name       string
	Code       string
	WebsiteURL string
	IsActive   bool
}

type CardProduct struct {
	ID           string
	BankID       string
	BankName     string
	Name         string
	IsActive     bool
	AccountTiers []string
	Activities   []CardProductActivity
}

type CardProductActivity struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	StartDate  string            `json:"start_date"`
	EndDate    string            `json:"end_date"`
	IsActive   bool              `json:"is_active"`
	SourceURL  string            `json:"source_url"`
	VerifiedAt *string           `json:"verified_at"`
	Benefits   []ActivityBenefit `json:"benefits"`
}

type CardProductInput struct {
	BankID       string
	Name         string
	IsActive     bool
	AccountTiers []string
}

type Category struct {
	Code     string
	Name     string
	IsActive bool
}

type PaymentMethod struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	IsSystem bool   `json:"is_system"`
}

type Merchant struct {
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	Aliases       []string `json:"aliases"`
	CategoryCodes []string `json:"category_codes"`
	IsActive      bool     `json:"is_active"`
	IsSystem      bool     `json:"is_system"`
}

type RewardUnit struct {
	ID             string
	Code           string
	Name           string
	Symbol         string
	SymbolPosition string
	TWDRate        string
	Precision      int
}

type RewardUnitInput struct {
	Code           string
	Name           string
	Symbol         string
	SymbolPosition string
	TWDRate        string
	Precision      int
}

type Activity struct {
	ID                string
	CardProductID     string
	Name              string
	StartDate         string
	EndDate           string
	IsActive          bool
	SourceURL         string
	VerifiedAt        *string
	SharedMonthlyCaps map[string]string
	Benefits          []ActivityBenefit
}

type ActivityBenefit struct {
	ID                   string   `json:"id"`
	RewardUnitID         string   `json:"reward_unit_id"`
	Name                 string   `json:"name"`
	Rate                 string   `json:"rate"`
	MonthlyCap           *string  `json:"monthly_cap"`
	StackGroup           string   `json:"stack_group"`
	Priority             int      `json:"priority"`
	RequiredAccountTiers []string `json:"required_account_tiers"`
	ActionRequired       string   `json:"action_required"`
	ActionMessage        string   `json:"action_message"`
	PaymentMethods       []string `json:"payment_methods"`
	CategoryCodes        []string `json:"category_codes"`
	MerchantCodes        []string `json:"merchant_codes"`
}
