package admin

type BankInput struct {
	Name       string
	Code       string
	WebsiteURL string
	IsActive   bool
}

type CardProductInput struct {
	BankID       string
	Name         string
	IsActive     bool
	AccountTiers []string
}

type RewardUnitInput struct {
	Code           string
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
	SharedMonthlyCaps map[string]string
	Benefits          []ActivityBenefitInput
}

type ActivityBenefitInput struct {
	ID                   string
	RewardUnitID         string
	Name                 string
	Rate                 string
	MonthlyCap           *string
	StackGroup           string
	Priority             int
	RequiredAccountTiers []string
	ActionRequired       string
	ActionMessage        string
	PaymentMethods       []string
	CategoryCodes        []string
	MerchantCodes        []string
}
