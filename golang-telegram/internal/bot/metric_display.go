package bot

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

var metricChineseNames = map[string]string{
	"http_server_requests_seconds_count": "HTTP 請求次數",
	"http_server_requests_seconds_sum":   "HTTP 請求總耗時",
	"http_server_requests_seconds_max":   "HTTP 最大耗時",
	"jvm_memory_used_bytes":              "JVM 已使用記憶體",
	"jvm_memory_max_bytes":               "JVM 最大記憶體",
	"jvm_gc_pause_seconds_count":         "GC 暫停次數",
	"jvm_gc_pause_seconds_sum":           "GC 暫停總時間",
	"jvm_threads_live_threads":           "JVM 存活執行緒",
	"jvm_threads_daemon_threads":         "JVM 背景執行緒",
	"jvm_threads_peak_threads":           "JVM 歷史最高執行緒",
	"jvm_threads_started_threads_total":  "JVM 累積啟動執行緒",
	"jvm_threads_states_threads":         "JVM 執行緒狀態",
	"hikaricp_connections_active":        "資料庫使用中連線",
	"hikaricp_connections_pending":       "資料庫等待連線",
	"process_cpu_usage":                  "程序 CPU 使用率",
	"system_cpu_usage":                   "系統 CPU 使用率",
	"system_cpu_count":                   "系統 CPU 核心數",
	"process_uptime_seconds":             "程序運行時間",
}

func formatMetricSnapshot(snapshot model.MetricSnapshot) string {
	chineseName := metricChineseNames[snapshot.Name]
	if chineseName == "" {
		chineseName = "監控指標"
	}
	lines := []string{fmt.Sprintf("- %s（%s）", chineseName, snapshot.Name)}
	if content := formatIdentifyingLabels(snapshot.Labels); content != "" {
		lines = append(lines, "  監控內容："+content)
	}
	lines = append(lines,
		"  數值："+formatMetricValue(snapshot.Name, snapshot.Value),
		"  收集時間："+snapshot.CollectedAt.Format("2006-01-02 15:04:05"),
	)
	return strings.Join(lines, "\n")
}

func formatIdentifyingLabels(labels map[string]string) string {
	names := make([]string, 0, len(labels))
	for name, value := range labels {
		if omitNormalLabel(name, value) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	pairs := make([]string, 0, len(names))
	for _, name := range names {
		pairs = append(pairs, fmt.Sprintf("%s=%q", name, labels[name]))
	}
	return strings.Join(pairs, "、")
}

func omitNormalLabel(name, value string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	value = strings.ToLower(strings.TrimSpace(value))
	switch name {
	case "error", "exception":
		return value == "none"
	case "method", "methods":
		return value == "options"
	case "outcome":
		return value == "success"
	case "status":
		return value == "200"
	default:
		return false
	}
}

func formatMetricValue(name string, value float64) string {
	switch {
	case name == "process_cpu_usage" || name == "system_cpu_usage":
		return fmt.Sprintf("%.2f%%", value*100)
	case strings.HasSuffix(name, "_bytes"):
		return formatBytes(value)
	case name == "process_uptime_seconds":
		return formatUptime(value)
	case strings.Contains(name, "_seconds_sum") || strings.Contains(name, "_seconds_max"):
		return formatSeconds(value)
	case strings.HasSuffix(name, "_count") || strings.HasSuffix(name, "_threads") || strings.Contains(name, "connections_"):
		return strconv.FormatFloat(value, 'f', -1, 64)
	default:
		return strconv.FormatFloat(value, 'f', -1, 64)
	}
}

func formatBytes(value float64) string {
	const (
		kib = 1024
		mib = 1024 * kib
		gib = 1024 * mib
	)
	switch {
	case value >= gib:
		return fmt.Sprintf("%.2f GB", value/gib)
	case value >= mib:
		return fmt.Sprintf("%.2f MB", value/mib)
	case value >= kib:
		return fmt.Sprintf("%.2f KB", value/kib)
	default:
		return fmt.Sprintf("%.0f bytes", value)
	}
}

func formatSeconds(value float64) string {
	if value < 1 {
		return fmt.Sprintf("%.0f 毫秒", value*1000)
	}
	return fmt.Sprintf("%.2f 秒", value)
}

func formatUptime(value float64) string {
	duration := time.Duration(value * float64(time.Second))
	days := duration / (24 * time.Hour)
	duration %= 24 * time.Hour
	hours := duration / time.Hour
	duration %= time.Hour
	minutes := duration / time.Minute
	seconds := duration % time.Minute / time.Second
	parts := make([]string, 0, 4)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d 天", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d 小時", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d 分鐘", minutes))
	}
	if len(parts) == 0 || seconds > 0 {
		parts = append(parts, fmt.Sprintf("%d 秒", seconds))
	}
	return strings.Join(parts, " ")
}
