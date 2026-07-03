package admin

import (
	"errors"
	"testing"

	"github.com/rafa/golang-cc/internal/domain"
)

func validActivityFlowInput() ActivityFlowInput {
	return ActivityFlowInput{
		Activity: ActivitySummaryInput{
			BankID: "bank", CardProductID: "card", Title: " 2026 Q3 ",
			EffectiveFrom: "2026-07-01", EffectiveTo: "2026-09-30", IsActive: true,
		},
		RewardGroups: []RewardGroupInput{{
			Name: " LINE Pay ", DisplayOrder: 10, IsActive: true,
			Components: []RewardComponentInput{{
				Name: " 基本回饋 ", Layer: 1, StackGroup: " PAYMENT ", StackMode: "ADDITIVE",
				EffectiveFrom: "2026-07-01", EffectiveTo: "2026-09-30", IsActive: true,
				Requirements: []RewardRequirementInput{{
					RequirementType: "PAYMENT_METHOD", Operator: "IN",
					Configuration: []byte(`{"payment_method_codes":["LINE_PAY"]}`), Description: "限 LINE Pay", IsActive: true,
				}},
				Benefits: []RewardBenefitInput{{
					BenefitType: "RATE_CASHBACK", Value: "3", RewardUnitID: "unit", CapAmount: stringPtr("300"), CapPeriod: stringPtr("MONTHLY"), IsActive: true,
				}, {
					BenefitType: "POINT", Value: "1", RewardUnitID: "point", IsActive: true,
				}},
			}},
		}},
	}
}

func TestPrepareActivityFlowNormalizesAndAcceptsMultipleBenefits(t *testing.T) {
	flow, err := prepareActivityFlow("activity", validActivityFlowInput())
	if err != nil {
		t.Fatal(err)
	}
	if flow.Activity.Title != "2026 Q3" || flow.RewardGroups[0].Name != "LINE Pay" {
		t.Fatalf("flow not normalized: %+v", flow)
	}
	component := flow.RewardGroups[0].Components[0]
	if component.StackMode != "ADDITIVE" || len(component.Benefits) != 2 || !component.Requirements[0].IsActive {
		t.Fatalf("unexpected component: %+v", component)
	}
}

func TestPrepareActivityFlowRejectsInvalidRequirementConfig(t *testing.T) {
	input := validActivityFlowInput()
	input.RewardGroups[0].Components[0].Requirements[0].Configuration = []byte(`{"network_codes":["VISA"]}`)
	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error=%v", err)
	}
}

func TestPrepareActivityFlowRejectsInvalidStackModeOrDecimal(t *testing.T) {
	input := validActivityFlowInput()
	input.RewardGroups[0].Components[0].StackMode = "BEST"
	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("stack mode error=%v", err)
	}
	input = validActivityFlowInput()
	input.RewardGroups[0].Components[0].Benefits[0].Value = "abc"
	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("decimal error=%v", err)
	}
}

func stringPtr(value string) *string { return &value }
