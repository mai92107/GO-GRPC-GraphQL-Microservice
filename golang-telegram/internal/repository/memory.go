package repository

import (
	"sort"
	"sync"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

type MemoryRepository struct {
	mu           sync.RWMutex
	services     map[string]model.Service
	healthChecks map[string][]model.HealthCheck
	metrics      map[string][]model.MetricSnapshot
	alerts       map[string]model.AlertEvent
	alertHistory []model.AlertEvent
	nextAlertID  int64
}

func NewMemoryRepository(services []model.Service) *MemoryRepository {
	repo := &MemoryRepository{
		services:     make(map[string]model.Service, len(services)),
		healthChecks: make(map[string][]model.HealthCheck),
		metrics:      make(map[string][]model.MetricSnapshot),
		alerts:       make(map[string]model.AlertEvent),
	}
	for _, service := range services {
		repo.services[service.Name] = service
	}
	return repo
}

func (repo *MemoryRepository) EnabledServices() []model.Service {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	return sortedServices(repo.services, true)
}

func (repo *MemoryRepository) Services() []model.Service {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	return sortedServices(repo.services, false)
}

func sortedServices(source map[string]model.Service, enabledOnly bool) []model.Service {
	services := make([]model.Service, 0, len(source))
	for _, service := range source {
		if !enabledOnly || service.Enabled {
			services = append(services, service)
		}
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services
}

func (repo *MemoryRepository) Service(name string) (model.Service, bool) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	service, ok := repo.services[name]
	return service, ok
}

func (repo *MemoryRepository) SaveHealthCheck(check model.HealthCheck) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.healthChecks[check.ServiceName] = append(repo.healthChecks[check.ServiceName], check)
}

func (repo *MemoryRepository) SaveMetricSnapshots(snapshots []model.MetricSnapshot) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, snapshot := range snapshots {
		repo.metrics[snapshot.ServiceName] = append(repo.metrics[snapshot.ServiceName], snapshot)
	}
}

func (repo *MemoryRepository) LastHealthCheck(serviceName string) (*model.HealthCheck, bool) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	checks := repo.healthChecks[serviceName]
	if len(checks) == 0 {
		return nil, false
	}
	check := checks[len(checks)-1]
	return &check, true
}

func (repo *MemoryRepository) LastMetricSnapshots(serviceName string) []model.MetricSnapshot {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	all := repo.metrics[serviceName]
	latestByKey := make(map[string]model.MetricSnapshot)
	for _, snapshot := range all {
		latestByKey[metricKey(snapshot)] = snapshot
	}

	result := make([]model.MetricSnapshot, 0, len(latestByKey))
	for _, snapshot := range latestByKey {
		result = append(result, snapshot)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (repo *MemoryRepository) UpsertOpenAlert(event model.AlertEvent) (model.AlertEvent, bool) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	key := alertKey(event.ServiceName, event.RuleKey)
	if existing, ok := repo.alerts[key]; ok && existing.Status == model.AlertStatusOpen {
		return existing, false
	}

	repo.nextAlertID++
	event.ID = repo.nextAlertID
	event.Status = model.AlertStatusOpen
	repo.alerts[key] = event
	repo.alertHistory = append(repo.alertHistory, event)
	return event, true
}

func (repo *MemoryRepository) ResolveAlert(serviceName, ruleKey string, resolvedAt time.Time) (model.AlertEvent, bool) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	key := alertKey(serviceName, ruleKey)
	event, ok := repo.alerts[key]
	if !ok || event.Status != model.AlertStatusOpen {
		return model.AlertEvent{}, false
	}

	event.Status = model.AlertStatusResolved
	event.ResolvedAt = &resolvedAt
	delete(repo.alerts, key)
	repo.alertHistory = append(repo.alertHistory, event)
	return event, true
}

func (repo *MemoryRepository) OpenAlerts() []model.AlertEvent {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	result := make([]model.AlertEvent, 0, len(repo.alerts))
	for _, event := range repo.alerts {
		result = append(result, event)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartedAt.Before(result[j].StartedAt)
	})
	return result
}

func (repo *MemoryRepository) Snapshot(serviceName string) (model.ServiceSnapshot, bool) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	service, ok := repo.services[serviceName]
	if !ok {
		return model.ServiceSnapshot{}, false
	}

	snapshot := model.ServiceSnapshot{Service: service}
	if checks := repo.healthChecks[serviceName]; len(checks) > 0 {
		check := checks[len(checks)-1]
		snapshot.LastHealth = &check
	}

	for _, metric := range repo.metrics[serviceName] {
		snapshot.LastMetrics = append(snapshot.LastMetrics, metric)
	}
	for _, alert := range repo.alerts {
		if alert.ServiceName == serviceName {
			snapshot.OpenAlerts = append(snapshot.OpenAlerts, alert)
		}
	}
	for _, alert := range repo.alertHistory {
		if alert.ServiceName == serviceName {
			snapshot.RecentAlerts = append(snapshot.RecentAlerts, alert)
		}
	}
	return snapshot, true
}

func alertKey(serviceName, ruleKey string) string {
	return serviceName + ":" + ruleKey
}

func metricKey(snapshot model.MetricSnapshot) string {
	key := snapshot.Name
	labels := make([]string, 0, len(snapshot.Labels))
	for label := range snapshot.Labels {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	for _, label := range labels {
		value := snapshot.Labels[label]
		key += "|" + label + "=" + value
	}
	return key
}
