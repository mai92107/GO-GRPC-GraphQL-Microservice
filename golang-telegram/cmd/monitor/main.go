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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewMemoryRepository(cfg.ModelServices())
	trendRepo, err := trend.NewRepository(cfg.TrendStorage.Directory, time.Duration(cfg.TrendStorage.RetentionHours)*time.Hour)
	if err != nil {
		appLogger.Error(ctx, "trend_repository_initialization_failed", err)
		return
	}
	defer func() {
		if err := trendRepo.Close(); err != nil {
			appLogger.Error(context.Background(), "trend_repository_close_failed", err)
		}
	}()
	go trendRepo.Run(ctx)
	go logTrendErrors(ctx, trendRepo, appLogger)
	checker := collector.NewHealthChecker(time.Duration(cfg.App.HTTPTimeoutSeconds) * time.Second)
	alertEngine := alert.NewEngine(repo, alert.Thresholds{
		ResponseTimeWarning: time.Duration(cfg.App.ResponseTimeWarningMS) * time.Millisecond,
	}, metricRules(cfg.AlertRules))
	notifierService, lifecycleNotifier := buildNotifier(cfg, appLogger)
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

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalChannel)
	shutdownReason := make(chan string, 1)
	go func() {
		sig := <-signalChannel
		shutdownReason <- signalReason(sig)
		cancel()
	}()

	sendLifecycleNotification(lifecycleNotifier, "系統重新上線拉", appLogger)

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
	reason := <-shutdownReason
	sendLifecycleNotification(lifecycleNotifier, "系統已被關閉\n原因："+reason, appLogger)
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

func buildNotifier(cfg config.Config, logger applog.ApplicationLogger) (notifier.Notifier, notifier.TextNotifier) {
	logNotifier := notifier.LogNotifier{Logger: logger}
	notifiers := []notifier.Notifier{logNotifier}
	var lifecycleNotifier notifier.TextNotifier
	if cfg.Telegram.Enabled {
		telegramNotifier := notifier.NewTelegramNotifier(
			cfg.Telegram.BotToken,
			cfg.Telegram.ChatID,
			time.Duration(cfg.App.HTTPTimeoutSeconds)*time.Second,
			logNotifier,
		)
		notifiers = append(notifiers, telegramNotifier)
		lifecycleNotifier = telegramNotifier
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
	return notifier.NewMulti(notifiers...), lifecycleNotifier
}

func sendLifecycleNotification(target notifier.TextNotifier, message string, logger applog.ApplicationLogger) {
	if target == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := target.SendText(ctx, message); err != nil {
		logger.Error(ctx, "lifecycle_notification_failed", err)
		return
	}
	logger.Info(ctx, "lifecycle_notification_sent", applog.Field{Key: "message", Value: message})
}

func signalReason(sig os.Signal) string {
	switch sig {
	case os.Interrupt:
		return "收到中斷訊號（SIGINT / Ctrl+C）"
	case syscall.SIGTERM:
		return "收到終止訊號（SIGTERM）"
	default:
		return fmt.Sprintf("收到系統訊號（%s）", sig)
	}
}

func logTrendErrors(ctx context.Context, repo *trend.Repository, logger applog.ApplicationLogger) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-repo.Errors():
			logger.Error(ctx, "trend_repository_failed", err)
		}
	}
}
