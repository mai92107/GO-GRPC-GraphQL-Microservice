package domain

import (
	"encoding/json"
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
	ID       string
	Name     string
	IsActive bool
}

type BankInput struct {
	Name       string
	WebsiteURL string
	IsActive   bool
}

type CardProduct struct {
	ID             string
	BankID         string
	BankName       string
	Name           string
	IsActive       bool
	AccountTiers   []string
	QualifiedType  string
	SelectableType string
	Networks       []CardNetwork
	Activities     []CardProductActivity
}

type CardProductActivity struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	StartDate  string            `json:"start_date"`
	EndDate    string            `json:"end_date"`
	IsActive   bool              `json:"is_active"`
	SourceURL  string            `json:"source_url"`
	VerifiedAt *string           `json:"verified_at"`
	NetworkIDs []string          `json:"network_ids"`
	Benefits   []ActivityBenefit `json:"benefits"`
}

type CardProductInput struct {
	BankID         string
	Name           string
	IsActive       bool
	AccountTiers   []string
	QualifiedType  string
	SelectableType string
	NetworkIDs     []string
}

type Category struct {
	ID       string
	Name     string
	IsActive bool
}

type PaymentMethod struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	IsSystem bool   `json:"is_system"`
}

type Merchant struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases"`
	CategoryIDs []string `json:"category_ids"`
	IsActive    bool     `json:"is_active"`
	IsSystem    bool     `json:"is_system"`
}

type RewardUnit struct {
	ID             string
	Name           string
	Symbol         string
	SymbolPosition string
	TWDRate        string
	Precision      int
}

type RewardUnitInput struct {
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
	NetworkIDs        []string
	SharedMonthlyCaps map[string]string
	Benefits          []ActivityBenefit
}

type ActivityBenefit struct {
	ID             string   `json:"id"`
	RewardUnitID   string   `json:"reward_unit_id"`
	Name           string   `json:"name"`
	DisplayOrder   int      `json:"display_order"`
	EffectType     string   `json:"effect_type"`
	RewardValue    string   `json:"reward_value"`
	MonthlyCap     *string  `json:"monthly_cap"`
	Layer          string   `json:"layer"`
	StackGroup     string   `json:"stack_group"`
	Priority       int      `json:"priority"`
	QualifiedType  string   `json:"qualified_type"`
	SelectableType string   `json:"selectable_type"`
	ActionRequired string   `json:"action_required"`
	ActionMessage  string   `json:"action_message"`
	PaymentMethods []string `json:"payment_methods"`
	CategoryIDs    []string `json:"category_ids"`
	MerchantIDs    []string `json:"merchant_ids"`
}

type RewardComponentVersionInput struct {
	RewardUnitID       string
	Name               string
	Description        string
	EffectType         string
	RewardValue        string
	EffectiveFrom      time.Time
	EffectiveTo        *time.Time
	AnnouncedAt        *time.Time
	ChangeReason       string
	DisplayChangeUntil *time.Time
}

type RewardConditionVersionInput struct {
	Operator      string
	Configuration json.RawMessage
	Description   string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

type RewardCapVersionInput struct {
	CapType       string
	LimitValue    string
	RewardUnitID  *string
	PeriodType    string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}
