package bot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/config"
	"golang-springboot-monitor-bot/internal/metric"
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
	expected := `order-service：UP`
	if response != expected {
		t.Fatalf("unexpected response: %q", response)
	}

	response = NewHandler(repo).HandleCommand("/metric order-service")
	expected = `order-service 指標資料：

- HTTP 請求次數（http_server_requests_seconds_count）
  監控內容：method="GET"、uri="/orders"
  數值：42
  收集時間：2026-06-04 12:00:01

- 程序 CPU 使用率（process_cpu_usage）
  數值：23.00%
  收集時間：2026-06-04 12:00:01`
	if response != expected {
		t.Fatalf("unexpected response: %q", response)
	}
}

func TestHandleStatusCommandWithoutCollectedData(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})

	response := NewHandler(repo).HandleCommand("/status order-service")
	expected := "order-service：尚未檢查"
	if response != expected {
		t.Fatalf("unexpected response: %q", response)
	}

	response = NewHandler(repo).HandleCommand("/metric order-service")
	expected = "order-service：目前尚無指標資料。"
	if response != expected {
		t.Fatalf("unexpected metric response: %q", response)
	}
}

func TestModificationCommandsAreUnsupported(t *testing.T) {
	repo := repository.NewMemoryRepository(nil)
	response := NewHandler(repo).HandleCommand("/set_host order-service http://10.0.0.12:8081")

	if response != "無法識別此指令。" {
		t.Fatalf("unexpected response: %q", response)
	}
	if len(repo.Services()) != 0 {
		t.Fatal("telegram command must not modify services")
	}
}

func TestAlertAcknowledgement(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	alert, _ := repo.UpsertOpenAlert(model.AlertEvent{ServiceName: "order-service", RuleKey: "health", StartedAt: time.Now()})
	handler := NewHandler(repo)
	reply := handler.Handle(context.Background(), "123", "", fmt.Sprintf("ack:%d", alert.ID))
	if reply.Text != "已確認告警。" {
		t.Fatalf("unexpected reply: %s", reply.Text)
	}
	if repo.OpenAlerts()[0].AcknowledgedAt == nil {
		t.Fatal("alert was not acknowledged")
	}
}

func TestMetricInlineKeyboardCategoryFlow(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "www-service", Enabled: true}})
	repo.SaveMetricSnapshots([]model.MetricSnapshot{
		{ServiceName: "www-service", Name: "http_server_requests_seconds_count", Value: 10, CollectedAt: time.Now()},
		{ServiceName: "www-service", Name: "jvm_memory_used_bytes", Value: 1024, CollectedAt: time.Now()},
		{ServiceName: "www-service", Name: "process_cpu_usage", Value: 0.2, CollectedAt: time.Now()},
	})
	handler := NewHandler(repo)

	reply := handler.Handle(context.Background(), "123", "/metric", "")
	if reply.Text != "請選擇要查詢指標的服務：" || !keyboardContains(reply.Keyboard, "metric_service:www-service") {
		t.Fatalf("unexpected service menu: %#v", reply)
	}

	reply = handler.Handle(context.Background(), "123", "", "metric_service:www-service")
	if !keyboardContains(reply.Keyboard, "metric_group:www-service:http") ||
		!keyboardContains(reply.Keyboard, "metric_group:www-service:jvm") ||
		!keyboardContains(reply.Keyboard, "metric_group:www-service:process") {
		t.Fatalf("unexpected group menu: %#v", reply.Keyboard)
	}

	reply = handler.Handle(context.Background(), "123", "", "metric_group:www-service:http")
	if !keyboardContains(reply.Keyboard, "metric_subgroup:www-service:http:server") {
		t.Fatalf("unexpected subgroup menu: %#v", reply.Keyboard)
	}

	reply = handler.Handle(context.Background(), "123", "", "metric_subgroup:www-service:http:server")
	if !strings.Contains(reply.Text, "http_server_requests_seconds_count") {
		t.Fatalf("expected selected metric in response: %s", reply.Text)
	}
	if strings.Contains(reply.Text, "jvm_memory_used_bytes") || strings.Contains(reply.Text, "process_cpu_usage") {
		t.Fatalf("response contains metrics outside selected category: %s", reply.Text)
	}
}

