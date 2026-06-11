package metric

import "testing"

func TestAllowedMetricNames(t *testing.T) {
	if !ProcessCPUUsage.Valid() {
		t.Fatal("expected process CPU metric to be valid")
	}
	if Name("process_cpu_usag").Valid() {
		t.Fatal("expected typo metric to be invalid")
	}
	if !JVMThreadsStatesThreads.Valid() || !JVMThreadsDaemonThreads.Valid() || !JVMThreadsPeakThreads.Valid() || !JVMThreadsStartedThreadsTotal.Valid() {
		t.Fatal("expected expanded JVM thread metrics to be valid")
	}
	if !SystemCPUUsage.Valid() || !SystemCPUCount.Valid() {
		t.Fatal("expected system CPU metrics to be valid")
	}
	if len(AllowedValues()) != 18 {
		t.Fatalf("expected 18 allowed metrics, got %d", len(AllowedValues()))
	}
}
