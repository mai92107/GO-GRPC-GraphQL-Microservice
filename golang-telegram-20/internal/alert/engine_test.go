package alert

import (
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/metric"
	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/repository"
)

func TestEvaluateHealthCreatesOneAlertThenRecovery(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	engine := NewEngine(repo, Thresholds{}, nil)
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	events := engine.EvaluateHealth(model.HealthCheck{
		ServiceName:  "order-service",
		Status:       "DOWN",
		ErrorMessage: "timeout",
		CheckedAt:    now,
	})
	if len(events) != 1 {
		t.Fatalf("expected first alert, got %d", len(events))
	}

	events = engine.EvaluateHealth(model.HealthCheck{
		ServiceName:  "order-service",
		Status:       "DOWN",
		ErrorMessage: "timeout",
		CheckedAt:    now.Add(time.Minute),
	})
	if len(events) != 0 {
		t.Fatalf("expected duplicate alert to be suppressed, got %d", len(events))
	}

	events = engine.EvaluateHealth(model.HealthCheck{
		ServiceName: "order-service",
		Status:      "UP",
		CheckedAt:   now.Add(2 * time.Minute),
	})
	if len(events) != 1 {
		t.Fatalf("expected recovery event, got %d", len(events))
	}
	if events[0].Status != model.AlertStatusResolved {
		t.Fatalf("expected resolved event, got %s", events[0].Status)
	}
}

func TestEvaluateMetricsUsesConfiguredRuleAndRecovers(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	engine := NewEngine(repo, Thresholds{}, []MetricRule{{
		Key:         "high-live-threads",
		ServiceName: "order-service",
		MetricName:  metric.JVMThreadsLiveThreads,
		Operator:    ">",
		Threshold:   100,
		Severity:    "warning",
		Enabled:     true,
	}})
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	events := engine.EvaluateMetrics("order-service", []model.MetricSnapshot{{
		ServiceName: "order-service",
		Name:        "jvm_threads_live_threads",
		Value:       120,
	}}, now)
	if len(events) != 1 || events[0].Status != model.AlertStatusOpen {
		t.Fatalf("expected open metric alert, got %#v", events)
	}

	events = engine.EvaluateMetrics("order-service", []model.MetricSnapshot{{
		ServiceName: "order-service",
		Name:        "jvm_threads_live_threads",
		Value:       80,
	}}, now.Add(time.Minute))
	if len(events) != 1 || events[0].Status != model.AlertStatusResolved {
		t.Fatalf("expected metric recovery, got %#v", events)
	}
}

func TestEvaluateMetricsDoesNotRecoverWhenMetricIsMissing(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	engine := NewEngine(repo, Thresholds{}, []MetricRule{{
		Key:         "high-heap",
		ServiceName: "*",
		MetricName:  metric.JVMMemoryUsedBytes,
		Labels:      map[string]string{"area": "heap"},
		Operator:    ">",
		Threshold:   100,
		Severity:    "warning",
		Enabled:     true,
	}})
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	events := engine.EvaluateMetrics("order-service", []model.MetricSnapshot{{
		Name:   "jvm_memory_used_bytes",
		Labels: map[string]string{"area": "heap"},
		Value:  120,
	}}, now)
	if len(events) != 1 {
		t.Fatalf("expected alert, got %#v", events)
	}

	events = engine.EvaluateMetrics("order-service", nil, now.Add(time.Minute))
	if len(events) != 0 {
		t.Fatalf("missing metric must not trigger recovery, got %#v", events)
	}
}
