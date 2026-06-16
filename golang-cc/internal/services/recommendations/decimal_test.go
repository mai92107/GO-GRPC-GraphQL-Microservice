package recommendations

import (
	"encoding/json"
	"testing"
)

func TestDecimalRoundHalfUp(t *testing.T) {
	tests := map[string]string{
		"1.004":  "1.00",
		"1.005":  "1.01",
		"1.006":  "1.01",
		"-1.005": "-1.01",
	}
	for input, expected := range tests {
		if actual := MustDecimal(input).Round(2).StringFixed(2); actual != expected {
			t.Errorf("Round(%s) = %s, want %s", input, actual, expected)
		}
	}
}

func TestDecimalJSONUsesString(t *testing.T) {
	value := MustDecimal("1.25")
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"1.250000"` {
		t.Fatalf("MarshalJSON = %s", data)
	}

	var decoded Decimal
	if err := json.Unmarshal([]byte(`"2.75"`), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.StringFixed(2) != "2.75" {
		t.Fatalf("decoded = %s", decoded.StringFixed(2))
	}
}
