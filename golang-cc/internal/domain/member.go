package domain

type MemberCard struct {
	ID            string
	CardProductID string
	Name          string
	Issuer        string
	LastFour      string
	IsActive      bool
	StatementDay  *int
	PaymentDueDay *int
	AccountTier   string
}

type CatalogCard struct {
	ID           string
	BankID       string
	BankName     string
	Name         string
	IsActive     bool
	AccountTiers []string
	Activities   []CardProductActivity
}

type MemberRewardUnit struct {
	ID             string
	Code           string
	Name           string
	Symbol         string
	SymbolPosition string
	TWDRate        string
	Precision      int
}

type MemberCategory struct {
	Code string
	Name string
}

type MemberPaymentMethod struct {
	Code string
	Name string
}

type MemberMerchant struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type RewardPreference struct {
	RewardUnitID string
	Code         string
	Name         string
	Symbol       string
	Weight       string
}

type TransactionSummary struct {
	ID                string
	CardID            string
	AmountMinor       int64
	CategoryCode      string
	MerchantName      string
	MerchantCode      string
	PaymentMethodCode string
	PaymentMethodName string
	TransactionDate   string
	Note              string
}
