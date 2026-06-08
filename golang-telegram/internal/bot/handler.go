package bot

import (
	"fmt"
	"sort"
	"strings"

	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/repository"
)

type Handler struct {
	repo *repository.MemoryRepository
}

func NewHandler(repo *repository.MemoryRepository) Handler {
	return Handler{repo: repo}
}

func (handler Handler) HandleCommand(text string) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "empty command"
	}

	switch fields[0] {
	case "/list_services":
		return handler.listServices()
	case "/status":
		if len(fields) == 1 {
			return handler.statusAll()
		}
		return handler.status(fields[1])
	case "/alerts":
		return handler.alerts()
	default:
		return "unknown command"
	}
}

func (handler Handler) listServices() string {
	services := handler.repo.Services()
	if len(services) == 0 {
		return "No services."
	}

	lines := make([]string, 0, len(services))
	for _, service := range services {
		state := "disabled"
		if service.Enabled {
			state = "enabled"
		}
		lines = append(lines, fmt.Sprintf(
			"name=%s env=%s state=%s url=%s interval=%ds",
			service.Name,
			service.Environment,
			state,
			service.BaseURL,
			service.CheckIntervalSeconds,
		))
	}
	return strings.Join(lines, "\n")
}

func (handler Handler) statusAll() string {
	services := handler.repo.EnabledServices()
	if len(services) == 0 {
		return "No enabled services."
	}

	lines := make([]string, 0, len(services))
	for _, service := range services {
		lines = append(lines, handler.statusLine(service.Name))
	}
	return strings.Join(lines, "\n")
}

func (handler Handler) status(serviceName string) string {
	if _, ok := handler.repo.Service(serviceName); !ok {
		return fmt.Sprintf("Service %q not found.", serviceName)
	}
	return handler.statusLine(serviceName)
}

func (handler Handler) statusLine(serviceName string) string {
	check, ok := handler.repo.LastHealthCheck(serviceName)
	if !ok {
		return fmt.Sprintf("%s: no health check yet", serviceName)
	}
	return fmt.Sprintf(
		"%s: %s, response=%dms, checked_at=%s",
		check.ServiceName,
		check.Status,
		check.ResponseTime.Milliseconds(),
		check.CheckedAt.Format("2006-01-02 15:04:05"),
	)
}

func (handler Handler) alerts() string {
	alerts := handler.repo.OpenAlerts()
	if len(alerts) == 0 {
		return "No open alerts."
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].ServiceName < alerts[j].ServiceName
	})

	lines := make([]string, 0, len(alerts))
	for _, event := range alerts {
		lines = append(lines, formatAlert(event))
	}
	return strings.Join(lines, "\n")
}

func formatAlert(event model.AlertEvent) string {
	return fmt.Sprintf("%s %s\n", event.Severity, event.Message)
}
