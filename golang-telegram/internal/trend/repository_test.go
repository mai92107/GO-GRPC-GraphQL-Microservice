package trend

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

func TestAddBuffersCurrentMinuteUntilFlush(t *testing.T) {
	repo := newTestRepository(t)
	now := time.Date(2026, 6, 12, 10, 46, 35, 0, time.UTC)
	repo.Add([]model.MetricSample{sample("www", "process_cpu_usage", 0.2, now)})

	assertPromFileCount(t, repo.directory, 0)
	if _, err := repo.Query("www", "process_cpu_usage", time.Hour, now); err == nil {
		t.Fatal("query must not include current minute buffer")
	}

	if err := repo.Flush(now.Truncate(time.Minute).Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	assertPromFileCount(t, repo.directory, 1)
	result, err := repo.Query("www", "process_cpu_usage", time.Hour, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Value != 0.2 {
		t.Fatalf("unexpected flushed result: %#v", result)
	}
}

func TestMinuteFileContainsAllServicesAndEscapedLabels(t *testing.T) {
	repo := newTestRepository(t)
	minute := time.Date(2026, 6, 12, 10, 46, 0, 0, time.UTC)
	repo.Add([]model.MetricSample{
		{ServiceName: "www", Name: "jvm_memory_used_bytes", Labels: map[string]string{"id": `G1 "Old" Gen`}, Value: 100, CollectedAt: minute.Add(5 * time.Second)},
		{ServiceName: "mgmt", Name: "process_cpu_usage", Value: 0.4, CollectedAt: minute.Add(10 * time.Second)},
	})
	if err := repo.Flush(minute.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(repo.minutePath(minute))
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if !strings.Contains(text, `monitor_service="www"`) || !strings.Contains(text, `monitor_service="mgmt"`) {
		t.Fatalf("minute file must contain all services: %s", text)
	}
	if !strings.Contains(text, `id="G1 \"Old\" Gen"`) {
		t.Fatalf("label was not escaped: %s", text)
	}
}

func TestCloseWritesPartialMinuteAndMergesWithoutDuplicates(t *testing.T) {
	repo := newTestRepository(t)
	now := time.Now().UTC().Truncate(time.Minute).Add(5 * time.Second)
	value := sample("www", "process_cpu_usage", 0.2, now)
	repo.Add([]model.MetricSample{value})
	if err := repo.Close(); err != nil {
		t.Fatal(err)
	}

	restarted, err := NewRepository(repo.directory, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	restarted.Add([]model.MetricSample{value, sample("www", "process_cpu_usage", 0.3, now.Add(5*time.Second))})
	if err := restarted.Close(); err != nil {
		t.Fatal(err)
	}
	samples, err := readPromFile(restarted.minutePath(now.Truncate(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected duplicate sample to be removed, got %#v", samples)
	}
}

func TestQuerySupportsAllRangesAndKeepsLabelSeriesSeparate(t *testing.T) {
	repo := newTestRepository(t)
	now := time.Date(2026, 6, 12, 10, 46, 30, 0, time.UTC)
	writeHistoricalSamples(t, repo, []model.MetricSample{
		{ServiceName: "order", Name: "jvm_memory_used_bytes", Labels: map[string]string{"area": "heap"}, Value: 100, CollectedAt: now.Add(-23 * time.Hour)},
		{ServiceName: "order", Name: "jvm_memory_used_bytes", Labels: map[string]string{"area": "nonheap"}, Value: 20, CollectedAt: now.Add(-30 * time.Minute)},
	})

	for _, duration := range []time.Duration{time.Hour, 4 * time.Hour, 8 * time.Hour, 16 * time.Hour, 24 * time.Hour} {
		result, err := repo.Query("order", "jvm_memory_used_bytes", duration, now)
		if err != nil {
			t.Fatalf("query %s: %v", duration, err)
		}
		if duration == 24*time.Hour && len(result) != 2 {
			t.Fatalf("24h query must include both series: %#v", result)
		}
	}
}

func TestQueryDownsamplesEachSeries(t *testing.T) {
	repo := newTestRepository(t)
	now := time.Date(2026, 6, 12, 10, 46, 30, 0, time.UTC)
	samples := make([]model.MetricSample, 0, 1200)
	for i := 0; i < 600; i++ {
		at := now.Add(time.Duration(i-600) * time.Second)
		samples = append(samples,
			model.MetricSample{ServiceName: "order", Name: "jvm_memory_used_bytes", Labels: map[string]string{"area": "heap"}, Value: float64(i), CollectedAt: at},
			model.MetricSample{ServiceName: "order", Name: "jvm_memory_used_bytes", Labels: map[string]string{"area": "nonheap"}, Value: float64(i), CollectedAt: at},
		)
	}
	writeHistoricalSamples(t, repo, samples)
	result, err := repo.Query("order", "jvm_memory_used_bytes", time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1000 {
		t.Fatalf("expected 500 samples per series, got %d", len(result))
	}
}

func TestCleanupKeepsFileOverlappingRetentionBoundary(t *testing.T) {
	repo := newTestRepository(t)
	now := time.Date(2026, 6, 12, 10, 46, 30, 0, time.UTC)
	overlapping := now.Add(-24 * time.Hour).Truncate(time.Minute)
	expired := overlapping.Add(-time.Minute)
	for _, minute := range []time.Time{overlapping, expired} {
		if err := os.WriteFile(repo.minutePath(minute), []byte{}, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if err := repo.cleanup(now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(repo.minutePath(overlapping)); err != nil {
		t.Fatal("file overlapping 24h boundary must be retained")
	}
	if _, err := os.Stat(repo.minutePath(expired)); !os.IsNotExist(err) {
		t.Fatal("fully expired file must be removed")
	}
}

func TestPendingBufferLimitDropsOldestMinute(t *testing.T) {
	repo := newTestRepository(t)
	current := time.Date(2026, 6, 12, 10, 46, 0, 0, time.UTC)
	repo.mu.Lock()
	for i := 6; i >= 1; i-- {
		at := current.Add(-time.Duration(i) * time.Minute)
		repo.buffers[at] = []model.MetricSample{sample("www", "process_cpu_usage", float64(i), at)}
	}
	repo.enforceBufferLimitLocked(current)
	repo.mu.Unlock()

	repo.mu.RLock()
	defer repo.mu.RUnlock()
	if len(repo.buffers) != maxPendingMinutes {
		t.Fatalf("expected %d pending buffers, got %d", maxPendingMinutes, len(repo.buffers))
	}
	if _, ok := repo.buffers[current.Add(-6*time.Minute)]; ok {
		t.Fatal("oldest pending minute must be dropped")
	}
}

func TestFlushFailureKeepsBufferForRetry(t *testing.T) {
	repo := newTestRepository(t)
	minute := time.Now().UTC().Truncate(time.Minute).Add(-time.Minute)
	repo.Add([]model.MetricSample{sample("www", "process_cpu_usage", 0.2, minute)})

	originalDirectory := repo.directory
	repo.directory = filepath.Join(originalDirectory, "missing", "trends")
	if err := repo.Flush(minute.Add(time.Minute)); err == nil {
		t.Fatal("expected flush failure")
	}
	if len(repo.buffers[minute]) != 1 {
		t.Fatal("failed minute must remain buffered for retry")
	}

	repo.directory = originalDirectory
	if err := repo.Flush(minute.Add(time.Minute)); err != nil {
		t.Fatalf("retry flush: %v", err)
	}
	if _, ok := repo.buffers[minute]; ok {
		t.Fatal("successful retry must remove buffered minute")
	}
}

func TestDownsampleToOneKeepsLatestSample(t *testing.T) {
	samples := []model.MetricSample{{Value: 1}, {Value: 2}}
	result := Downsample(samples, 1)
	if len(result) != 1 || result[0].Value != 2 {
		t.Fatalf("expected latest sample, got %#v", result)
	}
}

func BenchmarkQuery24Hours(b *testing.B) {
	repo, err := NewRepository(b.TempDir(), 24*time.Hour)
	if err != nil {
		b.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Minute)
	for minute := now.Add(-24 * time.Hour); minute.Before(now); minute = minute.Add(time.Minute) {
		var samples []model.MetricSample
		for second := 0; second < 60; second += 5 {
			samples = append(samples, sample("www", "process_cpu_usage", 0.2, minute.Add(time.Duration(second)*time.Second)))
		}
		if err := repo.writeMinute(minute, samples); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repo.Query("www", "process_cpu_usage", 24*time.Hour, now); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFlushMinute(b *testing.B) {
	minute := time.Now().UTC().Truncate(time.Minute).Add(-time.Minute)
	samples := make([]model.MetricSample, 0, 600)
	for series := 0; series < 50; series++ {
		for second := 0; second < 60; second += 5 {
			samples = append(samples, model.MetricSample{
				ServiceName: "www",
				Name:        "jvm_memory_used_bytes",
				Labels:      map[string]string{"id": strconv.Itoa(series)},
				Value:       float64(series),
				CollectedAt: minute.Add(time.Duration(second) * time.Second),
			})
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo, err := NewRepository(b.TempDir(), 24*time.Hour)
		if err != nil {
			b.Fatal(err)
		}
		repo.Add(samples)
		if err := repo.Flush(minute.Add(time.Minute)); err != nil {
			b.Fatal(err)
		}
	}
}

func newTestRepository(t testing.TB) *Repository {
	t.Helper()
	repo, err := NewRepository(t.TempDir(), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

func sample(service, name string, value float64, at time.Time) model.MetricSample {
	return model.MetricSample{ServiceName: service, Name: name, Value: value, CollectedAt: at}
}

func assertPromFileCount(t *testing.T, directory string, expected int) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(directory, "*.prom"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != expected {
		t.Fatalf("expected %d prom files, got %d", expected, len(paths))
	}
}

func writeHistoricalSamples(t *testing.T, repo *Repository, samples []model.MetricSample) {
	t.Helper()
	byMinute := make(map[time.Time][]model.MetricSample)
	for _, value := range samples {
		minute := value.CollectedAt.UTC().Truncate(time.Minute)
		byMinute[minute] = append(byMinute[minute], value)
	}
	for minute, values := range byMinute {
		if err := repo.writeMinute(minute, values); err != nil {
			t.Fatal(err)
		}
	}
}
