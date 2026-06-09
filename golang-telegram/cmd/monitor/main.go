package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-springboot-monitor-bot/internal/alert"
	"golang-springboot-monitor-bot/internal/bot"
	"golang-springboot-monitor-bot/internal/collector"
	"golang-springboot-monitor-bot/internal/config"
	"golang-springboot-monitor-bot/internal/notifier"
	"golang-springboot-monitor-bot/internal/repository"
	"golang-springboot-monitor-bot/internal/scheduler"
)

func main() {
	configPath := flag.String("config", "configs/telegram.json", "JSON config file path")
	once := flag.Bool("once", false, "run checks once and exit")
	command := flag.String("command", "", "run a local bot command after one check, for example: /status")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo := repository.NewMemoryRepository(cfg.ModelServices())
	checker := collector.NewHealthChecker(time.Duration(cfg.App.HTTPTimeoutSeconds) * time.Second)
	alertEngine := alert.NewEngine(repo, alert.Thresholds{
		ResponseTimeWarning: time.Duration(cfg.App.ResponseTimeWarningMS) * time.Millisecond,
	}, metricRules(cfg.AlertRules))
	notifierService := buildNotifier(cfg)
	schedulerService := scheduler.New(repo, checker, alertEngine, notifierService, cfg.App.MetricsEnabled)
	botHandler := bot.NewHandler(repo)

	log.Printf("started %s with %d enabled service(s)", cfg.App.Name, len(repo.EnabledServices()))

	if *once || *command != "" {
		schedulerService.RunOnce(ctx)
		if *command != "" {
			log.Print(botHandler.HandleCommand(*command))
		}
		return
	}

	if cfg.Telegram.Enabled {
		poller := bot.NewTelegramPoller(
			cfg.Telegram.BotToken,
			cfg.Telegram.ChatID,
			time.Duration(cfg.App.HTTPTimeoutSeconds+30)*time.Second,
			botHandler,
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

func buildNotifier(cfg config.Config) notifier.Notifier {
	logNotifier := notifier.LogNotifier{}
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
