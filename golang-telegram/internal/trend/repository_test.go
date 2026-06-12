package trend

import (
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

func TestQueryFiltersRangeAndDownsamples(t *testing.T) {
	repo := NewRepository()
	now := time.Now()
	samples := make([]model.MetricSample, 0, 700)
	for i := 0; i < 700; i++ {
		samples = append(samples, model.MetricSample{ServiceName: "order", Name: "process_cpu_usage", Value: float64(i), CollectedAt: now.Add(time.Duration(i-699) * time.Minute)})
	}
	repo.Add(samples)
	result, err := repo.Query("order", "process_cpu_usage", 8*time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) > 500 {
		t.Fatalf("expected at most 500 samples, got %d", len(result))
	}
	for _, sample := range result {
		if sample.CollectedAt.Before(now.Add(-8 * time.Hour)) {
			t.Fatal("query returned expired sample")
		}
	}
}

func TestQueryKeepsLabelSeriesSeparate(t *testing.T) {
	repo := NewRepository()
	now := time.Now()
	repo.Add([]model.MetricSample{
		{ServiceName: "order", Name: "jvm_memory_used_bytes", Labels: map[string]string{"area": "heap"}, Value: 100, CollectedAt: now},
		{ServiceName: "order", Name: "jvm_memory_used_bytes", Labels: map[string]string{"area": "nonheap"}, Value: 20, CollectedAt: now},
	})
	result, err := repo.Query("order", "jvm_memory_used_bytes", time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("expected two separate label series, got %#v", result)
	}
	values := map[string]float64{}
	for _, sample := range result {
		values[sample.Labels["area"]] = sample.Value
	}
	if values["heap"] != 100 || values["nonheap"] != 20 {
		t.Fatalf("unexpected separate series: %#v", result)
	}
}

func TestDownsampleToOneKeepsLatestSample(t *testing.T) {
	samples := []model.MetricSample{{Value: 1}, {Value: 2}}

	result := Downsample(samples, 1)

	if len(result) != 1 || result[0].Value != 2 {
		t.Fatalf("expected latest sample, got %#v", result)
	}
}
