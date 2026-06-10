package parser

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"golang-springboot-monitor-bot/internal/metric"
	"golang-springboot-monitor-bot/internal/model"
)

func ParsePrometheus(serviceName string, text string) ([]model.MetricSnapshot, error) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	snapshots := make([]model.MetricSnapshot, 0)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		snapshot, ok, err := parseMetricLine(serviceName, line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		if ok {
			snapshots = append(snapshots, snapshot)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return snapshots, nil
}

func parseMetricLine(serviceName, line string) (model.MetricSnapshot, bool, error) {
	nameAndLabels, valueText, err := splitMetricAndValue(line)
	if err != nil {
		return model.MetricSnapshot{}, false, err
	}

	name := nameAndLabels
	labels := map[string]string{}
	if strings.Contains(nameAndLabels, "{") {
		before, after, ok := strings.Cut(nameAndLabels, "{")
		if !ok || !strings.HasSuffix(after, "}") {
			return model.MetricSnapshot{}, false, fmt.Errorf("invalid label format")
		}
		name = before
		parsed, err := parseLabels(strings.TrimSuffix(after, "}"))
		if err != nil {
			return model.MetricSnapshot{}, false, err
		}
		labels = parsed
	}

	metricName := metric.Name(name)
	if !metricName.Valid() {
		return model.MetricSnapshot{}, false, nil
	}

	value, err := strconv.ParseFloat(valueText, 64)
	if err != nil {
		return model.MetricSnapshot{}, false, fmt.Errorf("invalid metric value %q: %w", valueText, err)
	}

	return model.MetricSnapshot{
		ServiceName: serviceName,
		Name:        name,
		Labels:      labels,
		Value:       value,
	}, true, nil
}

func splitMetricAndValue(line string) (string, string, error) {
	if open := strings.Index(line, "{"); open >= 0 {
		close := strings.Index(line[open:], "}")
		if close < 0 {
			return "", "", fmt.Errorf("invalid label format")
		}
		end := open + close + 1
		fields := strings.Fields(strings.TrimSpace(line[end:]))
		if len(fields) == 0 {
			return "", "", fmt.Errorf("missing metric value")
		}
		return line[:end], fields[0], nil
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", "", fmt.Errorf("missing metric value")
	}
	return fields[0], fields[1], nil
}

func parseLabels(text string) (map[string]string, error) {
	labels := map[string]string{}
	if text == "" {
		return labels, nil
	}

	parts := splitCSV(text)
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" && i == len(parts)-1 {
			continue
		}
		key, rawValue, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid label %q", part)
		}
		key = strings.TrimSpace(key)
		rawValue = strings.TrimSpace(rawValue)
		value, err := strconv.Unquote(rawValue)
		if err != nil {
			return nil, fmt.Errorf("invalid label value %q: %w", rawValue, err)
		}
		labels[key] = value
	}
	return labels, nil
}

func splitCSV(text string) []string {
	var result []string
	var current strings.Builder
	escaped := false
	quoted := false

	for _, r := range text {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			current.WriteRune(r)
			escaped = true
		case r == '"':
			current.WriteRune(r)
			quoted = !quoted
		case r == ',' && !quoted:
			result = append(result, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	result = append(result, current.String())
	return result
}
