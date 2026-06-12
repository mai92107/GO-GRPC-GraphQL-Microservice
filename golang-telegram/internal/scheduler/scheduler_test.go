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
