package scheduler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/alert"
	"golang-springboot-monitor-bot/internal/applog"
	"golang-springboot-monitor-bot/internal/collector"
	"golang-springboot-monitor-bot/internal/config"
	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/notifier"
	"golang-springboot-monitor-bot/internal/repository"
	"golang-springboot-monitor-bot/internal/trend"
)

type recordingNotifier struct {
	events []model.AlertEvent
}

func (notifier *recordingNotifier) Notify(_ context.Context, event model.AlertEvent) error {
	notifier.events = append(notifier.events, event)
	return nil
}

func TestSuccessfulCheckOnlyPrintsMonitoringProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/actuator/health" {
			_, _ = writer.Write([]byte(`{"status":"UP"}`))
			return
		}
		_, _ = writer.Write([]byte("process_cpu_usage 0.2\n"))
	}))
	defer server.Close()

	service := model.Service{Name: "order-service", BaseURL: server.URL, HealthPath: "/actuator/health", MetricsPath: "/actuator/prometheus", Enabled: true}
	repo := repository.NewMemoryRepository([]model.Service{service})
	var terminalOutput bytes.Buffer
	terminal := applog.NewTerminalWriter(&terminalOutput, "Asia/Taipei")
	logger, err := applog.New(config.LogConfig{Directory: filepath.Join(t.TempDir(), "logs"), Level: "info", RetentionDays: 14, Timezone: "Asia/Taipei"}, terminal)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()
	engine := alert.NewEngine(repo, alert.Thresholds{}, nil)
	trendRepo, err := trend.NewRepository(filepath.Join(t.TempDir(), "trends"), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	scheduler := New(repo, collector.NewHealthChecker(time.Second), engine, notifier.LogNotifier{Logger: logger}, true, trendRepo, logger, terminal)

	scheduler.RunOnce(context.Background())
	if got := terminalOutput.String(); got != "監控 order-service 中\n" {
		t.Fatalf("unexpected terminal output: %q", got)
	}
	if strings.Contains(terminalOutput.String(), "UP") || strings.Contains(terminalOutput.String(), "process_cpu_usage") {
		t.Fatal("successful details leaked to terminal")
	}

	scheduler.RunOnce(context.Background())
	if got := terminalOutput.String(); got != "監控 order-service 中\n" {
		t.Fatalf("unchanged health must not repeat monitoring progress: %q", got)
	}
}

func TestThirdHealthViolationSendsAlertWithCurrentTrendImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"status":"DOWN"}`))
	}))
	defer server.Close()

	service := model.Service{Name: "order-service", BaseURL: server.URL, HealthPath: "/actuator/health", Enabled: true}
	repo := repository.NewMemoryRepository([]model.Service{service})
	var terminalOutput bytes.Buffer
	terminal := applog.NewTerminalWriter(&terminalOutput, "Asia/Taipei")
	logger, err := applog.New(config.LogConfig{Directory: filepath.Join(t.TempDir(), "logs"), Level: "info", RetentionDays: 14, Timezone: "Asia/Taipei"}, terminal)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()
	trendRepo, err := trend.NewRepository(filepath.Join(t.TempDir(), "trends"), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	target := &recordingNotifier{}
	scheduler := New(repo, collector.NewHealthChecker(time.Second), alert.NewEngine(repo, alert.Thresholds{}, nil), target, true, trendRepo, logger, terminal)

	for i := 0; i < 3; i++ {
		scheduler.RunOnce(context.Background())
	}
	if len(target.events) != 1 {
		t.Fatalf("expected one alert after third violation, got %#v", target.events)
	}
	if target.events[0].MetricName != "monitor_health_up" || len(target.events[0].ImagePNG) == 0 {
		t.Fatalf("expected health alert trend image, got %#v", target.events[0])
	}
	samples, err := trendRepo.QueryMetrics("order-service", []string{"monitor_health_up", "monitor_response_time_ms"}, time.Hour, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	names := make(map[string]bool)
	for _, sample := range samples {
		names[sample.Name] = true
	}
	if !names["monitor_health_up"] || !names["monitor_response_time_ms"] {
		t.Fatalf("expected current health trend metrics, got %#v", names)
	}
}
