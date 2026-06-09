package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

func TestCronResultNotifierNotify(t *testing.T) {
	var received CronResultNotifyRequest
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", req.Method)
		}
		if req.URL.Host != "notify.example.com" {
			t.Errorf("unexpected host: %s", req.URL.Host)
		}
		if req.URL.Path != "/cron/result" {
			t.Errorf("unexpected path: %s", req.URL.Path)
		}
		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content type: %s", req.Header.Get("Content-Type"))
		}
		if req.Header.Get("Authorization") != "Bearer secret-token" {
			t.Errorf("unexpected authorization header: %s", req.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(req.Body).Decode(&received); err != nil {
			t.Errorf("decode request: %v", err)
		}
		return jsonResponse(`{"responseCode":"SUCCESS","responseMessage":"ok"}`), nil
	})

	startedAt := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	event := model.AlertEvent{
		ServiceName: "order-service",
		RuleKey:     "high-cpu",
		Severity:    "warning",
		Status:      model.AlertStatusOpen,
		Message:     "CPU is high",
		StartedAt:   startedAt,
	}
	notifier := NewCronResultNotifier("https://notify.example.com", "/cron/result", "secret-token", "SUCCESS", time.Second)
	notifier.client.Transport = transport

	if err := notifier.Notify(context.Background(), event); err != nil {
		t.Fatalf("notify: %v", err)
	}
	if received.ServiceName != event.ServiceName || received.RuleKey != event.RuleKey || received.Message != event.Message {
		t.Fatalf("unexpected request: %#v", received)
	}
}

func TestCronResultNotifierRejectsFailureResponseCode(t *testing.T) {
	notifier := NewCronResultNotifier("https://notify.example.com", "/cron/result", "secret-token", "SUCCESS", time.Second)
	notifier.client.Transport = roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return jsonResponse(`{"responseCode":"ERROR","responseMessage":"rejected"}`), nil
	})
	if err := notifier.Notify(context.Background(), model.AlertEvent{}); err == nil {
		t.Fatal("expected failure response code to return an error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
