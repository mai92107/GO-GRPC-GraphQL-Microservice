package domain

import "encoding/json"

type MemberCard struct {
	MemberCardId   string `json:"member_card_id"`
	Name           string `json:"name"`
	Issuer         string `json:"issuer"`
	LastFour       string `json:"last_four"`
	CreditLimit    string `json:"credit_limit"`
	IsActive       bool   `json:"is_active"`
	CardImageURL   string `json:"card_image_url"`
	PrimaryColor   string `json:"primary_color"`
	QualifiedType  string `json:"qualified_type"`
	SelectableType string `json:"selectable_type"`
	Network        string `json:"network"`
}

type MemberCardInfo struct {
	MemberCardId   string `json:"memberCardID"`
	Name           string `json:"name"`
	Nickname       string `json:"nickname"`
	Issuer         string `json:"issuer"`
	LastFour       string `json:"last_four"`
	IsActive       bool   `json:"is_active"`
	StatementDay   *int   `json:"statement_day"`
	PaymentDueDay  *int   `json:"payment_due_day"`
	AccountTier    string `json:"account_tier"`
	CreditLimit    string `json:"credit_limit"`
	CardImageURL   string `json:"card_image_url"`
	PrimaryColor   string `json:"primary_color"`
	QualifiedType  string `json:"qualified_type"`
	SelectableType string `json:"selectable_type"`
	Network        string `json:"network"`
}

type CatalogCard struct {
	ID             string
	BankID         string
	BankName       string
	Name           string
	CardImageURL   string
	PrimaryColor   string
	IsActive       bool
	AccountTiers   []string
	QualifiedType  string
	SelectableType string
	Activities     []CardProductActivity
	Networks       []string
}

type MemberRewardUnit struct {
	ID             string
	Name           string
	Symbol         string
	SymbolPosition string
	TWDRate        string
	Precision      int
}

type MemberCategory struct {
	ID   string
	Name string
}

type MemberPaymentMethod struct {
	ID          string
	Name        string
	Type        string
	IsAvailable bool
}

type QualificationStatus struct {
	PlanID        string `json:"plan_id"`
	Name          string `json:"name"`
	IsQualified   bool   `json:"is_qualified"`
	EffectiveFrom string `json:"effective_from"`
}

type RewardCapOverview struct {
	Type      string `json:"type"`
	Limit     string `json:"limit"`
	Period    string `json:"period"`
	Spendable string `json:"spendable"`
}

type RewardGroupOverview struct {
	ComponentID                 string             `json:"component_id"`
	Name                        string             `json:"name"`
	CurrentRate                 string             `json:"current_rate"`
	Layer                       string             `json:"layer"`
	DisplayOrder                int                `json:"display_order"`
	EffectType                  string             `json:"effect_type"`
	RewardValue                 string             `json:"reward_value"`
	PreviousRate                *string            `json:"previous_rate"`
	NextRate                    *string            `json:"next_rate"`
	ChangeEffectiveAt           *string            `json:"change_effective_at"`
	ShowPreviousAsStrikethrough bool               `json:"show_previous_as_strikethrough"`
	Requirements                []string           `json:"requirements"`
	Reminders                   []string           `json:"reminders"`
	Cap                         *RewardCapOverview `json:"cap"`
}

type MemberCardRewardOverview struct {
	Card           MemberCard            `json:"card"`
	QualifiedPlans []QualificationStatus `json:"qualified_plans"`
	RewardGroups   []RewardGroupOverview `json:"reward_groups"`
}

type RewardCalculationComponentSnapshot struct {
	ComponentID           *string         `json:"component_id"`
	ComponentVersionID    *string         `json:"component_version_id"`
	RewardUnitID          string          `json:"reward_unit_id"`
	SuggestedCardPlanID   *string         `json:"suggested_card_plan_id"`
	QualifiedCardPlanID   *string         `json:"qualified_card_plan_id"`
	QualificationStatusID *string         `json:"qualification_status_id"`
	EffectType            string          `json:"effect_type"`
	RewardValue           string          `json:"reward_value"`
	RewardRate            string          `json:"reward_rate"`
	UncappedReward        string          `json:"uncapped_reward"`
	AllocatedReward       string          `json:"allocated_reward"`
	CapUsedBefore         string          `json:"cap_used_before"`
	CapRemainingBefore    *string         `json:"cap_remaining_before"`
	ConditionSnapshot     json.RawMessage `json:"condition_snapshot"`
	ReminderSnapshot      json.RawMessage `json:"reminder_snapshot"`
}

type RewardCalculationSnapshot struct {
	ID                 string                               `json:"id"`
	CalculationNo      int                                  `json:"calculation_no"`
	CalculationType    string                               `json:"calculation_type"`
	TransactionAt      string                               `json:"transaction_at"`
	TotalEffectiveRate string                               `json:"total_effective_rate"`
	TotalRewardValue   string                               `json:"total_reward_value"`
	EngineVersion      string                               `json:"engine_version"`
	Status             string                               `json:"status"`
	InputSnapshot      json.RawMessage                      `json:"input_snapshot"`
	CalculatedAt       string                               `json:"calculated_at"`
	Components         []RewardCalculationComponentSnapshot `json:"components"`
}

type MemberMerchant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RewardPreference struct {
	RewardUnitID string
	Name         string
	Symbol       string
	Weight       string
}

type TransactionSummary struct {
	ID                string
	CardID            string
	AmountMinor       int64
	CategoryID        string
	MerchantName      string
	MerchantID        string
	PaymentMethodID   string
	PaymentMethodName string
	TransactionDate   string
	Note              string
}
