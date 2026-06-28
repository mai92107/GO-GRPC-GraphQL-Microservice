package admin

type BankInput struct {
	Name       string
	WebsiteURL string
	IsActive   bool
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
