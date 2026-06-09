package bot

import (
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/repository"
)

func TestHandleStatusCommand(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	repo.SaveHealthCheck(model.HealthCheck{
		ServiceName:  "order-service",
		Status:       "UP",
		ResponseTime: 120 * time.Millisecond,
		CheckedAt:    time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC),
	})
	repo.SaveMetricSnapshots([]model.MetricSnapshot{
		{
			ServiceName: "order-service",
			Name:        "process_cpu_usage",
			Value:       0.23,
			CollectedAt: time.Date(2026, 6, 4, 12, 0, 1, 0, time.UTC),
		},
		{
			ServiceName: "order-service",
			Name:        "http_server_requests_seconds_count",
			Labels:      map[string]string{"uri": "/orders", "method": "GET"},
			Value:       42,
			CollectedAt: time.Date(2026, 6, 4, 12, 0, 1, 0, time.UTC),
		},
	})

	response := NewHandler(repo).HandleCommand("/status order-service")
	expected := `order-service: UP, response=120ms, checked_at=2026-06-04 12:00:00
metrics:
- http_server_requests_seconds_count{method="GET",uri="/orders"}=42, collected_at=2026-06-04 12:00:01
- process_cpu_usage=0.23, collected_at=2026-06-04 12:00:01`
	if response != expected {
		t.Fatalf("unexpected response: %q", response)
	}
}

func TestHandleStatusCommandWithoutCollectedData(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})

	response := NewHandler(repo).HandleCommand("/status order-service")
	expected := "order-service: no health check yet\nmetrics: no metrics yet"
	if response != expected {
		t.Fatalf("unexpected response: %q", response)
	}
}

func TestModificationCommandsAreUnsupported(t *testing.T) {
	repo := repository.NewMemoryRepository(nil)
	response := NewHandler(repo).HandleCommand("/set_host order-service http://10.0.0.12:8081")

	if response != "unknown command" {
		t.Fatalf("unexpected response: %q", response)
	}
	if len(repo.Services()) != 0 {
		t.Fatal("telegram command must not modify services")
	}
}
