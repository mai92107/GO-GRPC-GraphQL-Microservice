package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"golang-springboot-monitor-bot/internal/alert"
	"golang-springboot-monitor-bot/internal/collector"
	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/notifier"
	"golang-springboot-monitor-bot/internal/repository"
)

type Scheduler struct {
	repo           *repository.MemoryRepository
	checker        *collector.HealthChecker
	alertEngine    *alert.Engine
	notifier       notifier.Notifier
	metricsEnabled bool
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
) *Scheduler {
	return &Scheduler{
		repo:           repo,
		checker:        checker,
		alertEngine:    alertEngine,
		notifier:       notifier,
		metricsEnabled: metricsEnabled,
		inFlight:       make(map[string]bool),
		lastCheckedAt:  make(map[string]time.Time),
	}
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

	scheduler.repo.SaveHealthCheck(check)
	logHealth(check)
	scheduler.notify(ctx, scheduler.alertEngine.EvaluateHealth(check))

	if !scheduler.metricsEnabled || check.Status != "UP" {
		return
	}

	metricsResult := scheduler.checker.CollectMetrics(ctx, service)
	if metricsResult.Error != nil {
		log.Printf("metrics service=%s error=%q", service.Name, metricsResult.Error.Error())
		return
	}

	scheduler.repo.SaveMetricSnapshots(metricsResult.Snapshots)
	log.Printf("metrics service=%s snapshots=%d response=%dms", service.Name, len(metricsResult.Snapshots), metricsResult.ResponseTime.Milliseconds())
	scheduler.notify(ctx, scheduler.alertEngine.EvaluateMetrics(service.Name, metricsResult.Snapshots, metricsResult.CollectedAt))
}

func (scheduler *Scheduler) notify(ctx context.Context, events []model.AlertEvent) {
	for _, event := range events {
		if err := scheduler.notifier.Notify(ctx, event); err != nil {
			log.Printf("notify service=%s rule=%s error=%q", event.ServiceName, event.RuleKey, err.Error())
		}
	}
}

func logHealth(check model.HealthCheck) {
	if check.ErrorMessage != "" {
		log.Printf(
			"health service=%s env=%s status=%s http=%d response=%dms error=%q",
			check.ServiceName,
			check.Environment,
			check.Status,
			check.HTTPStatusCode,
			check.ResponseTime.Milliseconds(),
			check.ErrorMessage,
		)
		return
	}

	log.Printf(
		"health service=%s env=%s status=%s http=%d response=%dms",
		check.ServiceName,
		check.Environment,
		check.Status,
		check.HTTPStatusCode,
		check.ResponseTime.Milliseconds(),
	)
}
