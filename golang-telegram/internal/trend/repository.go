package trend

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

const (
	fileTimeLayout    = "20060102T150400Z"
	serviceLabel      = "monitor_service"
	maxPendingMinutes = 5
	maxSeriesSamples  = 500
)

var supportedDurations = map[time.Duration]struct{}{
	time.Hour:      {},
	4 * time.Hour:  {},
	8 * time.Hour:  {},
	16 * time.Hour: {},
	24 * time.Hour: {},
}

type Repository struct {
	mu        sync.RWMutex
	fileMu    sync.RWMutex
	directory string
	retention time.Duration
	buffers   map[time.Time][]model.MetricSample
	errors    chan error
}

func NewRepository(directory string, retention time.Duration) (*Repository, error) {
	if strings.TrimSpace(directory) == "" {
		directory = "trends"
	}
	if retention <= 0 {
		retention = 24 * time.Hour
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create trend directory: %w", err)
	}
	repo := &Repository{
		directory: directory,
		retention: retention,
		buffers:   make(map[time.Time][]model.MetricSample),
		errors:    make(chan error, 16),
	}
	if err := repo.cleanup(time.Now()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (repo *Repository) Errors() <-chan error {
	return repo.errors
}

func (repo *Repository) Add(samples []model.MetricSample) {
	if len(samples) == 0 {
		return
	}
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, sample := range samples {
		minute := sample.CollectedAt.UTC().Truncate(time.Minute)
		repo.buffers[minute] = append(repo.buffers[minute], sample)
	}
	repo.enforceBufferLimitLocked(time.Now().UTC().Truncate(time.Minute))
}

func (repo *Repository) Run(ctx context.Context) {
	for {
		now := time.Now()
		nextMinute := now.Truncate(time.Minute).Add(time.Minute)
		timer := time.NewTimer(time.Until(nextMinute))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if err := repo.Flush(nextMinute); err != nil {
				repo.report(err)
			}
		}
	}
}

func (repo *Repository) Close() error {
	return repo.flushBefore(time.Now().UTC().Truncate(time.Minute).Add(time.Minute), true)
}

func (repo *Repository) Flush(now time.Time) error {
	return repo.flushBefore(now.UTC().Truncate(time.Minute), false)
}

func (repo *Repository) flushBefore(currentMinute time.Time, includeCurrent bool) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	minutes := repo.flushableMinutesLocked(currentMinute, includeCurrent)
	var firstErr error
	for _, minute := range minutes {
		if err := repo.writeMinute(minute, repo.buffers[minute]); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		delete(repo.buffers, minute)
	}
	repo.enforceBufferLimitLocked(currentMinute)
	if err := repo.cleanupFiles(currentMinute); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func (repo *Repository) Query(serviceName, metricName string, duration time.Duration, now time.Time) ([]model.MetricSample, error) {
	if _, ok := supportedDurations[duration]; !ok {
		return nil, fmt.Errorf("time range must be 1h, 4h, 8h, 16h, or 24h")
	}

	repo.fileMu.RLock()
	defer repo.fileMu.RUnlock()

	now = now.UTC()
	currentMinute := now.Truncate(time.Minute)
	cutoff := now.Add(-duration)
	paths, err := repo.queryFilesLocked(cutoff, currentMinute)
	if err != nil {
		return nil, err
	}

	series := make(map[string][]model.MetricSample)
	for _, path := range paths {
		samples, err := readMatchingPromFile(path, serviceName, metricName)
		if err != nil {
			return nil, fmt.Errorf("read trend file %s: %w", filepath.Base(path), err)
		}
		for _, sample := range samples {
			if sample.CollectedAt.Before(cutoff) || sample.CollectedAt.After(now) {
				continue
			}
			key := seriesKey(sample.ServiceName, sample.Name, sample.Labels)
			series[key] = append(series[key], sample)
		}
	}

	var result []model.MetricSample
	for _, samples := range series {
		sort.Slice(samples, func(i, j int) bool {
			return samples[i].CollectedAt.Before(samples[j].CollectedAt)
		})
		result = append(result, Downsample(samples, maxSeriesSamples)...)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CollectedAt.Equal(result[j].CollectedAt) {
			return seriesKey(result[i].ServiceName, result[i].Name, result[i].Labels) <
				seriesKey(result[j].ServiceName, result[j].Name, result[j].Labels)
		}
		return result[i].CollectedAt.Before(result[j].CollectedAt)
	})
	if len(result) == 0 {
		return nil, fmt.Errorf("no trend data for service %q metric %q in the requested %s range", serviceName, metricName, duration)
	}
	return result, nil
}

