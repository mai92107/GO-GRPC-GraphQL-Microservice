package bot

import (
	"strings"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

func TestFormatMetricSnapshotOmitsNormalLabelsAndKeepsIdentifiers(t *testing.T) {
	output := formatMetricSnapshot(model.MetricSnapshot{
		Name: "http_server_requests_seconds_count",
		Labels: map[string]string{
			"error":     "none",
			"exception": "NONE",
			"method":    "OPTIONS",
			"outcome":   "SUCCESS",
			"status":    "200",
			"uri":       "/v3/api-docs/{group}",
		},
		Value:       12,
		CollectedAt: time.Date(2026, 6, 11, 15, 30, 0, 0, time.Local),
	})
	if strings.Contains(output, "error=") || strings.Contains(output, "status=") || strings.Contains(output, "OPTIONS") {
		t.Fatalf("normal labels were not omitted: %s", output)
	}
	if !strings.Contains(output, `監控內容：uri="/v3/api-docs/{group}"`) {
		t.Fatalf("identifying URI was not retained: %s", output)
	}
}

func TestFormatMetricValuesUsesReadableUnits(t *testing.T) {
	cases := map[string]string{
		formatMetricValue("process_cpu_usage", 0.234):            "23.40%",
		formatMetricValue("jvm_memory_used_bytes", 536870912):    "512.00 MB",
		formatMetricValue("process_uptime_seconds", 356580):      "4 天 3 小時 3 分鐘",
		formatMetricValue("http_server_requests_seconds_max", 2): "2.00 秒",
	}
	for got, want := range cases {
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}

func TestSystemCPUAndThreadStateDisplay(t *testing.T) {
	if got := formatMetricValue("system_cpu_usage", 0.35); got != "35.00%" {
		t.Fatalf("unexpected system CPU value: %s", got)
	}
	output := formatMetricSnapshot(model.MetricSnapshot{
		Name:        "jvm_threads_states_threads",
		Labels:      map[string]string{"state": "blocked"},
		Value:       2,
		CollectedAt: time.Now(),
	})
	if !strings.Contains(output, "JVM 執行緒狀態") || !strings.Contains(output, `state="blocked"`) {
		t.Fatalf("unexpected thread state output: %s", output)
	}
}
