package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang-springboot-monitor-bot/internal/applog"
	"golang-springboot-monitor-bot/internal/model"
)

type Notifier interface {
	Notify(context.Context, model.AlertEvent) error
}

type LogNotifier struct{ Logger applog.ApplicationLogger }

func (notifier LogNotifier) Notify(ctx context.Context, event model.AlertEvent) error {
	if notifier.Logger != nil {
		notifier.Logger.Info(ctx, "notification_dispatched",
			applog.Field{Key: "service", Value: event.ServiceName},
			applog.Field{Key: "severity", Value: event.Severity},
			applog.Field{Key: "status", Value: event.Status},
		)
	}
	return nil
}

type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
	fallback Notifier
}

func NewTelegramNotifier(botToken, chatID string, timeout time.Duration, fallback Notifier) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: timeout},
		fallback: fallback,
	}
}

func (notifier *TelegramNotifier) Notify(ctx context.Context, event model.AlertEvent) error {
	payload := map[string]string{
		"chat_id": notifier.chatID,
		"text":    event.Message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", notifier.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := notifier.client.Do(req)
	if err != nil {
		if notifier.fallback != nil {
			_ = notifier.fallback.Notify(ctx, event)
		}
		return fmt.Errorf("telegram sendMessage request failed: %w", sanitizeTelegramError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if notifier.fallback != nil {
			_ = notifier.fallback.Notify(ctx, event)
		}
		return fmt.Errorf("telegram sendMessage failed: http %d", resp.StatusCode)
	}
	return nil
}

type MultiNotifier struct {
	notifiers []Notifier
}

func NewMulti(notifiers ...Notifier) MultiNotifier {
	return MultiNotifier{notifiers: notifiers}
}

func (notifier MultiNotifier) Notify(ctx context.Context, event model.AlertEvent) error {
	var lastErr error
	for _, child := range notifier.notifiers {
		if child == nil {
			continue
		}
		if err := child.Notify(ctx, event); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

type sanitizedTelegramError struct {
	message string
}

func (err sanitizedTelegramError) Error() string {
	return err.message
}

func sanitizeTelegramError(err error) error {
	if err == nil {
		return nil
	}
	return sanitizedTelegramError{message: "request failed"}
}
