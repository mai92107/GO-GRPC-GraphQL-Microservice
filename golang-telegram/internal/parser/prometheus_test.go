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

func TestParsePrometheusAcceptsBraceInsideQuotedLabelValue(t *testing.T) {
	input := `http_server_requests_seconds_count{method="GET",uri="/v3/api-docs/{group}",status="200"} 12`

	snapshots, err := ParsePrometheus("www-service", input)
	if err != nil {
		t.Fatalf("parse prometheus: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected one snapshot, got %d", len(snapshots))
	}
	if snapshots[0].Labels["uri"] != "/v3/api-docs/{group}" {
		t.Fatalf("unexpected URI label: %q", snapshots[0].Labels["uri"])
	}
	if snapshots[0].Value != 12 {
		t.Fatalf("unexpected value: %g", snapshots[0].Value)
	}
}

func TestParsePrometheusAcceptsEscapedQuoteAndBraceInsideLabelValue(t *testing.T) {
	input := `http_server_requests_seconds_count{uri="/example/\"quoted}\""} 1`

	snapshots, err := ParsePrometheus("www-service", input)
	if err != nil {
		t.Fatalf("parse prometheus: %v", err)
	}
	if snapshots[0].Labels["uri"] != `/example/"quoted}"` {
		t.Fatalf("unexpected URI label: %q", snapshots[0].Labels["uri"])
	}
}

func TestParsePrometheusSupportsThreadAndSystemCPUMetrics(t *testing.T) {
	input := `
jvm_threads_live_threads 30
jvm_threads_daemon_threads 20
jvm_threads_peak_threads 42
jvm_threads_started_threads_total 100
jvm_threads_states_threads{state="runnable"} 8
jvm_threads_states_threads{state="blocked"} 1
system_cpu_usage 0.35
system_cpu_count 8
`
	snapshots, err := ParsePrometheus("www-service", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshots) != 8 {
		t.Fatalf("expected 8 snapshots, got %d", len(snapshots))
	}
	if snapshots[4].Labels["state"] != "runnable" {
		t.Fatalf("thread state label was not retained: %#v", snapshots[4].Labels)
	}
}
