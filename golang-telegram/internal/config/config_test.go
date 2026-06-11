package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
		"services": [
			{
				"name": "order-service",
				"base_url": "http://localhost:8081/",
				"enabled": true
			}
		]
	}`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.App.Name != "golang-springboot-monitor-bot" {
		t.Fatalf("unexpected app name: %q", cfg.App.Name)
	}
	if cfg.App.DefaultCheckIntervalSeconds != 60 {
		t.Fatalf("unexpected default interval: %d", cfg.App.DefaultCheckIntervalSeconds)
	}

	service := cfg.Services[0]
	if service.BaseURL != "http://localhost:8081" {
		t.Fatalf("unexpected base url: %q", service.BaseURL)
	}
	if service.HealthPath != "/actuator/health" {
		t.Fatalf("unexpected health path: %q", service.HealthPath)
	}
	if service.MetricsPath != "/actuator/prometheus" {
		t.Fatalf("unexpected metrics path: %q", service.MetricsPath)
	}
	if service.CheckIntervalSeconds != 60 {
		t.Fatalf("unexpected service interval: %d", service.CheckIntervalSeconds)
	}
}

func TestLoadAllowsEmptyServices(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
		"app": {
			"name": "telegram-managed"
		},
		"services": []
	}`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err != nil {
		t.Fatalf("expected empty services config to be allowed: %v", err)
	}
}

func TestLoadAllowsWhenAllServicesAreDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
		"services": [
			{"name": "order-service", "base_url": "http://localhost:8081", "enabled": false},
			{"name": "payment-service", "base_url": "http://localhost:8082", "enabled": false}
		]
	}`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err != nil {
		t.Fatalf("expected config with no enabled services to be allowed: %v", err)
	}
}

func TestLoadValidatesEnabledCronResultNotify(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
		"cron_result_notify": {
			"enabled": true,
			"host": "https://notify.example.com",
			"path": "/cron/result"
		},
		"services": [
			{"name": "order-service", "base_url": "http://localhost:8081", "enabled": true}
		]
	}`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected enabled cron result notification without bearer token to be rejected")
	}
}

func TestLoadValidatesAlertRule(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
		"services": [
			{"name": "order-service", "base_url": "http://localhost:8081", "enabled": true}
		],
		"alert_rules": [
			{
				"key": "high-threads",
				"service_name": "order-service",
				"metric_name": "jvm_threads_live_threads",
				"operator": ">",
				"threshold": 100,
				"enabled": true
			}
		]
	}`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := Load(path); err != nil {
		t.Fatalf("load config: %v", err)
	}
}

func TestLoadRejectsUnsupportedMetricName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
		"services": [
			{"name": "order-service", "base_url": "http://localhost:8081", "enabled": true}
		],
		"alert_rules": [
			{
				"key": "typo-rule",
				"service_name": "order-service",
				"metric_name": "jvm_thread_live_thread",
				"operator": ">",
				"threshold": 100,
				"enabled": true
			}
		]
	}`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unsupported metric name to be rejected")
	}
}
