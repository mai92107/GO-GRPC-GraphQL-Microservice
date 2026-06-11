package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-springboot-monitor-bot/internal/alert"
	"golang-springboot-monitor-bot/internal/applog"
	"golang-springboot-monitor-bot/internal/bot"
	"golang-springboot-monitor-bot/internal/chart"
	"golang-springboot-monitor-bot/internal/collector"
	"golang-springboot-monitor-bot/internal/config"
	"golang-springboot-monitor-bot/internal/notifier"
	"golang-springboot-monitor-bot/internal/repository"
	"golang-springboot-monitor-bot/internal/scheduler"
	"golang-springboot-monitor-bot/internal/trend"
)

func main() {
	configPath := flag.String("config", "configs/demo.json", "JSON config file path")
	once := flag.Bool("once", false, "run checks once and exit")
	command := flag.String("command", "", "run a local bot command after one check, for example: /status")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR %s config load config: %v\n", time.Now().Format(time.RFC3339), err)
		return
	}
	terminal := applog.NewTerminalWriter(os.Stdout, cfg.Logging.Timezone)
	fileLogger, err := applog.New(cfg.Logging, terminal)
	var appLogger applog.ApplicationLogger = fileLogger
	if err != nil {
		terminal.Error("logger", err)
		appLogger = applog.NewFallback(terminal)
	}
	defer appLogger.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo := repository.NewMemoryRepository(cfg.ModelServices())
	trendRepo := trend.NewRepository()
	checker := collector.NewHealthChecker(time.Duration(cfg.App.HTTPTimeoutSeconds) * time.Second)
	alertEngine := alert.NewEngine(repo, alert.Thresholds{
		ResponseTimeWarning: time.Duration(cfg.App.ResponseTimeWarningMS) * time.Millisecond,
	}, metricRules(cfg.AlertRules))
	notifierService := buildNotifier(cfg, appLogger)
	schedulerService := scheduler.New(repo, checker, alertEngine, notifierService, cfg.App.MetricsEnabled, trendRepo, appLogger, terminal)
	configManager := config.NewManager(*configPath, cfg, func(updated config.Config) {
		repo.ReplaceServices(updated.ModelServices())
		alertEngine.ReplaceRules(alert.Thresholds{ResponseTimeWarning: time.Duration(updated.App.ResponseTimeWarningMS) * time.Millisecond}, metricRules(updated.AlertRules))
		schedulerService.SetMetricsEnabled(updated.App.MetricsEnabled)
		appLogger.Info(ctx, "configuration_applied")
	})
	botHandler := bot.NewManagementHandler(repo, trendRepo, chart.NewRenderer(), configManager, schedulerService.CheckNow, schedulerService.RunOnce, appLogger)

	appLogger.Info(ctx, "application_started", applog.Field{Key: "enabled_services", Value: len(repo.EnabledServices())})

	if *once || *command != "" {
		schedulerService.RunOnce(ctx)
		if *command != "" {
			fmt.Fprintln(os.Stdout, botHandler.HandleCommand(*command))
		}
		return
	}

	if cfg.Telegram.Enabled {
		poller := bot.NewTelegramPoller(
			cfg.Telegram.BotToken,
			cfg.Telegram.ChatID,
			time.Duration(cfg.App.HTTPTimeoutSeconds+30)*time.Second,
			botHandler,
			appLogger,
		)
		go poller.Run(ctx)
	}

	schedulerService.Run(ctx)
}

func metricRules(rules []config.AlertRule) []alert.MetricRule {
	result := make([]alert.MetricRule, 0, len(rules))
	for _, rule := range rules {
		result = append(result, alert.MetricRule{
			Key:         rule.Key,
			ServiceName: rule.ServiceName,
			MetricName:  rule.MetricName,
			Labels:      rule.Labels,
			Operator:    rule.Operator,
			Threshold:   rule.Threshold,
			Severity:    rule.Severity,
			Message:     rule.Message,
			Enabled:     rule.Enabled,
		})
	}
	return result
}

func buildNotifier(cfg config.Config, logger applog.ApplicationLogger) notifier.Notifier {
	logNotifier := notifier.LogNotifier{Logger: logger}
	notifiers := []notifier.Notifier{logNotifier}
	if cfg.Telegram.Enabled {
		notifiers = append(notifiers, notifier.NewTelegramNotifier(
			cfg.Telegram.BotToken,
			cfg.Telegram.ChatID,
			time.Duration(cfg.App.HTTPTimeoutSeconds)*time.Second,
			logNotifier,
		))
	}

	if cfg.CronResultNotify.Enabled {
		notifiers = append(notifiers, notifier.NewCronResultNotifier(
			cfg.CronResultNotify.Host,
			cfg.CronResultNotify.Path,
			cfg.CronResultNotify.BearerToken,
			cfg.CronResultNotify.SuccessResponseCode,
			time.Duration(cfg.App.HTTPTimeoutSeconds)*time.Second,
		))
	}
	return notifier.NewMulti(notifiers...)
}
