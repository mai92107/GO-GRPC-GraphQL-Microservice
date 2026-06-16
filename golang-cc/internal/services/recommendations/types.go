package recommendations

import (
	"fmt"
	"time"
)

type ID string

type LocalDate struct {
	time.Time
}

func ParseLocalDate(value string) (LocalDate, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return LocalDate{}, fmt.Errorf("invalid local date %q: %w", value, err)
	}
	return LocalDate{Time: date}, nil
}

func MustLocalDate(value string) LocalDate {
	date, err := ParseLocalDate(value)
	if err != nil {
		panic(err)
	}
	return date
}

func (d LocalDate) String() string {
	return d.Format("2006-01-02")
}

func (d LocalDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *LocalDate) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("invalid local date")
	}
	value, err := ParseLocalDate(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*d = value
	return nil
}

type RewardUnit struct {
	ID             ID
	Code           string
	Name           string
	Symbol         string
	SymbolPosition string
	TWDRate        Decimal
	Precision      int
}

type Card struct {
	ID          ID
	UserID      ID
	Name        string
	AccountTier string
	IsActive    bool
}

type PaymentMethod struct {
	Code string
	Name string
}

type RewardRule struct {
	ID                   ID
	ActivityID           ID
	ActivityName         string
	UserID               ID
	CardID               ID
	Name                 string
	RewardUnit           RewardUnit
	Rate                 Decimal
	MonthlyCap           *Decimal
	StartDate            *LocalDate
	EndDate              *LocalDate
	IsActive             bool
	StackGroup           string
	Priority             int
	RequiredAccountTiers []string
	ActionRequired       string
	ActionMessage        string
	PaymentMethods       []string
	SharedMonthlyCap     *Decimal
	CategoryCode         []string
	MerchantCodes        []string
}

type RecommendationInput struct {
	UserID               ID
	AmountMinor          int64
	CategoryCode         string
	MerchantCode         string
	MerchantName         string
	PaymentMethods       []PaymentMethod
	Date                 LocalDate
	Cards                []Card
	Rules                []RewardRule
	Preferences          map[ID]Decimal
	MonthlyUsage         map[ID]Decimal
	ActivityMonthlyUsage map[string]Decimal
}

type RuleEvaluation struct {
	RuleID           ID
	RuleName         string
	ActivityID       ID
	ActivityName     string
	StackGroup       string
	ActionRequired   string
	ActionMessage    string
	RewardUnit       RewardUnit
	Rate             Decimal
	UncappedReward   Decimal
	MonthlyCap       *Decimal
	UsedBefore       Decimal
	RemainingBefore  *Decimal
	AllocatedReward  Decimal
	PreferenceWeight Decimal
	Score            Decimal
}

type CardRecommendation struct {
	Rank            int
	CardID          ID
	CardName        string
	TotalScore      Decimal
	TotalUnweighted Decimal
	Allocations     []RuleEvaluation
	Explanation     []string
	Reminders       []string
	PaymentOptions  []PaymentOption
}

type PaymentOption struct {
	PaymentMethodCode string
	PaymentMethodName string
	PaymentMethods    []PaymentMethod
	TotalScore        Decimal
	TotalUnweighted   Decimal
	Allocations       []RuleEvaluation
	Explanation       []string
	Reminders         []string
}

type Result struct {
	Recommendations []CardRecommendation
	EmptyReason     string
}
