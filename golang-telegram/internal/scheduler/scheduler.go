package scheduler

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang-springboot-monitor-bot/internal/alert"
	"golang-springboot-monitor-bot/internal/applog"
	"golang-springboot-monitor-bot/internal/chart"
	"golang-springboot-monitor-bot/internal/collector"
	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/notifier"
	"golang-springboot-monitor-bot/internal/repository"
	"golang-springboot-monitor-bot/internal/trend"
)

type Scheduler struct {
	repo           *repository.MemoryRepository
	checker        *collector.HealthChecker
	alertEngine    *alert.Engine
	notifier       notifier.Notifier
	metricsEnabled bool
	trends         *trend.Repository
	renderer       *chart.Renderer
	logger         applog.ApplicationLogger
	terminal       applog.TerminalStatusWriter
	inFlight       map[string]bool
	lastCheckedAt  map[string]time.Time
	mu             sync.Mutex
}

func New(
	repo *repository.MemoryRepository,
	checker *collector.HealthChecker,
	alertEngine *alert.Engine,
	notifier notifier.Notifier,
	metricsEnabled bool,
	trends *trend.Repository,
	logger applog.ApplicationLogger,
	terminal applog.TerminalStatusWriter,
) *Scheduler {
	return &Scheduler{
		repo:           repo,
		checker:        checker,
		alertEngine:    alertEngine,
		notifier:       notifier,
		metricsEnabled: metricsEnabled,
		trends:         trends,
		renderer:       chart.NewRenderer(),
		logger:         logger,
		terminal:       terminal,
		inFlight:       make(map[string]bool),
		lastCheckedAt:  make(map[string]time.Time),
	}
}

func (scheduler *Scheduler) SetMetricsEnabled(enabled bool) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	scheduler.metricsEnabled = enabled
}

func (scheduler *Scheduler) CheckNow(ctx context.Context, serviceName string) error {
	service, ok := scheduler.repo.Service(serviceName)
	if !ok {
		return fmt.Errorf("service %q not found", serviceName)
	}
	scheduler.mu.Lock()
	if scheduler.inFlight[serviceName] {
		scheduler.mu.Unlock()
		return fmt.Errorf("service %q check is already in progress", serviceName)
	}
	scheduler.inFlight[serviceName] = true
	scheduler.mu.Unlock()
	defer scheduler.finishCheck(serviceName)
	scheduler.checkService(ctx, service)
	return nil
}

func (scheduler *Scheduler) RunOnce(ctx context.Context) {
	var wg sync.WaitGroup
	for _, service := range scheduler.repo.EnabledServices() {
		wg.Add(1)
		go func(service model.Service) {
			defer wg.Done()
			scheduler.checkService(ctx, service)
		}(service)
	}
	wg.Wait()
}

func (scheduler *Scheduler) Run(ctx context.Context) {
	var wg sync.WaitGroup
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		for _, service := range scheduler.repo.EnabledServices() {
			if !scheduler.shouldCheck(service, time.Now()) {
				continue
			}
			wg.Add(1)
			go func(service model.Service) {
				defer wg.Done()
				defer scheduler.finishCheck(service.Name)
				scheduler.checkService(ctx, service)
			}(service)
		}

		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case <-ticker.C:
		}
	}
}

func (scheduler *Scheduler) shouldCheck(service model.Service, now time.Time) bool {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	if scheduler.inFlight[service.Name] {
		return false
	}

	interval := time.Duration(service.CheckIntervalSeconds) * time.Second
	if lastCheckedAt, ok := scheduler.lastCheckedAt[service.Name]; ok && now.Sub(lastCheckedAt) < interval {
		return false
	}

	scheduler.inFlight[service.Name] = true
	scheduler.lastCheckedAt[service.Name] = now
	return true
}

func (scheduler *Scheduler) finishCheck(serviceName string) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	delete(scheduler.inFlight, serviceName)
}

