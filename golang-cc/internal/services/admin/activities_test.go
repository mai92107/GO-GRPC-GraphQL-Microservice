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
			ID: "group-a", Name: " LINE Pay ", DisplayOrder: 10, IsActive: true,
			Components: []RewardComponentInput{{
				ID: "component-a", RewardGroupIDs: []string{"group-a"}, Name: " 基本回饋 ", Layer: 1, StackGroup: " PAYMENT ", StackMode: "ADDITIVE",
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

func TestPrepareActivityFlowAcceptsSharedComponentGroupIDs(t *testing.T) {
	input := validActivityFlowInput()
	input.RewardGroups = append(input.RewardGroups, RewardGroupInput{
		ID: "group-b", Name: "街口支付", DisplayOrder: 20, IsActive: true,
		Components: []RewardComponentInput{input.RewardGroups[0].Components[0]},
	})
	input.RewardGroups[0].Components[0].RewardGroupIDs = []string{"group-a", "group-b"}
	input.RewardGroups[1].Components[0].RewardGroupIDs = []string{"group-a", "group-b"}

	flow, err := prepareActivityFlow("activity", input)
	if err != nil {
		t.Fatal(err)
	}
	first := flow.RewardGroups[0].Components[0]
	second := flow.RewardGroups[1].Components[0]
	if first.ID != second.ID || len(first.RewardGroupIDs) != 2 || len(second.RewardGroupIDs) != 2 {
		t.Fatalf("component was not shared: first=%+v second=%+v", first, second)
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
	} else if err.Error() != "Group[1].Component[1].Requirement[1].configuration_json 與 requirement_type 不相符" {
		t.Fatalf("message=%q", err.Error())
	}
}

func TestPrepareActivityFlowAcceptsInstallmentRequirement(t *testing.T) {
	input := validActivityFlowInput()
	input.RewardGroups[0].Components[0].Requirements[0].RequirementType = "INSTALLMENT"
	input.RewardGroups[0].Components[0].Requirements[0].Operator = "EQ"
	input.RewardGroups[0].Components[0].Requirements[0].Configuration = []byte(`{"is_installment":true}`)

	flow, err := prepareActivityFlow("activity", input)
	if err != nil {
		t.Fatal(err)
	}
	requirement := flow.RewardGroups[0].Components[0].Requirements[0]
	if requirement.RequirementType != "INSTALLMENT" || string(requirement.Configuration) != `{"is_installment":true}` {
		t.Fatalf("unexpected requirement: %+v", requirement)
	}
}

func TestPrepareActivityFlowRejectsInstallmentRequirementWithoutFlag(t *testing.T) {
	input := validActivityFlowInput()
	input.RewardGroups[0].Components[0].Requirements[0].RequirementType = "INSTALLMENT"
	input.RewardGroups[0].Components[0].Requirements[0].Operator = "EQ"
	input.RewardGroups[0].Components[0].Requirements[0].Configuration = []byte(`{"values":[true]}`)

	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error=%v", err)
	}
}

func TestPrepareActivityFlowRejectsInvalidStackModeOrDecimal(t *testing.T) {
	input := validActivityFlowInput()
	input.RewardGroups[0].Components[0].StackMode = "BEST"
	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("stack mode error=%v", err)
	} else if err.Error() != "Group[1].Component[1].stack_mode 無效" {
		t.Fatalf("stack mode message=%q", err.Error())
	}
	input = validActivityFlowInput()
	input.RewardGroups[0].Components[0].Benefits[0].Value = "abc"
	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("decimal error=%v", err)
	} else if err.Error() != "Group[1].Component[1].Benefit[1].value 必須是數字" {
		t.Fatalf("decimal message=%q", err.Error())
	}
}

func TestPrepareActivityFlowRejectsNonYYYYMMDDDates(t *testing.T) {
	input := validActivityFlowInput()
	input.Activity.EffectiveFrom = "2026-07-01T00:00:00Z"
	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("activity date error=%v", err)
	}

	input = validActivityFlowInput()
	input.RewardGroups[0].Components[0].EffectiveTo = "2026/09/30"
	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("component date error=%v", err)
	} else if err.Error() != "Group[1].Component[1].effective_to 必須是 YYYY-MM-DD" {
		t.Fatalf("component date message=%q", err.Error())
	}
}

func TestPrepareActivityFlowRejectsMissingBenefitUnitWithFieldMessage(t *testing.T) {
	input := validActivityFlowInput()
	input.RewardGroups[0].Components[0].Benefits[0].RewardUnitID = ""

	if _, err := prepareActivityFlow("activity", input); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("benefit unit error=%v", err)
	} else if err.Error() != "Group[1].Component[1].Benefit[1].reward_unit_id 必填" {
		t.Fatalf("benefit unit message=%q", err.Error())
	}
}

func stringPtr(value string) *string { return &value }
