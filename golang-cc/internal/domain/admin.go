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

type Users struct {
	User
	PasswordHash string    `gorm:"column:password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdateddAt   time.Time `gorm:"column:updated_at"`
}

func (Users) TableName() string {
	return "identity.users"
}

type Session struct {
	Id         string    `gorm:"column:id"`
	UserID     string    `gorm:"column:user_id"`
	TokenHash  []byte    `gorm:"column:token_hash"`
	ExpiresAt  time.Time `gorm:"column:expires_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	LastSeenAt time.Time `gorm:"column:last_seen_at"`
}

func (Session) TableName() string {
	return "identity.sessions"
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
	ID         string     `gorm:"column:id;primaryKey"`
	Email      string     `gorm:"column:email"`
	AcceptedAt *time.Time `gorm:"column:accepted_at"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (Invitation) TableName() string {
	return "invitations"
}

type PasswordResetToken struct {
	ID        string    `gorm:"column:id;primaryKey"`
	UserID    string    `gorm:"column:user_id"`
	TokenHash []byte    `gorm:"column:token_hash"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	UsedAt    time.Time `gorm:"column:used_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (PasswordResetToken) TableName() string {
	return "identity.password_reset_tokens"
}

type Bank struct {
	ID         string    `gorm:"column:id;primaryKey"`
	Name       string    `gorm:"column:name"`
	WebsiteURL *string   `gorm:"column:website_url"`
	IsActive   bool      `gorm:"column:is_active"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdateddAt time.Time `gorm:"column:updated_at"`
}

func (Bank) TableName() string {
	return "banks"
}

type BankInput struct {
	Name       string
	WebsiteURL string
	IsActive   bool
}

type CardInput struct {
	BankID         string
	Name           string
	IsActive       bool
	QualifiedType  string
	SelectableType string
	Networks       []string
}

type Card struct {
	ID             string `gorm:"column:id;primaryKey"`
	BankID         string `gorm:"column:bank_id"`
	BankName       string `gorm:"column:bank_name"`
	Name           string `gorm:"column:name"`
	IsActive       bool   `gorm:"column:is_active"`
	QualifiedType  string `gorm:"column:qualified_type"`
	SelectableType string `gorm:"column:selectable_type"`
	Networks       string `gorm:"column:networks"`
}

type CardInfo struct {
	ID             string `gorm:"column:id;primaryKey"`
	BankID         string `gorm:"column:bank_id"`
	BankName       string `gorm:"column:bank_name"`
	Name           string `gorm:"column:name"`
	IsActive       bool   `gorm:"column:is_active"`
	QualifiedType  string `gorm:"column:qualified_type"`
	SelectableType string `gorm:"column:selectable_type"`
	Networks       string `gorm:"column:networks"`
}

type CardProduct struct {
	ID             string  `gorm:"column:id;primaryKey"`
	BankID         string  `gorm:"column:bank_id"`
	Name           string  `gorm:"column:name"`
	IsActive       bool    `gorm:"column:is_active"`
	Description    string  `gorm:"column:description"`
	QualifiedType  *string `gorm:"column:qualified_type"`
	SelectableType *string `gorm:"column:selectable_type"`
	Networks       string  `gorm:"column:networks"`
}

func (CardProduct) TableName() string {
	return "catalog.card_products"
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

type Category struct {
	ID       string
	Name     string
	IsActive bool
}

type LookupItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
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
	LimitFormula  string
	RewardUnitID  *string
	PeriodType    string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

type ActivitySummary struct {
	ID             string
	BankID         string
	BankName       string
	CardProductID  string
	CardName       string
	Title          string
	Description    string
	SourceURL      string
	EffectiveFrom  string
	EffectiveTo    string
	IsActive       bool
	GroupCount     int64
	ComponentCount int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ActivityFlow struct {
	Activity     ActivitySummary
	RewardGroups []RewardGroup
}

type RewardGroup struct {
	ID           string
	ActivityID   string
	Name         string
	Description  string
	DisplayOrder int
	IsActive     bool
	Components   []RewardComponent
}

type RewardComponent struct {
	ID             string
	RewardGroupIDs []string
	Name           string
	Description    string
	Layer          int
	StackGroup     string
	StackMode      string
	Priority       int
	EffectiveFrom  string
	EffectiveTo    string
	IsActive       bool
	Requirements   []RewardRequirement
	Benefits       []RewardBenefit
}

type RewardRequirement struct {
	ID                string
	RewardComponentID string
	RequirementType   string
	Operator          string
	Configuration     json.RawMessage
	Description       string
	IsActive          bool
}

type RewardBenefit struct {
	ID                string
	RewardComponentID string
	BenefitType       string
	Value             string
	RewardUnitID      string
	CapAmount         *string
	CapFormula        *string
	CapPeriod         *string
	Description       string
	IsActive          bool
}

type RequirementOptionSet struct {
	RequirementTypes   []RequirementTypeOption
	Operators          []CodeNameOption
	PaymentMethods     []CodeNameOption
	Merchants          []CodeNameOption
	Categories         []CodeNameOption
	CardNetworks       []CodeNameOption
	CardProducts       []CodeNameOption
	CardPlans          []CardPlanOption
	AccountTiers       []CodeNameOption
	UserQualifications []CodeNameOption
	Channels           []CodeNameOption
	Regions            []CodeNameOption
	Currencies         []CodeNameOption
}

type RequirementTypeOption struct {
	Code        string
	Name        string
	ValueKey    string
	ValueSource string
}

type CodeNameOption struct {
	Code string
	Name string
}

type CardPlanOption struct {
	ID            string
	CardProductID string
	PlanType      string
	Name          string
}