func (scheduler *Scheduler) checkService(ctx context.Context, service model.Service) {
	previousHealth, hadPreviousHealth := scheduler.repo.LastHealthCheck(service.Name)
	healthResult := scheduler.checker.Check(ctx, service)
	check := model.HealthCheck{
		ServiceName:    service.Name,
		Environment:    service.Environment,
		Status:         healthResult.Status,
		HTTPStatusCode: healthResult.HTTPStatusCode,
		ResponseTime:   healthResult.ResponseTime,
		CheckedAt:      healthResult.CheckedAt,
	}
	if healthResult.Error != nil {
		check.ErrorMessage = healthResult.Error.Error()
	}

	if !hadPreviousHealth || healthStateChanged(*previousHealth, check) {
		scheduler.terminal.Monitoring(service.Name)
	}
	scheduler.repo.SaveHealthCheck(check)
	scheduler.trends.Add([]model.MetricSample{
		{
			ServiceName: service.Name,
			Name:        "monitor_health_up",
			Value:       healthValue(check),
			CollectedAt: check.CheckedAt,
		},
		{
			ServiceName: service.Name,
			Name:        "monitor_response_time_ms",
			Value:       float64(check.ResponseTime.Milliseconds()),
			CollectedAt: check.CheckedAt,
		},
	})
	if check.ErrorMessage != "" {
		scheduler.logger.Error(ctx, "health_check_failed", healthResult.Error, applog.Field{Key: "service", Value: service.Name})
	} else {
		scheduler.logger.Info(ctx, "health_check_completed",
			applog.Field{Key: "service", Value: service.Name},
			applog.Field{Key: "duration_ms", Value: check.ResponseTime.Milliseconds()},
			applog.Field{Key: "status", Value: check.Status},
		)
	}
	scheduler.notify(ctx, scheduler.alertEngine.EvaluateHealth(check))

	scheduler.mu.Lock()
	metricsEnabled := scheduler.metricsEnabled
	scheduler.mu.Unlock()
	if !metricsEnabled || check.Status != "UP" {
		return
	}

	metricsResult := scheduler.checker.CollectMetrics(ctx, service)
	if metricsResult.Error != nil {
		scheduler.logger.Error(ctx, "metrics_collection_failed", metricsResult.Error, applog.Field{Key: "service", Value: service.Name})
		return
	}

	scheduler.repo.SaveMetricSnapshots(metricsResult.Snapshots)
	scheduler.trends.Add(metricsResult.Snapshots)
	scheduler.logger.Info(ctx, "metrics_collection_completed",
		applog.Field{Key: "service", Value: service.Name},
		applog.Field{Key: "snapshots", Value: len(metricsResult.Snapshots)},
	)
	scheduler.notify(ctx, scheduler.alertEngine.EvaluateMetrics(service.Name, metricsResult.Snapshots, metricsResult.CollectedAt))
}

func healthStateChanged(previous, current model.HealthCheck) bool {
	return previous.Status != current.Status ||
		previous.HTTPStatusCode != current.HTTPStatusCode ||
		previous.ErrorMessage != current.ErrorMessage
}

func (scheduler *Scheduler) notify(ctx context.Context, events []model.AlertEvent) {
	for _, event := range events {
		if event.Status == model.AlertStatusOpen {
			scheduler.attachAlertChart(ctx, &event)
		}
		if err := scheduler.notifier.Notify(ctx, event); err != nil {
			scheduler.logger.Error(ctx, "notification_failed", err,
				applog.Field{Key: "service", Value: event.ServiceName},
				applog.Field{Key: "rule", Value: event.RuleKey},
			)
		}
	}
}

func (scheduler *Scheduler) attachAlertChart(ctx context.Context, event *model.AlertEvent) {
	if scheduler.trends == nil || scheduler.renderer == nil || event.MetricName == "" {
		return
	}
	metricNames := scheduler.relatedMetricNames(event.ServiceName, event.MetricName)
	samples, err := scheduler.trends.QueryMetrics(event.ServiceName, metricNames, time.Hour, time.Now())
	if err != nil {
		scheduler.logger.Error(ctx, "alert_chart_query_failed", err,
			applog.Field{Key: "service", Value: event.ServiceName},
			applog.Field{Key: "rule", Value: event.RuleKey},
		)
		return
	}
	var output bytes.Buffer
	if err := scheduler.renderer.RenderNormalized(&output, event.ServiceName, event.MetricName, "1h", samples); err != nil {
		scheduler.logger.Error(ctx, "alert_chart_render_failed", err,
			applog.Field{Key: "service", Value: event.ServiceName},
			applog.Field{Key: "rule", Value: event.RuleKey},
		)
		return
	}
	event.ImagePNG = output.Bytes()
}

func (scheduler *Scheduler) relatedMetricNames(serviceName, primary string) []string {
	groups := [][]string{
		{"monitor_health_up", "monitor_response_time_ms", "process_uptime_seconds"},
		{"process_cpu_usage", "system_cpu_usage"},
		{"http_server_requests_seconds_count", "http_server_requests_seconds_sum", "http_server_requests_seconds_max"},
		{"jvm_memory_used_bytes", "jvm_memory_max_bytes", "jvm_gc_pause_seconds_count", "jvm_gc_pause_seconds_sum"},
	}
	for _, group := range groups {
		for _, name := range group {
			if name == primary {
				return group
			}
		}
	}
	prefix := ""
	switch {
	case strings.HasPrefix(primary, "jvm_threads_"):
		prefix = "jvm_threads_"
	case strings.HasPrefix(primary, "hikaricp_connections_"):
		prefix = "hikaricp_connections_"
	}
	if prefix == "" {
		return []string{primary}
	}
	names := map[string]struct{}{primary: {}}
	for _, snapshot := range scheduler.repo.LastMetricSnapshots(serviceName) {
		if strings.HasPrefix(snapshot.Name, prefix) {
			names[snapshot.Name] = struct{}{}
		}
	}
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	return result
}

func healthValue(check model.HealthCheck) float64 {
	if check.Status == "UP" && check.ErrorMessage == "" {
		return 1
	}
	return 0
}
