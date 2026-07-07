package recommendations

import "testing"

func TestEvaluateCapFormula(t *testing.T) {
	creditLimit := MustDecimal("100000")
	tests := []struct {
		name    string
		formula string
		want    string
	}{
		{name: "credit limit plus constant", formula: "member_card.credit_limit + 500000", want: "600000"},
		{name: "credit limit multiplier", formula: "member_card.credit_limit * 2", want: "200000"},
		{name: "parentheses and decimals", formula: "(member_card.credit_limit + 50000.5) / 2", want: "75000.25"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := evaluateCapFormula(test.formula, formulaVariables{MemberCardCreditLimit: &creditLimit})
			if err != nil {
				t.Fatal(err)
			}
			assertDecimal(t, got, test.want)
		})
	}
}

func TestEvaluateCapFormulaRejectsUnsafeInput(t *testing.T) {
	creditLimit := MustDecimal("100000")
	tests := []string{
		"member_card.available_limit + 1",
		"min(member_card.credit_limit, 1000)",
		"member_card.credit_limit / 0",
	}
	for _, formula := range tests {
		t.Run(formula, func(t *testing.T) {
			if _, err := evaluateCapFormula(formula, formulaVariables{MemberCardCreditLimit: &creditLimit}); err == nil {
				t.Fatal("expected formula error")
			}
		})
	}
}
