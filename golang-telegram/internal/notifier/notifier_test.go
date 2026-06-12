package notifier

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

type recordingNotifier struct {
	event model.AlertEvent
}

func (notifier *recordingNotifier) Notify(_ context.Context, event model.AlertEvent) error {
	notifier.event = event
	return nil
}

func TestTelegramNotifierFallbackPreservesAlertContext(t *testing.T) {
	fallback := &recordingNotifier{}
	notifier := NewTelegramNotifier("token", "chat", time.Second, fallback)
	notifier.client.Transport = roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("network unavailable")
	})
	event := model.AlertEvent{
		ServiceName: "order-service",
		RuleKey:     "high-cpu",
		Message:     "CPU is high",
	}

	if err := notifier.Notify(context.Background(), event); err == nil {
		t.Fatal("expected telegram request error")
	}
	if fallback.event.ServiceName != event.ServiceName || fallback.event.RuleKey != event.RuleKey {
		t.Fatalf("fallback lost alert context: %#v", fallback.event)
	}
}
