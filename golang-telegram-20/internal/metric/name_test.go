package metric

import "testing"

func TestAllowedMetricNames(t *testing.T) {
	if !ProcessCPUUsage.Valid() {
		t.Fatal("expected process CPU metric to be valid")
	}
	if Name("process_cpu_usag").Valid() {
		t.Fatal("expected typo metric to be invalid")
	}
	if len(AllowedValues()) != 12 {
		t.Fatalf("expected 12 allowed metrics, got %d", len(AllowedValues()))
	}
}
