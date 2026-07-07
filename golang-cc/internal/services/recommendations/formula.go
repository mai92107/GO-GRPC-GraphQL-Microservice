package recommendations

import (
	"fmt"
	"strings"
	"unicode"
)

const memberCardCreditLimitVariable = "member_card.credit_limit"

type formulaVariables struct {
	MemberCardCreditLimit *Decimal
}

func evaluateCapFormula(raw string, variables formulaVariables) (Decimal, error) {
	value, err := evaluateCapFormulaValue(raw, variables)
	if err != nil {
		return Decimal{}, err
	}
	if value.Sign() <= 0 {
		return Decimal{}, fmt.Errorf("formula result must be positive")
	}
	return value, nil
}

func evaluateCapFormulaValue(raw string, variables formulaVariables) (Decimal, error) {
	parser := capFormulaParser{input: strings.TrimSpace(raw), variables: variables}
	if parser.input == "" {
		return Decimal{}, fmt.Errorf("formula is empty")
	}
	value, err := parser.parseExpression()
	if err != nil {
		return Decimal{}, err
	}
	parser.skipSpaces()
	if !parser.atEnd() {
		return Decimal{}, fmt.Errorf("unexpected token %q", parser.remaining())
	}
	return value, nil
}

func ValidateCapFormula(raw string) error {
	_, err := evaluateCapFormulaValue(raw, formulaVariables{
		MemberCardCreditLimit: decimalPointerForFormula(MustDecimal("1")),
	})
	return err
}

func decimalPointerForFormula(value Decimal) *Decimal {
	return &value
}

type capFormulaParser struct {
	input     string
	position  int
	variables formulaVariables
}

func (p *capFormulaParser) parseExpression() (Decimal, error) {
	left, err := p.parseTerm()
	if err != nil {
		return Decimal{}, err
	}
	for {
		p.skipSpaces()
		switch p.peek() {
		case '+':
			p.position++
			right, err := p.parseTerm()
			if err != nil {
				return Decimal{}, err
			}
			left = left.Add(right)
		case '-':
			p.position++
			right, err := p.parseTerm()
			if err != nil {
				return Decimal{}, err
			}
			left = left.Sub(right)
		default:
			return left, nil
		}
	}
}

func (p *capFormulaParser) parseTerm() (Decimal, error) {
	left, err := p.parseFactor()
	if err != nil {
		return Decimal{}, err
	}
	for {
		p.skipSpaces()
		switch p.peek() {
		case '*':
			p.position++
			right, err := p.parseFactor()
			if err != nil {
				return Decimal{}, err
			}
			left = left.Mul(right)
		case '/':
			p.position++
			right, err := p.parseFactor()
			if err != nil {
				return Decimal{}, err
			}
			if right.Sign() == 0 {
				return Decimal{}, fmt.Errorf("division by zero")
			}
			left = left.Div(right)
		default:
			return left, nil
		}
	}
}

func (p *capFormulaParser) parseFactor() (Decimal, error) {
	p.skipSpaces()
	switch {
	case p.atEnd():
		return Decimal{}, fmt.Errorf("unexpected end of formula")
	case p.peek() == '(':
		p.position++
		value, err := p.parseExpression()
		if err != nil {
			return Decimal{}, err
		}
		p.skipSpaces()
		if p.peek() != ')' {
			return Decimal{}, fmt.Errorf("missing closing parenthesis")
		}
		p.position++
		return value, nil
	case p.peek() == '-':
		p.position++
		value, err := p.parseFactor()
		if err != nil {
			return Decimal{}, err
		}
		return Decimal{}.Sub(value), nil
	case isFormulaDigit(p.peek()) || p.peek() == '.':
		return p.parseNumber()
	case isIdentifierStart(p.peek()):
		return p.parseVariable()
	default:
		return Decimal{}, fmt.Errorf("unexpected character %q", p.peek())
	}
}

func (p *capFormulaParser) parseNumber() (Decimal, error) {
	start := p.position
	dotCount := 0
	for !p.atEnd() {
		next := p.peek()
		if next == '.' {
			dotCount++
			if dotCount > 1 {
				return Decimal{}, fmt.Errorf("invalid decimal")
			}
			p.position++
			continue
		}
		if !isFormulaDigit(next) {
			break
		}
		p.position++
	}
	value, err := ParseDecimal(p.input[start:p.position])
	if err != nil {
		return Decimal{}, err
	}
	return value, nil
}

func (p *capFormulaParser) parseVariable() (Decimal, error) {
	start := p.position
	for !p.atEnd() {
		next := p.peek()
		if !isIdentifierPart(next) {
			break
		}
		p.position++
	}
	name := p.input[start:p.position]
	if name != memberCardCreditLimitVariable {
		return Decimal{}, fmt.Errorf("unknown variable %q", name)
	}
	if p.variables.MemberCardCreditLimit == nil {
		return Decimal{}, fmt.Errorf("member card credit limit is missing")
	}
	return *p.variables.MemberCardCreditLimit, nil
}

func (p *capFormulaParser) skipSpaces() {
	for !p.atEnd() && unicode.IsSpace(rune(p.peek())) {
		p.position++
	}
}

func (p capFormulaParser) atEnd() bool {
	return p.position >= len(p.input)
}

func (p capFormulaParser) peek() byte {
	if p.atEnd() {
		return 0
	}
	return p.input[p.position]
}

func (p capFormulaParser) remaining() string {
	if p.atEnd() {
		return ""
	}
	return p.input[p.position:]
}

func isFormulaDigit(value byte) bool {
	return value >= '0' && value <= '9'
}

func isIdentifierStart(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z') || value == '_'
}

func isIdentifierPart(value byte) bool {
	return isIdentifierStart(value) || isFormulaDigit(value) || value == '.'
}
