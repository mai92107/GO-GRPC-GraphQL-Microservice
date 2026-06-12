package alert

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"golang-springboot-monitor-bot/internal/metric"
	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/repository"
)

type Thresholds struct {
	ResponseTimeWarning time.Duration
}

type MetricRule struct {
	Key         string
	ServiceName string
	MetricName  metric.Name
	Labels      map[string]string
	Operator    string
	Threshold   float64
	Severity    string
	Message     string
	Enabled     bool
}

type Engine struct {
	repo              *repository.MemoryRepository
	thresholds        Thresholds
	metricRules       []MetricRule
	consecutive       map[string]int
	lastNotifications map[string]time.Time
	mu                sync.Mutex
}

func (engine *Engine) ReplaceRules(thresholds Thresholds, rules []MetricRule) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	if thresholds.ResponseTimeWarning <= 0 {
		thresholds.ResponseTimeWarning = 2 * time.Second
	}
	engine.thresholds = thresholds
	engine.metricRules = append([]MetricRule(nil), rules...)
	clear(engine.consecutive)
}

func NewEngine(repo *repository.MemoryRepository, thresholds Thresholds, metricRules []MetricRule) *Engine {
	if thresholds.ResponseTimeWarning <= 0 {
		thresholds.ResponseTimeWarning = 2 * time.Second
	}
	return &Engine{
		repo:              repo,
		thresholds:        thresholds,
		metricRules:       metricRules,
		consecutive:       make(map[string]int),
		lastNotifications: make(map[string]time.Time),
	}
}

func (engine *Engine) EvaluateHealth(check model.HealthCheck) []model.AlertEvent {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	now := check.CheckedAt
	var events []model.AlertEvent

	if check.Status != "UP" || check.ErrorMessage != "" {
		message := fmt.Sprintf("[ALERT] %s health %s", check.ServiceName, check.Status)
		if check.ErrorMessage != "" {
			message = fmt.Sprintf("%s: %s", message, check.ErrorMessage)
		}
		if event, notify := engine.evaluateViolation(model.AlertEvent{
			ServiceName: check.ServiceName,
			RuleKey:     "health",
			MetricName:  "monitor_health_up",
			Severity:    "🆘",
			Message:     message,
			StartedAt:   now,
		}); notify {
			events = append(events, event)
		}
	} else {
		engine.resetViolation(check.ServiceName, "health")
		if event, resolved := engine.repo.ResolveAlert(check.ServiceName, "health", now); resolved {
			event.Message = fmt.Sprintf("[RECOVERY] %s is healthy again", check.ServiceName)
			events = append(events, event)
		}
	}

	if check.ResponseTime > engine.thresholds.ResponseTimeWarning {
		if event, notify := engine.evaluateViolation(model.AlertEvent{
			ServiceName: check.ServiceName,
			RuleKey:     "response_time",
			MetricName:  "monitor_response_time_ms",
			Severity:    "⚠️",
			Message: fmt.Sprintf(
				"[ALERT] %s response time high: %dms > %dms",
				check.ServiceName,
				check.ResponseTime.Milliseconds(),
				engine.thresholds.ResponseTimeWarning.Milliseconds(),
			),
			StartedAt: now,
		}); notify {
			events = append(events, event)
		}
	} else {
		engine.resetViolation(check.ServiceName, "response_time")
		if event, resolved := engine.repo.ResolveAlert(check.ServiceName, "response_time", now); resolved {
			event.Message = fmt.Sprintf("[RECOVERY] %s response time is normal", check.ServiceName)
			events = append(events, event)
		}
	}

	return events
}

func (engine *Engine) EvaluateMetrics(serviceName string, snapshots []model.MetricSnapshot, collectedAt time.Time) []model.AlertEvent {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	var events []model.AlertEvent

	for _, rule := range engine.metricRules {
		if !rule.Enabled || (rule.ServiceName != "*" && rule.ServiceName != serviceName) {
			continue
		}
		events = append(events, engine.evaluateMetricRule(serviceName, snapshots, rule, collectedAt)...)
	}

	return events
}

func (engine *Engine) evaluateMetricRule(serviceName string, snapshots []model.MetricSnapshot, rule MetricRule, at time.Time) []model.AlertEvent {
	matches := matchingSnapshots(snapshots, rule)
	if len(matches) == 0 {
		return nil
	}

	violations := make([]model.MetricSnapshot, 0, len(matches))
	for _, snapshot := range matches {
		if compare(snapshot.Value, rule.Operator, rule.Threshold) {
			violations = append(violations, snapshot)
		}
	}

	if len(violations) > 0 {
		snapshot := violations[0]
		message := rule.Message
		if message == "" {
			message = fmt.Sprintf(
				"[ALERT] %s metric %s%s value=%g %s threshold=%g",
				serviceName,
				rule.MetricName,
				formatLabels(snapshot.Labels),
				snapshot.Value,
				rule.Operator,
				rule.Threshold,
			)
		}
		event, notify := engine.evaluateViolation(model.AlertEvent{
			ServiceName: serviceName,
			RuleKey:     rule.Key,
			MetricName:  rule.MetricName.String(),
			Severity:    rule.Severity,
			Message:     message,
			StartedAt:   at,
		})
		if notify {
			return []model.AlertEvent{event}
		}
		return nil
	}

	engine.resetViolation(serviceName, rule.Key)
	if event, resolved := engine.repo.ResolveAlert(serviceName, rule.Key, at); resolved {
		event.Message = fmt.Sprintf("[RECOVERY] %s metric %s is normal", serviceName, rule.MetricName)
		return []model.AlertEvent{event}
	}
	return nil
}

const (
	requiredConsecutiveViolations = 3
	repeatNotificationInterval    = 15 * time.Minute
)

func (engine *Engine) evaluateViolation(candidate model.AlertEvent) (model.AlertEvent, bool) {
	key := candidate.ServiceName + ":" + candidate.RuleKey
	engine.consecutive[key]++
	if engine.consecutive[key] < requiredConsecutiveViolations {
		return model.AlertEvent{}, false
	}

	event, created := engine.repo.UpsertOpenAlert(candidate)
	lastNotified := engine.lastNotifications[key]
	if created || lastNotified.IsZero() || candidate.StartedAt.Sub(lastNotified) >= repeatNotificationInterval {
		engine.lastNotifications[key] = candidate.StartedAt
		return event, true
	}
	return model.AlertEvent{}, false
}

func (engine *Engine) resetViolation(serviceName, ruleKey string) {
	key := serviceName + ":" + ruleKey
	delete(engine.consecutive, key)
	delete(engine.lastNotifications, key)
}

func matchingSnapshots(snapshots []model.MetricSnapshot, rule MetricRule) []model.MetricSnapshot {
	var matches []model.MetricSnapshot
	for _, snapshot := range snapshots {
		if snapshot.Name != rule.MetricName.String() || !labelsMatch(snapshot.Labels, rule.Labels) {
			continue
		}
		matches = append(matches, snapshot)
	}
	return matches
}

func labelsMatch(actual, expected map[string]string) bool {
	for key, value := range expected {
		if actual[key] != value {
			return false
		}
	}
	return true
}

func compare(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", key, labels[key]))
	}
	return "{" + strings.Join(parts, ",") + "}"
}
