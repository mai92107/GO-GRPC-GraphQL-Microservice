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
	Name           string
	Symbol         string
	SymbolPosition string
	TWDRate        Decimal
	Precision      int
}

type Card struct {
	ID                   ID
	UserID               ID
	Name                 string
	AccountTier          string
	NetworkID            ID
	QualifiedCardPlanIDs []ID
	IsActive             bool
}

type PaymentMethod struct {
	ID   string
	Name string
}

type EffectType string

const (
	EffectAddRate      EffectType = "ADD_RATE"
	EffectSetRate      EffectType = "SET_RATE"
	EffectMultiplyRate EffectType = "MULTIPLY_RATE"
	EffectAddCash      EffectType = "ADD_CASH"
	EffectDiscount     EffectType = "DISCOUNT"
)

type ExcludeReason string

const (
	ExcludeStackGroupLost ExcludeReason = "STACK_GROUP_LOST"
	ExcludeCapReached     ExcludeReason = "CAP_REACHED"
)

type RewardRule struct {
	ID                   ID
	ActivityID           ID
	ActivityName         string
	UserID               ID
	CardID               ID
	Name                 string
	RewardUnit           RewardUnit
	MonthlyCap           *Decimal
	StartDate            *LocalDate
	EndDate              *LocalDate
	IsActive             bool
	StackGroup           string
	Priority             int
	Layer                string
	DisplayOrder         int
	EffectType           EffectType
	RewardValue          Decimal
	QualifiedType        string
	ActionRequired       string
	ActionMessage        string
	PaymentMethods       []string
	SharedMonthlyCap     *Decimal
	CategoryID           []string
	MerchantIDs          []string
	CardNetworkIDs       []ID
	QualifiedCardPlanIDs []ID
	SuggestedCardPlanID  ID
	SuggestedPlanName    string
}

type RecommendationInput struct {
	UserID               ID
	AmountMinor          int64
	CategoryID           string
	MerchantID           string
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
	RuleID              ID
	RuleName            string
	ActivityID          ID
	ActivityName        string
	StackGroup          string
	StackPolicy         string
	Layer               string
	DisplayOrder        int
	EffectType          EffectType
	RewardValue         Decimal
	ActionRequired      string
	ActionMessage       string
	RewardUnit          RewardUnit
	RewardRate          Decimal
	UncappedReward      Decimal
	MonthlyCap          *Decimal
	UsedBefore          Decimal
	RemainingBefore     *Decimal
	AllocatedReward     Decimal
	PreferenceWeight    Decimal
	Score               Decimal
	SuggestedCardPlanID ID
	SuggestedPlanName   string
}

type BenefitResult struct {
	BenefitID       ID
	Name            string
	ActivityID      ID
	ActivityName    string
	EffectType      EffectType
	RewardValue     Decimal
	RewardRate      Decimal
	RewardAmount    Decimal
	StackGroup      string
	ExclusiveGroup  string
	MonthlyCap      *Decimal
	RemainingBefore *Decimal
	IsApplied       bool
	ExcludeReason   ExcludeReason
}

type LayerResult struct {
	Layer        string
	RewardRate   Decimal
	RewardAmount Decimal
	Benefits     []BenefitResult
}

type ExcludedBenefit struct {
	BenefitID    ID
	Name         string
	ActivityID   ID
	ActivityName string
	Reason       ExcludeReason
	StackGroup   string
	Layer        string
	LayerName    string
}

type CardRecommendation struct {
	Rank             int
	CardID           ID
	CardName         string
	TotalScore       Decimal
	TotalUnweighted  Decimal
	Layers           []LayerResult
	ExcludedBenefits []ExcludedBenefit
	Allocations      []RuleEvaluation
	Explanation      []string
	Reminders        []string
	PaymentOptions   []PaymentOption
}

type PaymentOption struct {
	PaymentMethodID   string
	PaymentMethodName string
	PaymentMethods    []PaymentMethod
	TotalScore        Decimal
	TotalUnweighted   Decimal
	Allocations       []RuleEvaluation
	Layers            []LayerResult
	ExcludedBenefits  []ExcludedBenefit
	Explanation       []string
	Reminders         []string
}

type Result struct {
	Recommendations []CardRecommendation
	EmptyReason     string
}
