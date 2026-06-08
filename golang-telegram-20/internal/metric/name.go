package metric

import (
	"fmt"
	"sort"
	"strings"
)

type Name string

const (
	HTTPServerRequestsSecondsCount Name = "http_server_requests_seconds_count"
	HTTPServerRequestsSecondsSum   Name = "http_server_requests_seconds_sum"
	HTTPServerRequestsSecondsMax   Name = "http_server_requests_seconds_max"
	JVMMemoryUsedBytes             Name = "jvm_memory_used_bytes"
	JVMMemoryMaxBytes              Name = "jvm_memory_max_bytes"
	JVMGCPauseSecondsCount         Name = "jvm_gc_pause_seconds_count"
	JVMGCPauseSecondsSum           Name = "jvm_gc_pause_seconds_sum"
	JVMThreadsLiveThreads          Name = "jvm_threads_live_threads"
	HikariCPConnectionsActive      Name = "hikaricp_connections_active"
	HikariCPConnectionsPending     Name = "hikaricp_connections_pending"
	ProcessCPUUsage                Name = "process_cpu_usage"
	ProcessUptimeSeconds           Name = "process_uptime_seconds"
)

var allowed = map[Name]struct{}{
	HTTPServerRequestsSecondsCount: {},
	HTTPServerRequestsSecondsSum:   {},
	HTTPServerRequestsSecondsMax:   {},
	JVMMemoryUsedBytes:             {},
	JVMMemoryMaxBytes:              {},
	JVMGCPauseSecondsCount:         {},
	JVMGCPauseSecondsSum:           {},
	JVMThreadsLiveThreads:          {},
	HikariCPConnectionsActive:      {},
	HikariCPConnectionsPending:     {},
	ProcessCPUUsage:                {},
	ProcessUptimeSeconds:           {},
}

func (name Name) Valid() bool {
	_, ok := allowed[name]
	return ok
}

func (name Name) String() string {
	return string(name)
}

func AllowedValues() []Name {
	values := make([]Name, 0, len(allowed))
	for name := range allowed {
		values = append(values, name)
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	return values
}

func AllowedValuesText() string {
	values := AllowedValues()
	text := make([]string, 0, len(values))
	for _, value := range values {
		text = append(text, value.String())
	}
	return strings.Join(text, ", ")
}

func Validate(name Name) error {
	if !name.Valid() {
		return fmt.Errorf("unsupported metric_name %q; allowed values: %s", name, AllowedValuesText())
	}
	return nil
}
