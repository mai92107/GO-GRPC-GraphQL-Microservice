package notifier

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

type recordingNotifier struct {
	event model.AlertEvent
}

func TestTelegramNotifierSendsPhotoForImageAlert(t *testing.T) {
	notifier := NewTelegramNotifier("token", "chat", time.Second, nil)
	var path, contentType, body string
	notifier.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		path = req.URL.Path
		contentType = req.Header.Get("Content-Type")
		data, _ := io.ReadAll(req.Body)
		body = string(data)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}, nil
	})
	event := model.AlertEvent{Message: "CPU is high", ImagePNG: []byte("png")}
	if err := notifier.Notify(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "/sendPhoto") || !strings.HasPrefix(contentType, "multipart/form-data") || !strings.Contains(body, "CPU is high") {
		t.Fatalf("unexpected photo request: path=%q content-type=%q body=%q", path, contentType, body)
	}
}

func TestTelegramNotifierFallsBackToTextWhenPhotoFails(t *testing.T) {
	notifier := NewTelegramNotifier("token", "chat", time.Second, nil)
	var paths []string
	notifier.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		paths = append(paths, req.URL.Path)
		if strings.HasSuffix(req.URL.Path, "/sendPhoto") {
			return nil, errors.New("photo unavailable")
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}, nil
	})
	err := notifier.Notify(context.Background(), model.AlertEvent{Message: "CPU is high", ImagePNG: []byte("png")})
	if err == nil {
		t.Fatal("expected photo failure to be reported")
	}
	if len(paths) != 2 || !strings.HasSuffix(paths[1], "/sendMessage") {
		t.Fatalf("expected text fallback after photo failure, got %#v", paths)
	}
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
