package recommendations

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// Decimal is an exact base-10 value used by the recommendation domain.
type Decimal struct {
	value *big.Rat
}

func ParseDecimal(s string) (Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Decimal{}, fmt.Errorf("decimal is empty")
	}
	value, ok := new(big.Rat).SetString(s)
	if !ok {
		return Decimal{}, fmt.Errorf("invalid decimal %q", s)
	}
	return Decimal{value: value}, nil
}

func MustDecimal(s string) Decimal {
	value, err := ParseDecimal(s)
	if err != nil {
		panic(err)
	}
	return value
}

func DecimalFromInt(value int64) Decimal {
	return Decimal{value: new(big.Rat).SetInt64(value)}
}

func (d Decimal) rat() *big.Rat {
	if d.value == nil {
		return new(big.Rat)
	}
	return new(big.Rat).Set(d.value)
}

func (d Decimal) Add(other Decimal) Decimal {
	return Decimal{value: new(big.Rat).Add(d.rat(), other.rat())}
}

func (d Decimal) Sub(other Decimal) Decimal {
	return Decimal{value: new(big.Rat).Sub(d.rat(), other.rat())}
}

func (d Decimal) Mul(other Decimal) Decimal {
	return Decimal{value: new(big.Rat).Mul(d.rat(), other.rat())}
}

func (d Decimal) Div(other Decimal) Decimal {
	if other.Sign() == 0 {
		panic("division by zero")
	}
	return Decimal{value: new(big.Rat).Quo(d.rat(), other.rat())}
}

func (d Decimal) Cmp(other Decimal) int {
	return d.rat().Cmp(other.rat())
}

func (d Decimal) Sign() int {
	return d.rat().Sign()
}

func (d Decimal) Min(other Decimal) Decimal {
	if d.Cmp(other) <= 0 {
		return d
	}
	return other
}

func (d Decimal) Max(other Decimal) Decimal {
	if d.Cmp(other) >= 0 {
		return d
	}
	return other
}

// Round applies decimal half-up rounding at the requested precision.
func (d Decimal) Round(precision int) Decimal {
	if precision < 0 {
		panic("negative decimal precision")
	}

	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil)
	scaledNumerator := new(big.Int).Mul(d.rat().Num(), scale)
	denominator := d.rat().Denom()
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(scaledNumerator, denominator, remainder)

	absoluteRemainder := new(big.Int).Abs(remainder)
	twiceRemainder := new(big.Int).Mul(absoluteRemainder, big.NewInt(2))
	if twiceRemainder.Cmp(denominator) >= 0 {
		if d.Sign() >= 0 {
			quotient.Add(quotient, big.NewInt(1))
		} else {
			quotient.Sub(quotient, big.NewInt(1))
		}
	}

	return Decimal{value: new(big.Rat).SetFrac(quotient, scale)}
}

func (d Decimal) String() string {
	return d.rat().FloatString(6)
}

func (d Decimal) StringFixed(precision int) string {
	return d.rat().FloatString(precision)
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decimal must be a JSON string: %w", err)
	}
	value, err := ParseDecimal(raw)
	if err != nil {
		return err
	}
	*d = value
	return nil
}
