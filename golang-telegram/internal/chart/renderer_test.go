package chart

import (
	"bytes"
	"image/png"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

func TestRendererProducesPNG(t *testing.T) {
	now := time.Now()
	var output bytes.Buffer
	err := NewRenderer().Render(&output, "order-service", "process_cpu_usage", "1h", []model.MetricSample{
		{Value: 0.2, CollectedAt: now.Add(-time.Minute)},
		{Value: 0.4, CollectedAt: now},
	}, []float64{0.8})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(bytes.NewReader(output.Bytes())); err != nil {
		t.Fatalf("invalid PNG: %v", err)
	}
}

func TestMetricScaleUsesReadableYAxisUnits(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      float64
		unit       string
		factor     float64
	}{
		{name: "CPU percentage", metricName: "process_cpu_usage", value: 0.5, unit: "百分比（%）", factor: 100},
		{name: "system CPU percentage", metricName: "system_cpu_usage", value: 0.5, unit: "百分比（%）", factor: 100},
		{name: "system CPU count", metricName: "system_cpu_count", value: 8, unit: "CPU 核心數", factor: 1},
		{name: "memory GB", metricName: "jvm_memory_used_bytes", value: 2 * 1024 * 1024 * 1024, unit: "GB", factor: 1.0 / (1024 * 1024 * 1024)},
		{name: "milliseconds", metricName: "http_server_requests_seconds_max", value: 0.25, unit: "毫秒", factor: 1000},
		{name: "seconds", metricName: "http_server_requests_seconds_max", value: 2, unit: "秒", factor: 1},
		{name: "count", metricName: "http_server_requests_seconds_count", value: 10, unit: "次數", factor: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scale := metricScaleFor(test.metricName, []model.MetricSample{{Value: test.value}})
			if scale.unit != test.unit || scale.factor != test.factor {
				t.Fatalf("unexpected scale: %#v", scale)
			}
		})
	}
}

func TestDataYRangeUsesActualDataMaximum(t *testing.T) {
	yRange := dataYRange([]float64{0.01, 0.03, 0.05})
	if yRange.Min != 0.01 || yRange.Max != 0.05 {
		t.Fatalf("unexpected Y range: min=%g max=%g", yRange.Min, yRange.Max)
	}
}

func TestDataYRangeKeepsConstantPositiveMaximum(t *testing.T) {
	yRange := dataYRange([]float64{0.05, 0.05})
	if yRange.Min != 0 || yRange.Max != 0.05 {
		t.Fatalf("unexpected constant Y range: min=%g max=%g", yRange.Min, yRange.Max)
	}
}

func TestTrendTimeFormatterIncludesTime(t *testing.T) {
	at := time.Date(2026, 6, 11, 15, 30, 45, 0, time.FixedZone("Asia/Taipei", 8*60*60))
	if got := trendTimeFormatter("1h")(at); got != "15:30:45" {
		t.Fatalf("unexpected 1h formatter output: %q", got)
	}
	if got := trendTimeFormatter("8h")(at); got != "06-11 15:30" {
		t.Fatalf("unexpected 8h formatter output: %q", got)
	}
}