func (repo *Repository) writeMinute(minute time.Time, samples []model.MetricSample) error {
	repo.fileMu.Lock()
	defer repo.fileMu.Unlock()

	path := repo.minutePath(minute)
	existing, err := readPromFileIfExists(path)
	if err != nil {
		return fmt.Errorf("read existing trend minute: %w", err)
	}
	merged := deduplicateSamples(append(existing, samples...))
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].CollectedAt.Equal(merged[j].CollectedAt) {
			return seriesKey(merged[i].ServiceName, merged[i].Name, merged[i].Labels) <
				seriesKey(merged[j].ServiceName, merged[j].Name, merged[j].Labels)
		}
		return merged[i].CollectedAt.Before(merged[j].CollectedAt)
	})

	temp, err := os.CreateTemp(repo.directory, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create trend temporary file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	writer := bufio.NewWriter(temp)
	for _, sample := range merged {
		if _, err := writer.WriteString(formatPromSample(sample)); err != nil {
			temp.Close()
			return fmt.Errorf("write trend sample: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		temp.Close()
		return fmt.Errorf("flush trend temporary file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sync trend temporary file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close trend temporary file: %w", err)
	}
	if err := replaceFile(tempPath, path); err != nil {
		return fmt.Errorf("replace trend minute file: %w", err)
	}
	return nil
}

func (repo *Repository) queryFilesLocked(cutoff, currentMinute time.Time) ([]string, error) {
	entries, err := os.ReadDir(repo.directory)
	if err != nil {
		return nil, fmt.Errorf("read trend directory: %w", err)
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		minute, ok := parseMinuteFilename(entry.Name())
		if !ok || !minute.Before(currentMinute) || minute.Add(time.Minute).Before(cutoff) {
			continue
		}
		paths = append(paths, filepath.Join(repo.directory, entry.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}

func (repo *Repository) cleanup(now time.Time) error {
	return repo.cleanupFiles(now.UTC())
}

func (repo *Repository) cleanupFiles(now time.Time) error {
	repo.fileMu.Lock()
	defer repo.fileMu.Unlock()

	entries, err := os.ReadDir(repo.directory)
	if err != nil {
		return fmt.Errorf("read trend directory for cleanup: %w", err)
	}
	cutoff := now.Add(-repo.retention)
	for _, entry := range entries {
		minute, ok := parseMinuteFilename(entry.Name())
		if !ok || !minute.Add(time.Minute).Before(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(repo.directory, entry.Name())); err != nil {
			return fmt.Errorf("remove expired trend file %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func (repo *Repository) flushableMinutesLocked(currentMinute time.Time, includeCurrent bool) []time.Time {
	var minutes []time.Time
	for minute := range repo.buffers {
		if minute.Before(currentMinute) || includeCurrent && minute.Equal(currentMinute) {
			minutes = append(minutes, minute)
		}
	}
	sort.Slice(minutes, func(i, j int) bool { return minutes[i].Before(minutes[j]) })
	return minutes
}

func (repo *Repository) enforceBufferLimitLocked(currentMinute time.Time) {
	var pending []time.Time
	for minute := range repo.buffers {
		if minute.Before(currentMinute) {
			pending = append(pending, minute)
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Before(pending[j]) })
	for len(pending) > maxPendingMinutes {
		dropped := pending[0]
		delete(repo.buffers, dropped)
		pending = pending[1:]
		repo.report(fmt.Errorf("dropped trend buffer for %s after exceeding %d pending minutes", dropped.Format(time.RFC3339), maxPendingMinutes))
	}
}

func (repo *Repository) minutePath(minute time.Time) string {
	return filepath.Join(repo.directory, minute.UTC().Format(fileTimeLayout)+".prom")
}

func (repo *Repository) report(err error) {
	if err == nil {
		return
	}
	select {
	case repo.errors <- err:
	default:
	}
}

func readPromFileIfExists(path string) ([]model.MetricSample, error) {
	samples, err := readPromFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return samples, err
}

func readPromFile(path string) ([]model.MetricSample, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var samples []model.MetricSample
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
	for scanner.Scan() {
		sample, err := parsePromSample(scanner.Text())
		if err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}
	return samples, scanner.Err()
}

func readMatchingPromFile(path, serviceName, metricName string) ([]model.MetricSample, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metricWithLabels := metricName + "{"
	metricWithoutLabels := metricName + " "
	var samples []model.MetricSample
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, metricWithLabels) && !strings.HasPrefix(line, metricWithoutLabels) {
			continue
		}
		sample, err := parsePromSample(line)
		if err != nil {
			return nil, err
		}
		if sample.ServiceName == serviceName {
			samples = append(samples, sample)
		}
	}
	return samples, scanner.Err()
}

func formatPromSample(sample model.MetricSample) string {
	labels := cloneLabels(sample.Labels)
	labels[serviceLabel] = sample.ServiceName
	var builder strings.Builder
	builder.WriteString(sample.Name)
	if len(labels) > 0 {
		builder.WriteByte('{')
		keys := sortedLabelKeys(labels)
		for i, key := range keys {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.WriteString(key)
			builder.WriteByte('=')
			builder.WriteString(strconv.Quote(labels[key]))
		}
		builder.WriteByte('}')
	}
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatFloat(sample.Value, 'g', -1, 64))
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatInt(sample.CollectedAt.UnixMilli(), 10))
	builder.WriteByte('\n')
	return builder.String()
}

func parsePromSample(line string) (model.MetricSample, error) {
	nameAndLabels, valueText, timestampText, err := splitPromSample(line)
	if err != nil {
		return model.MetricSample{}, err
	}
	name := nameAndLabels
	labels := make(map[string]string)
	if open := strings.IndexByte(nameAndLabels, '{'); open >= 0 {
		if !strings.HasSuffix(nameAndLabels, "}") {
			return model.MetricSample{}, fmt.Errorf("invalid trend label format")
		}
		name = nameAndLabels[:open]
		labels, err = parseLabels(nameAndLabels[open+1 : len(nameAndLabels)-1])
		if err != nil {
			return model.MetricSample{}, err
		}
	}
	serviceName := labels[serviceLabel]
	delete(labels, serviceLabel)
	if serviceName == "" {
		return model.MetricSample{}, fmt.Errorf("trend sample missing %s label", serviceLabel)
	}
	value, err := strconv.ParseFloat(valueText, 64)
	if err != nil {
		return model.MetricSample{}, fmt.Errorf("invalid trend value: %w", err)
	}
	timestamp, err := strconv.ParseInt(timestampText, 10, 64)
	if err != nil {
		return model.MetricSample{}, fmt.Errorf("invalid trend timestamp: %w", err)
	}
	return model.MetricSample{
		ServiceName: serviceName,
		Name:        name,
		Labels:      labels,
		Value:       value,
		CollectedAt: time.UnixMilli(timestamp),
	}, nil
}

func splitPromSample(line string) (string, string, string, error) {
	end := strings.LastIndexByte(line, ' ')
	if end < 0 {
		return "", "", "", fmt.Errorf("invalid trend sample")
	}
	timestamp := strings.TrimSpace(line[end+1:])
	line = strings.TrimSpace(line[:end])
	end = strings.LastIndexByte(line, ' ')
	if end < 0 {
		return "", "", "", fmt.Errorf("invalid trend sample")
	}
	return line[:end], strings.TrimSpace(line[end+1:]), timestamp, nil
}

func parseLabels(text string) (map[string]string, error) {
	labels := make(map[string]string)
	for len(text) > 0 {
		equal := strings.IndexByte(text, '=')
		if equal <= 0 {
			return nil, fmt.Errorf("invalid trend label")
		}
		key := text[:equal]
		text = text[equal+1:]
		if len(text) == 0 || text[0] != '"' {
			return nil, fmt.Errorf("invalid trend label value")
		}
		end := quotedStringEnd(text)
		if end < 0 {
			return nil, fmt.Errorf("unterminated trend label value")
		}
		value, err := strconv.Unquote(text[:end+1])
		if err != nil {
			return nil, fmt.Errorf("invalid trend label value: %w", err)
		}
		labels[key] = value
		text = text[end+1:]
		if text == "" {
			break
		}
		if text[0] != ',' {
			return nil, fmt.Errorf("invalid trend label separator")
		}
		text = text[1:]
	}
	return labels, nil
}

func quotedStringEnd(value string) int {
	escaped := false
	for i := 1; i < len(value); i++ {
		switch {
		case escaped:
			escaped = false
		case value[i] == '\\':
			escaped = true
		case value[i] == '"':
			return i
		}
	}
	return -1
}

func deduplicateSamples(samples []model.MetricSample) []model.MetricSample {
	unique := make(map[string]model.MetricSample, len(samples))
	for _, sample := range samples {
		key := seriesKey(sample.ServiceName, sample.Name, sample.Labels) + "|" + strconv.FormatInt(sample.CollectedAt.UnixMilli(), 10)
		unique[key] = sample
	}
	result := make([]model.MetricSample, 0, len(unique))
	for _, sample := range unique {
		result = append(result, sample)
	}
	return result
}

func replaceFile(source, target string) error {
	if err := os.Rename(source, target); err == nil {
		return nil
	}
	old := target + ".replacing"
	_ = os.Remove(old)
	if err := os.Rename(target, old); err != nil {
		return err
	}
	if err := os.Rename(source, target); err != nil {
		_ = os.Rename(old, target)
		return err
	}
	return os.Remove(old)
}

func parseMinuteFilename(name string) (time.Time, bool) {
	if filepath.Ext(name) != ".prom" {
		return time.Time{}, false
	}
	minute, err := time.Parse(fileTimeLayout, strings.TrimSuffix(name, ".prom"))
	return minute, err == nil
}

func sortedLabelKeys(labels map[string]string) []string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneLabels(labels map[string]string) map[string]string {
	result := make(map[string]string, len(labels)+1)
	for key, value := range labels {
		result[key] = value
	}
	return result
}

func Downsample(samples []model.MetricSample, max int) []model.MetricSample {
	if len(samples) <= max || max <= 0 {
		return append([]model.MetricSample(nil), samples...)
	}
	if max == 1 {
		return []model.MetricSample{samples[len(samples)-1]}
	}
	result := make([]model.MetricSample, 0, max)
	step := float64(len(samples)-1) / float64(max-1)
	for i := 0; i < max; i++ {
		result = append(result, samples[int(float64(i)*step)])
	}
	return result
}

func seriesKey(serviceName, metricName string, labels map[string]string) string {
	var builder strings.Builder
	builder.WriteString(serviceName)
	builder.WriteByte('|')
	builder.WriteString(metricName)
	builder.WriteByte('|')
	for _, key := range sortedLabelKeys(labels) {
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(labels[key])
		builder.WriteByte(',')
	}
	return builder.String()
}