func TestTrendInlineKeyboardCategoryFlow(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "www-service", Enabled: true}})
	repo.SaveMetricSnapshots([]model.MetricSnapshot{
		{ServiceName: "www-service", Name: "http_server_requests_seconds_count", Value: 10, CollectedAt: time.Now()},
		{ServiceName: "www-service", Name: "http_server_requests_seconds_max", Value: 0.5, CollectedAt: time.Now()},
	})
	handler := NewHandler(repo)

	reply := handler.Handle(context.Background(), "123", "/trend", "")
	if !keyboardContains(reply.Keyboard, "ts:www-service") {
		t.Fatalf("unexpected trend service menu: %#v", reply.Keyboard)
	}

	reply = handler.Handle(context.Background(), "123", "", "ts:www-service")
	if !keyboardContains(reply.Keyboard, "tg:www-service:http") {
		t.Fatalf("unexpected trend group menu: %#v", reply.Keyboard)
	}

	reply = handler.Handle(context.Background(), "123", "", "tg:www-service:http")
	if !keyboardContains(reply.Keyboard, "tu:www-service:http:server") {
		t.Fatalf("unexpected trend subgroup menu: %#v", reply.Keyboard)
	}

	reply = handler.Handle(context.Background(), "123", "", "tu:www-service:http:server")
	if !keyboardContains(reply.Keyboard, "tm:www-service:http_server_requests_seconds_count") ||
		!keyboardContains(reply.Keyboard, "tm:www-service:http_server_requests_seconds_max") {
		t.Fatalf("unexpected trend metric menu: %#v", reply.Keyboard)
	}

	reply = handler.Handle(context.Background(), "123", "", "tm:www-service:http_server_requests_seconds_count")
	if !keyboardContains(reply.Keyboard, "tr:www-service:http_server_requests_seconds_count:1h") {
		t.Fatalf("unexpected trend range menu: %#v", reply.Keyboard)
	}
	for _, row := range reply.Keyboard {
		for _, button := range row {
			if len([]byte(button.Data)) > 64 {
				t.Fatalf("callback_data exceeds Telegram limit: %q", button.Data)
			}
		}
	}
}

func TestRuleMenuUsesReadableRowsForLongRuleKey(t *testing.T) {
	repo := repository.NewMemoryRepository(nil)
	handler := NewHandler(repo)
	handler.config = config.NewManager("", config.Config{AlertRules: []config.AlertRule{{
		Key:        "very-long-alert-rule-key-that-does-not-fit-on-one-row",
		MetricName: metric.JVMMemoryUsedBytes,
		Enabled:    false,
	}}}, nil)

	reply := handler.ruleMenu()
	if len(reply.Keyboard) < 4 {
		t.Fatalf("expected title and operation rows: %#v", reply.Keyboard)
	}
	titleRow := reply.Keyboard[1]
	if len(titleRow) != 1 || titleRow[0].Text != "JVM 已使用記憶體" || titleRow[0].Data != "noop" {
		t.Fatalf("unexpected title row: %#v", titleRow)
	}
	operationRow := reply.Keyboard[2]
	if len(operationRow) != 2 || operationRow[0].Text != "啟用規則" || operationRow[1].Text != "刪除" {
		t.Fatalf("unexpected operation row: %#v", operationRow)
	}
}

