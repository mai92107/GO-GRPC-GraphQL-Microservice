package collector

import (
	"testing"

	"golang-springboot-monitor-bot/internal/model"
)

func TestExcludeMonitoringRequestMetrics(t *testing.T) {
	snapshots := []model.MetricSnapshot{
		{Name: "http_server_requests_seconds_count", Labels: map[string]string{"uri": "/actuator/health"}},
		{Name: "http_server_requests_seconds_sum", Labels: map[string]string{"uri": "/actuator/prometheus/"}},
		{Name: "http_server_requests_seconds_max", Labels: map[string]string{"uri": "/api/orders"}},
		{Name: "process_cpu_usage"},
	}

	result := excludeMonitoringRequestMetrics(snapshots, "/actuator/health", "/actuator/prometheus")
	if len(result) != 2 {
		t.Fatalf("expected two retained metrics, got %#v", result)
	}
	if result[0].Labels["uri"] != "/api/orders" || result[1].Name != "process_cpu_usage" {
		t.Fatalf("unexpected retained metrics: %#v", result)
	}
}

func TestExcludeMonitoringRequestMetricsUsesConfiguredPaths(t *testing.T) {
	snapshots := []model.MetricSnapshot{
		{Name: "http_server_requests_seconds_count", Labels: map[string]string{"uri": "/management/health"}},
		{Name: "http_server_requests_seconds_count", Labels: map[string]string{"uri": "/actuator/health"}},
	}

	result := excludeMonitoringRequestMetrics(snapshots, "/management/health?details=false", "/management/prometheus")
	if len(result) != 1 || result[0].Labels["uri"] != "/actuator/health" {
		t.Fatalf("configured monitoring path was not excluded correctly: %#v", result)
	}
}

func TestExcludeMonitoringRequestDoesNotRemoveNonHTTPRequestMetric(t *testing.T) {
	snapshots := []model.MetricSnapshot{
		{Name: "custom_metric", Labels: map[string]string{"uri": "/actuator/prometheus"}},
	}

	result := excludeMonitoringRequestMetrics(snapshots, "/actuator/prometheus")
	if len(result) != 1 {
		t.Fatalf("non-HTTP metric must be retained: %#v", result)
	}
}
