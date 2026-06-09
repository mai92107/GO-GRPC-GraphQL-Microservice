package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang-springboot-monitor-bot/internal/metric"
	"golang-springboot-monitor-bot/internal/model"
)

type Config struct {
	App              AppConfig              `json:"app"`
	Telegram         TelegramConfig         `json:"telegram"`
	CronResultNotify CronResultNotifyConfig `json:"cron_result_notify"`
	Services         []Service              `json:"services"`
	AlertRules       []AlertRule            `json:"alert_rules"`
}

type AppConfig struct {
	Name                        string `json:"name"`
	DefaultCheckIntervalSeconds int    `json:"default_check_interval_seconds"`
	HTTPTimeoutSeconds          int    `json:"http_timeout_seconds"`
	MetricsEnabled              bool   `json:"metrics_enabled"`
	ResponseTimeWarningMS       int    `json:"response_time_warning_ms"`
}

type TelegramConfig struct {
	Enabled  bool   `json:"enabled"`
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

type CronResultNotifyConfig struct {
	Enabled             bool   `json:"enabled"`
	Host                string `json:"host"`
	Path                string `json:"path"`
	BearerToken         string `json:"bearer_token"`
	SuccessResponseCode string `json:"success_response_code"`
}

type Service struct {
	Name                 string `json:"name"`
	Environment          string `json:"environment"`
	BaseURL              string `json:"base_url"`
	HealthPath           string `json:"health_path"`
	MetricsPath          string `json:"metrics_path"`
	CheckIntervalSeconds int    `json:"check_interval_seconds"`
	Enabled              bool   `json:"enabled"`
}

type AlertRule struct {
	Key         string            `json:"key"`
	ServiceName string            `json:"service_name"`
	MetricName  metric.Name       `json:"metric_name"`
	Labels      map[string]string `json:"labels"`
	Operator    string            `json:"operator"`
	Threshold   float64           `json:"threshold"`
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	Enabled     bool              `json:"enabled"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config json: %w", err)
	}

	applyDefaults(&cfg)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) ModelServices() []model.Service {
	services := make([]model.Service, 0, len(cfg.Services))
	for _, service := range cfg.Services {
		services = append(services, model.Service{
			Name:                 service.Name,
			Environment:          service.Environment,
			BaseURL:              service.BaseURL,
			HealthPath:           service.HealthPath,
			MetricsPath:          service.MetricsPath,
			CheckIntervalSeconds: service.CheckIntervalSeconds,
			Enabled:              service.Enabled,
		})
	}
	return services
}

func applyDefaults(cfg *Config) {
	if cfg.App.Name == "" {
		cfg.App.Name = "golang-springboot-monitor-bot"
	}
	if cfg.App.DefaultCheckIntervalSeconds <= 0 {
		cfg.App.DefaultCheckIntervalSeconds = 60
	}
	if cfg.App.HTTPTimeoutSeconds <= 0 {
		cfg.App.HTTPTimeoutSeconds = 5
	}
	if cfg.App.ResponseTimeWarningMS <= 0 {
		cfg.App.ResponseTimeWarningMS = 2000
	}
	if cfg.CronResultNotify.SuccessResponseCode == "" {
		cfg.CronResultNotify.SuccessResponseCode = "SUCCESS"
	}
	cfg.CronResultNotify.Host = strings.TrimRight(cfg.CronResultNotify.Host, "/")
	for i := range cfg.Services {
		if cfg.Services[i].Environment == "" {
			cfg.Services[i].Environment = "production"
		}
		if cfg.Services[i].HealthPath == "" {
			cfg.Services[i].HealthPath = "/actuator/health"
		}
		if cfg.Services[i].MetricsPath == "" {
			cfg.Services[i].MetricsPath = "/actuator/prometheus"
		}
		if cfg.Services[i].CheckIntervalSeconds <= 0 {
			cfg.Services[i].CheckIntervalSeconds = cfg.App.DefaultCheckIntervalSeconds
		}
		cfg.Services[i].BaseURL = strings.TrimRight(cfg.Services[i].BaseURL, "/")
	}

	for i := range cfg.AlertRules {
		if cfg.AlertRules[i].ServiceName == "" {
			cfg.AlertRules[i].ServiceName = "*"
		}
		if cfg.AlertRules[i].Severity == "" {
			cfg.AlertRules[i].Severity = "warning"
		}
	}
}

func validate(cfg Config) error {
	if len(cfg.Services) == 0 {
		return errors.New("config must contain at least one service")
	}

	seen := make(map[string]bool, len(cfg.Services))
	enabledCount := 0
	for _, service := range cfg.Services {
		if service.Name == "" {
			return errors.New("service name is required")
		}
		if service.BaseURL == "" {
			return fmt.Errorf("service %q base_url is required", service.Name)
		}
		if seen[service.Name] {
			return fmt.Errorf("service %q is duplicated", service.Name)
		}
		seen[service.Name] = true
		if service.Enabled {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		return errors.New("config must contain at least one enabled service; set services[].enabled to true")
	}

	ruleKeys := make(map[string]bool, len(cfg.AlertRules))
	for i, rule := range cfg.AlertRules {
		if !rule.Enabled {
			continue
		}
		if rule.Key == "" {
			return fmt.Errorf("alert_rules[%d].key is required", i)
		}
		if ruleKeys[rule.Key] {
			return fmt.Errorf("alert rule key %q is duplicated", rule.Key)
		}
		ruleKeys[rule.Key] = true
		if rule.MetricName == "" {
			return fmt.Errorf("alert rule %q metric_name is required", rule.Key)
		}
		if err := metric.Validate(rule.MetricName); err != nil {
			return fmt.Errorf("alert rule %q: %w", rule.Key, err)
		}
		if !validOperator(rule.Operator) {
			return fmt.Errorf("alert rule %q has invalid operator %q", rule.Key, rule.Operator)
		}
		if rule.ServiceName != "*" && !seen[rule.ServiceName] {
			return fmt.Errorf("alert rule %q references unknown service %q", rule.Key, rule.ServiceName)
		}
	}

	if cfg.Telegram.Enabled && (cfg.Telegram.BotToken == "" || cfg.Telegram.ChatID == "") {
		return errors.New("telegram.bot_token and telegram.chat_id are required when telegram.enabled is true")
	}
	if cfg.CronResultNotify.Enabled {
		if cfg.CronResultNotify.Host == "" || cfg.CronResultNotify.Path == "" || cfg.CronResultNotify.BearerToken == "" {
			return errors.New("cron_result_notify.host, path, and bearer_token are required when cron_result_notify.enabled is true")
		}
	}

	return nil
}

func validOperator(operator string) bool {
	switch operator {
	case ">", ">=", "<", "<=", "==", "!=":
		return true
	default:
		return false
	}
}
