package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/parser"
)

type MetricsResult struct {
	Service      model.Service
	Snapshots    []model.MetricSnapshot
	ResponseTime time.Duration
	Error        error
	CollectedAt  time.Time
}

func (checker *HealthChecker) CollectMetrics(ctx context.Context, service model.Service) MetricsResult {
	startedAt := time.Now()
	result := MetricsResult{
		Service:     service,
		CollectedAt: startedAt,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, service.BaseURL+service.MetricsPath, nil)
	if err != nil {
		result.Error = fmt.Errorf("build metrics request: %w", err)
		return result
	}

	resp, err := checker.client.Do(req)
	result.ResponseTime = time.Since(startedAt)
	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		result.Error = fmt.Errorf("unexpected metrics http status: %d", resp.StatusCode)
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("read metrics response: %w", err)
		return result
	}

	snapshots, err := parser.ParsePrometheus(service.Name, string(body))
	if err != nil {
		result.Error = fmt.Errorf("parse metrics: %w", err)
		return result
	}
	for i := range snapshots {
		snapshots[i].CollectedAt = startedAt
	}
	result.Snapshots = snapshots
	return result
}
