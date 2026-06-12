package alert

import (
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/metric"
	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/repository"
)

func TestEvaluateHealthCreatesAlertAfterThreeViolationsThenRecovery(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	engine := NewEngine(repo, Thresholds{}, nil)
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 3; i++ {
		events := engine.EvaluateHealth(model.HealthCheck{
			ServiceName:  "order-service",
			Status:       "DOWN",
			ErrorMessage: "timeout",
			CheckedAt:    now.Add(time.Duration(i) * time.Minute),
		})
		if i < 2 && len(events) != 0 {
			t.Fatalf("expected violation %d to be suppressed, got %#v", i+1, events)
		}
		if i == 2 && (len(events) != 1 || events[0].MetricName != "monitor_health_up") {
			t.Fatalf("expected third violation alert, got %#v", events)
		}
	}

	events := engine.EvaluateHealth(model.HealthCheck{
		ServiceName: "order-service",
		Status:      "UP",
		CheckedAt:   now.Add(3 * time.Minute),
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

	var events []model.AlertEvent
	for i := 0; i < 3; i++ {
		events = engine.EvaluateMetrics("order-service", []model.MetricSnapshot{{
			ServiceName: "order-service",
			Name:        "jvm_threads_live_threads",
			Value:       120,
		}}, now.Add(time.Duration(i)*time.Minute))
	}
	if len(events) != 1 || events[0].Status != model.AlertStatusOpen || events[0].MetricName != "jvm_threads_live_threads" {
		t.Fatalf("expected open metric alert on third violation, got %#v", events)
	}

	events = engine.EvaluateMetrics("order-service", []model.MetricSnapshot{{
		ServiceName: "order-service",
		Name:        "jvm_threads_live_threads",
		Value:       80,
	}}, now.Add(3*time.Minute))
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

	var events []model.AlertEvent
	for i := 0; i < 3; i++ {
		events = engine.EvaluateMetrics("order-service", []model.MetricSnapshot{{
			Name:   "jvm_memory_used_bytes",
			Labels: map[string]string{"area": "heap"},
			Value:  120,
		}}, now.Add(time.Duration(i)*time.Minute))
	}
	if len(events) != 1 {
		t.Fatalf("expected alert, got %#v", events)
	}

	events = engine.EvaluateMetrics("order-service", nil, now.Add(3*time.Minute))
	if len(events) != 0 {
		t.Fatalf("missing metric must not trigger recovery, got %#v", events)
	}
}

func TestNormalResultResetsConsecutiveViolationCount(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	engine := NewEngine(repo, Thresholds{}, nil)
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 2; i++ {
		if events := engine.EvaluateHealth(model.HealthCheck{ServiceName: "order-service", Status: "DOWN", CheckedAt: now.Add(time.Duration(i) * time.Minute)}); len(events) != 0 {
			t.Fatalf("unexpected early alert: %#v", events)
		}
	}
	engine.EvaluateHealth(model.HealthCheck{ServiceName: "order-service", Status: "UP", CheckedAt: now.Add(2 * time.Minute)})
	for i := 0; i < 2; i++ {
		if events := engine.EvaluateHealth(model.HealthCheck{ServiceName: "order-service", Status: "DOWN", CheckedAt: now.Add(time.Duration(i+3) * time.Minute)}); len(events) != 0 {
			t.Fatalf("normal result must reset count: %#v", events)
		}
	}
}

func TestOpenAlertRepeatsEveryFifteenMinutes(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	engine := NewEngine(repo, Thresholds{}, nil)
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	check := func(at time.Time) []model.AlertEvent {
		return engine.EvaluateHealth(model.HealthCheck{ServiceName: "order-service", Status: "DOWN", CheckedAt: at})
	}
	check(now)
	check(now.Add(time.Minute))
	if events := check(now.Add(2 * time.Minute)); len(events) != 1 {
		t.Fatalf("expected initial alert, got %#v", events)
	}
	if events := check(now.Add(16 * time.Minute)); len(events) != 0 {
		t.Fatalf("must not repeat before 15 minutes, got %#v", events)
	}
	if events := check(now.Add(17 * time.Minute)); len(events) != 1 {
		t.Fatalf("expected 15-minute repeat, got %#v", events)
	}
}
