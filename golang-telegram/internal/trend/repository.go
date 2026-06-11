package trend

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

type Repository struct {
	mu        sync.RWMutex
	retention time.Duration
	series    map[string][]model.MetricSample
}

func NewRepository() *Repository {
	return &Repository{retention: 8 * time.Hour, series: make(map[string][]model.MetricSample)}
}

func (repo *Repository) Add(samples []model.MetricSample) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	cutoff := time.Now().Add(-repo.retention)
	for _, sample := range samples {
		key := seriesKey(sample.ServiceName, sample.Name, sample.Labels)
		values := append(repo.series[key], sample)
		first := sort.Search(len(values), func(i int) bool { return !values[i].CollectedAt.Before(cutoff) })
		repo.series[key] = append([]model.MetricSample(nil), values[first:]...)
	}
}

func (repo *Repository) Query(serviceName, metricName string, duration time.Duration, now time.Time) ([]model.MetricSample, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	if duration != time.Hour && duration != 4*time.Hour && duration != 8*time.Hour {
		return nil, fmt.Errorf("time range must be 1h, 4h, or 8h")
	}
	prefix := serviceName + "|" + metricName + "|"
	var matching [][]model.MetricSample
	for key, samples := range repo.series {
		if strings.HasPrefix(key, prefix) {
			matching = append(matching, samples)
		}
	}
	if len(matching) == 0 {
		return nil, fmt.Errorf("no trend data for service %q metric %q", serviceName, metricName)
	}
	cutoff := now.Add(-duration)
	aggregated := make(map[int64]model.MetricSample)
	for _, series := range matching {
		for _, sample := range series {
			if sample.CollectedAt.Before(cutoff) || sample.CollectedAt.After(now) {
				continue
			}
			key := sample.CollectedAt.UnixNano()
			existing, ok := aggregated[key]
			if !ok {
				sample.Labels = nil
				aggregated[key] = sample
				continue
			}
			if strings.HasSuffix(metricName, "_max") {
				if sample.Value > existing.Value {
					existing.Value = sample.Value
				}
			} else {
				existing.Value += sample.Value
			}
			aggregated[key] = existing
		}
	}
	result := make([]model.MetricSample, 0, len(aggregated))
	for _, sample := range aggregated {
		result = append(result, sample)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CollectedAt.Before(result[j].CollectedAt) })
	if len(result) == 0 {
		return nil, fmt.Errorf("no samples in the requested %s range", duration)
	}
	return Downsample(result, 500), nil
}

func Downsample(samples []model.MetricSample, max int) []model.MetricSample {
	if len(samples) <= max || max <= 0 {
		return append([]model.MetricSample(nil), samples...)
	}
	result := make([]model.MetricSample, 0, max)
	step := float64(len(samples)-1) / float64(max-1)
	for i := 0; i < max; i++ {
		result = append(result, samples[int(float64(i)*step)])
	}
	return result
}

func seriesKey(serviceName, metricName string, labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var values []string
	for _, key := range keys {
		values = append(values, key+"="+labels[key])
	}
	return serviceName + "|" + metricName + "|" + strings.Join(values, ",")
}
