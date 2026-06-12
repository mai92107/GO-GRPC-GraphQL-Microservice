package model

import "time"

type Service struct {
	Name                 string
	Environment          string
	BaseURL              string
	HealthPath           string
	MetricsPath          string
	CheckIntervalSeconds int
	Enabled              bool
}

type HealthCheck struct {
	ServiceName    string
	Environment    string
	Status         string
	HTTPStatusCode int
	ResponseTime   time.Duration
	ErrorMessage   string
	CheckedAt      time.Time
}

type MetricSnapshot struct {
	ServiceName string
	Name        string
	Labels      map[string]string
	Value       float64
	CollectedAt time.Time
}

type AlertStatus string

const (
	AlertStatusOpen     AlertStatus = "open"
	AlertStatusResolved AlertStatus = "resolved"
)

type AlertEvent struct {
	ID              int64
	ServiceName     string
	RuleKey         string
	MetricName      string
	Severity        string
	Status          AlertStatus
	Message         string
	ImagePNG        []byte
	StartedAt       time.Time
	LastTriggeredAt time.Time
	ResolvedAt      *time.Time
	AcknowledgedAt  *time.Time
	AcknowledgedBy  string
}

type MetricSample = MetricSnapshot

type TelegramSession struct {
	ChatID    string
	Operation string
	Step      string
	Values    map[string]string
	ExpiresAt time.Time
}

type ConfigBackupMetadata struct {
	Path      string
	CreatedAt time.Time
}

type ServiceSnapshot struct {
	Service      Service
	LastHealth   *HealthCheck
	LastMetrics  []MetricSnapshot
	OpenAlerts   []AlertEvent
	RecentAlerts []AlertEvent
}
