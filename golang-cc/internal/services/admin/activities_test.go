package admin

import (
	"errors"
	"testing"

	"github.com/rafa/golang-cc/internal/domain"
)

func validActivityInput() ActivityInput {
	return ActivityInput{
		CardProductID: "card", Name: " 2026 上半年 ", StartDate: "2026-01-01", EndDate: "2026-06-30",
		Benefits: []ActivityBenefitInput{{
			RewardUnitID: "unit", Name: " 全聯活動 ", Rate: "0.1", StackGroup: " bonus ",
			CategoryCodes: []string{"dining"}, MerchantCodes: []string{" px_mart ", "px_mart_en"},
			PaymentMethods: []string{"apple_pay"},
		}},
	}
}

func TestPrepareActivityNormalizesNestedBenefits(t *testing.T) {
	activity, err := prepareActivity("activity", validActivityInput(), true)
	if err != nil {
		t.Fatal(err)
	}
	b := activity.Benefits[0]
	if activity.Name != "2026 上半年" || b.Name != "全聯活動" || b.StackGroup != "bonus" || b.MerchantCodes[0] != "px_mart" || len(b.PaymentMethods) != 1 {
		t.Fatalf("activity not normalized: %+v", activity)
	}
}

func TestPrepareActivityRejectsDuplicateMerchantOrInvalidPeriod(t *testing.T) {
	input := validActivityInput()
	input.Benefits[0].MerchantCodes = []string{"px_mart", "px_mart"}
	if _, err := prepareActivity("activity", input, true); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error=%v", err)
	}
	input = validActivityInput()
	input.EndDate = "2025-12-31"
	if _, err := prepareActivity("activity", input, true); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error=%v", err)
	}
	input = validActivityInput()
	input.Benefits[0].PaymentMethods = []string{"apple_pay", "apple_pay"}
	if _, err := prepareActivity("activity", input, true); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("duplicate payment error=%v", err)
	}
}

func TestPrepareActivityAcceptsAccountSetupReminder(t *testing.T) {
	input := validActivityInput()
	input.Benefits[0].ActionRequired = "account_setup"
	input.Benefits[0].ActionMessage = "需完成自動扣繳"
	if _, err := prepareActivity("activity", input, true); err != nil {
		t.Fatal(err)
	}
}