func TestEnabledRuleMenuShowsDisableAndDeleteOnSameRow(t *testing.T) {
	repo := repository.NewMemoryRepository(nil)
	handler := NewHandler(repo)
	handler.config = config.NewManager("", config.Config{AlertRules: []config.AlertRule{{
		Key:        "high-cpu",
		MetricName: metric.ProcessCPUUsage,
		Enabled:    true,
	}}}, nil)

	reply := handler.ruleMenu()
	operationRow := reply.Keyboard[2]
	if len(operationRow) != 2 ||
		operationRow[0].Text != "停用規則" ||
		operationRow[0].Data != "rule_disable:high-cpu" ||
		operationRow[1].Text != "刪除" {
		t.Fatalf("unexpected operation row: %#v", operationRow)
	}
	noop := handler.Handle(context.Background(), "123", "", "noop")
	if !noop.Silent {
		t.Fatal("rule title must not perform an action")
	}
}

func TestServiceMenuUsesNameAndOperationRows(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{
		{Name: "enabled-service", Enabled: true},
		{Name: "disabled-service", Enabled: false},
	})
	reply := NewHandler(repo).serviceMenu()

	if len(reply.Keyboard) != 6 {
		t.Fatalf("unexpected keyboard rows: %#v", reply.Keyboard)
	}
	if reply.Keyboard[1][0].Text != "disabled-service" || reply.Keyboard[1][0].Data != "noop" {
		t.Fatalf("unexpected disabled service title row: %#v", reply.Keyboard[1])
	}
	if reply.Keyboard[2][0].Text != "啟用服務" || reply.Keyboard[2][1].Text != "刪除" {
		t.Fatalf("unexpected disabled service operation row: %#v", reply.Keyboard[2])
	}
	if reply.Keyboard[3][0].Text != "enabled-service" || reply.Keyboard[3][0].Data != "noop" {
		t.Fatalf("unexpected enabled service title row: %#v", reply.Keyboard[3])
	}
	if reply.Keyboard[4][0].Text != "停用服務" || reply.Keyboard[4][1].Text != "刪除" {
		t.Fatalf("unexpected enabled service operation row: %#v", reply.Keyboard[4])
	}
}

func TestEnableRuleWithCustomThresholdWritesConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	initial := `{
		"services": [{"name":"www-service","base_url":"http://localhost","enabled":true}],
		"alert_rules": [{
			"key":"high-cpu",
			"service_name":"www-service",
			"metric_name":"process_cpu_usage",
			"operator":">",
			"threshold":0.8,
			"enabled":false
		}]
	}`
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(repository.NewMemoryRepository(cfg.ModelServices()))
	handler.config = config.NewManager(path, cfg, nil)

	reply := handler.Handle(context.Background(), "123", "", "rule_enable:high-cpu")
	if !keyboardContains(reply.Keyboard, "rule_enable_json:high-cpu") ||
		!keyboardContains(reply.Keyboard, "rule_enable_custom:high-cpu") {
		t.Fatalf("unexpected enable menu: %#v", reply.Keyboard)
	}
	handler.Handle(context.Background(), "123", "", "rule_enable_custom:high-cpu")
	reply = handler.Handle(context.Background(), "123", "0.65", "")
	if !keyboardContains(reply.Keyboard, "confirm_rule_enable_custom") {
		t.Fatalf("expected custom threshold confirmation: %#v", reply.Keyboard)
	}
	reply = handler.Handle(context.Background(), "123", "", "confirm_rule_enable_custom")
	if reply.Text != "自訂門檻值已回寫 JSON，規則已啟用。" {
		t.Fatalf("unexpected reply: %s", reply.Text)
	}
	reloaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded.AlertRules[0].Enabled || reloaded.AlertRules[0].Threshold != 0.65 {
		t.Fatalf("custom threshold was not persisted: %#v", reloaded.AlertRules[0])
	}
}

func keyboardContains(keyboard [][]Button, callbackData string) bool {
	for _, row := range keyboard {
		for _, button := range row {
			if button.Data == callbackData {
				return true
			}
		}
	}
	return false
}
