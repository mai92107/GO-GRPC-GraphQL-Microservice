package parser

import "testing"

func TestParsePrometheusKeepsOnlySupportedMetrics(t *testing.T) {
	input := `
# HELP jvm_memory_used_bytes Used bytes
jvm_memory_used_bytes{area="heap",id="G1 Eden Space"} 128
jvm_memory_max_bytes{area="heap",id="G1 Eden Space"} 256
unrelated_metric 99
process_cpu_usage 0.42
`

	snapshots, err := ParsePrometheus("order-service", input)
	if err != nil {
		t.Fatalf("parse prometheus: %v", err)
	}

	if len(snapshots) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(snapshots))
	}
	if snapshots[0].ServiceName != "order-service" {
		t.Fatalf("unexpected service name: %q", snapshots[0].ServiceName)
	}
	if snapshots[0].Labels["area"] != "heap" {
		t.Fatalf("expected heap label, got %#v", snapshots[0].Labels)
	}
}
