package admin

import "encoding/json"

type BankInput struct {
	Name       string
	WebsiteURL string
	IsActive   bool
}

type CardProductInput struct {
	BankID         string
	Name           string
	IsActive       bool
	QualifiedType  string
	SelectableType string
	Networks       []string
}

type RewardUnitInput struct {
	Name           string
	Symbol         string
	SymbolPosition string
	TWDRate        string
	Precision      int
}

type ActivityInput struct {
	CardProductID     string
	Name              string
	StartDate         string
	EndDate           string
	IsActive          *bool
	SourceURL         string
	VerifiedAt        *string
	NetworkIDs        []string
	SharedMonthlyCaps map[string]string
	Benefits          []ActivityBenefitInput
}

type ActivityBenefitInput struct {
	ID             string
	RewardUnitID   string
	Name           string
	Layer          string
	DisplayOrder   int
	EffectType     string
	RewardValue    string
	MonthlyCap     *string
	StackGroup     string
	StackPolicy    string
	Priority       int
	QualifiedType  string
	SelectableType string
	ActionRequired string
	ActionMessage  string
	PaymentMethods []string
	CategoryIDs    []string
	MerchantIds    []string
}

type ActivityFilters struct {
	BankID        string
	CardProductID string
	IsActive      *bool
}

type RequirementOptionFilters struct {
	RequirementType string
	CardProductID   string
}

type ActivityFlowInput struct {
	Activity     ActivitySummaryInput
	RewardGroups []RewardGroupInput
}

type ActivitySummaryInput struct {
	BankID        string
	CardProductID string
	Title         string
	Description   string
	SourceURL     string
	EffectiveFrom string
	EffectiveTo   string
	IsActive      bool
}

type RewardGroupInput struct {
	ID           string
	Name         string
	Description  string
	DisplayOrder int
	IsActive     bool
	Components   []RewardComponentInput
}

type RewardComponentInput struct {
	ID            string
	Name          string
	Description   string
	Layer         int
	StackGroup    string
	StackMode     string
	Priority      int
	EffectiveFrom string
	EffectiveTo   string
	IsActive      bool
	Requirements  []RewardRequirementInput
	Benefits      []RewardBenefitInput
}

type RewardRequirementInput struct {
	ID              string
	RequirementType string
	Operator        string
	Configuration   json.RawMessage
	Description     string
	IsActive        bool
}

type RewardBenefitInput struct {
	ID           string
	BenefitType  string
	Value        string
	RewardUnitID string
	CapAmount    *string
	CapPeriod    *string
	Description  string
	IsActive     bool
}
