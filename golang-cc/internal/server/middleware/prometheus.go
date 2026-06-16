package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	actuatorrepo "github.com/rafa/golang-cc/internal/repositories/actuator"
)

type Metrics struct {
	Registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewMetrics(stats *actuatorrepo.PoolStatsRepository) *Metrics {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_server_requests_total",
		Help: "Total HTTP requests handled by route, method, and status.",
	}, []string{"route", "method", "status"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_server_request_duration_seconds",
		Help: "HTTP request duration by route and method.",
	}, []string{"route", "method"})
	registry.MustRegister(prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}), requests, duration, newPoolCollector(stats))
	return &Metrics{Registry: registry, requests: requests, duration: duration}
}

func (m *Metrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		m.requests.WithLabelValues(route, c.Request.Method, strconv.Itoa(c.Writer.Status())).Inc()
		m.duration.WithLabelValues(route, c.Request.Method).Observe(time.Since(start).Seconds())
	}
}

type poolCollector struct {
	stats *actuatorrepo.PoolStatsRepository
	descs map[string]*prometheus.Desc
}

func newPoolCollector(stats *actuatorrepo.PoolStatsRepository) prometheus.Collector {
	names := []string{"total_connections", "acquired_connections", "idle_connections", "constructing_connections", "max_connections", "acquire_duration_seconds"}
	descs := make(map[string]*prometheus.Desc, len(names))
	for _, name := range names {
		descs[name] = prometheus.NewDesc("pgxpool_"+name, "Current pgxpool "+name+".", nil, nil)
	}
	return &poolCollector{stats: stats, descs: descs}
}

func (c *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
}

func (c *poolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.stats.Snapshot()
	values := map[string]float64{
		"total_connections":        float64(stat.TotalConns),
		"acquired_connections":     float64(stat.AcquiredConns),
		"idle_connections":         float64(stat.IdleConns),
		"constructing_connections": float64(stat.ConstructingConns),
		"max_connections":          float64(stat.MaxConns),
		"acquire_duration_seconds": stat.AcquireSeconds,
	}
	for name, value := range values {
		ch <- prometheus.MustNewConstMetric(c.descs[name], prometheus.GaugeValue, value)
	}
}
