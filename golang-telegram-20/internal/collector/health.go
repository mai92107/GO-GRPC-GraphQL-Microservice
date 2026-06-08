package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

type HealthChecker struct {
	client *http.Client
}

type HealthResult struct {
	Service        model.Service
	Status         string
	HTTPStatusCode int
	ResponseTime   time.Duration
	Error          error
	CheckedAt      time.Time
}

func NewHealthChecker(timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		client: &http.Client{Timeout: timeout},
	}
}

func (checker *HealthChecker) Check(ctx context.Context, service model.Service) HealthResult {
	startedAt := time.Now()
	result := HealthResult{
		Service:   service,
		Status:    "UNKNOWN",
		CheckedAt: startedAt,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, service.BaseURL+service.HealthPath, nil)
	if err != nil {
		result.Error = fmt.Errorf("build request: %w", err)
		return result
	}

	resp, err := checker.client.Do(req)
	result.ResponseTime = time.Since(startedAt)
	if err != nil {
		result.Status = "DOWN"
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	result.HTTPStatusCode = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		result.Status = "DOWN"
		result.Error = fmt.Errorf("unexpected http status: %d", resp.StatusCode)
		return result
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Status = "UNKNOWN"
		result.Error = fmt.Errorf("decode health response: %w", err)
		return result
	}

	if body.Status == "" {
		body.Status = "UNKNOWN"
	}
	result.Status = body.Status
	if body.Status != "UP" {
		result.Error = fmt.Errorf("service health status is %s", body.Status)
	}

	return result
}
