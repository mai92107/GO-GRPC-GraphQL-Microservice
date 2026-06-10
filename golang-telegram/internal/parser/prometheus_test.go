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

func TestParsePrometheusAcceptsSpringBootTrailingLabelComma(t *testing.T) {
	input := `
jvm_memory_used_bytes{area="heap",id="G1 Old Gen",} 6.6771404E8
hikaricp_connections_pending{pool="HikariPool-1",} 0.0
process_cpu_usage 0.004802195289275097
`

	snapshots, err := ParsePrometheus("order-service", input)
	if err != nil {
		t.Fatalf("parse prometheus: %v", err)
	}

	if len(snapshots) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(snapshots))
	}
	if snapshots[0].Labels["id"] != "G1 Old Gen" {
		t.Fatalf("unexpected labels: %#v", snapshots[0].Labels)
	}
	if snapshots[0].Value != 6.6771404e8 {
		t.Fatalf("unexpected scientific notation value: %g", snapshots[0].Value)
	}
}

func TestParsePrometheusRejectsEmptyLabelBetweenLabels(t *testing.T) {
	_, err := ParsePrometheus("order-service", `jvm_memory_used_bytes{area="heap",,id="G1 Old Gen"} 1`)
	if err == nil {
		t.Fatal("expected empty label between labels to be rejected")
	}
}
