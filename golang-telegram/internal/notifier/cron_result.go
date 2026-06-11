package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

type CronResultNotifyRequest struct {
	ServiceName string            `json:"serviceName"`
	RuleKey     string            `json:"ruleKey"`
	Severity    string            `json:"severity"`
	Status      model.AlertStatus `json:"status"`
	Message     string            `json:"message"`
	StartedAt   time.Time         `json:"startedAt"`
	ResolvedAt  *time.Time        `json:"resolvedAt,omitempty"`
}

type CronResultNotifyResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}

type CronResultNotifier struct {
	url                 string
	bearerToken         string
	successResponseCode string
	client              *http.Client
}

func NewCronResultNotifier(host, path, bearerToken, successResponseCode string, timeout time.Duration) *CronResultNotifier {
	return &CronResultNotifier{
		url:                 strings.TrimRight(host, "/") + "/" + strings.TrimLeft(path, "/"),
		bearerToken:         bearerToken,
		successResponseCode: successResponseCode,
		client:              &http.Client{Timeout: timeout},
	}
}

func (notifier *CronResultNotifier) Notify(ctx context.Context, event model.AlertEvent) error {
	payload := CronResultNotifyRequest{
		ServiceName: event.ServiceName,
		RuleKey:     event.RuleKey,
		Severity:    event.Severity,
		Status:      event.Status,
		Message:     event.Message,
		StartedAt:   event.StartedAt,
		ResolvedAt:  event.ResolvedAt,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal cron result notification: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, notifier.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build cron result notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+notifier.bearerToken)

	resp, err := notifier.client.Do(req)
	if err != nil {
		return fmt.Errorf("cron result notification request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read cron result notification response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("cron result notification failed: http %d", resp.StatusCode)
	}

	var result CronResultNotifyResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return fmt.Errorf("parse cron result notification response: %w", err)
	}
	if result.ResponseCode != notifier.successResponseCode {
		return fmt.Errorf(
			"cron result notification failed: response_code=%s response_message=%s",
			result.ResponseCode,
			result.ResponseMessage,
		)
	}
	return nil
}
