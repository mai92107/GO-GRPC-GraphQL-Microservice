CREATE TABLE IF NOT EXISTS services (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    environment VARCHAR(50) NOT NULL DEFAULT 'production',
    base_url TEXT NOT NULL,
    health_path VARCHAR(255) NOT NULL DEFAULT '/actuator/health',
    metrics_path VARCHAR(255) NOT NULL DEFAULT '/actuator/prometheus',
    check_interval_seconds INT NOT NULL DEFAULT 60,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS health_checks (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    status VARCHAR(50) NOT NULL,
    http_status_code INT,
    response_time_ms INT NOT NULL,
    error_message TEXT,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_checks_service_time
ON health_checks(service_id, checked_at DESC);

CREATE TABLE IF NOT EXISTS metric_snapshots (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    metric_name VARCHAR(255) NOT NULL,
    labels_json JSONB,
    value DOUBLE PRECISION NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metric_snapshots_service_metric_time
ON metric_snapshots(service_id, metric_name, collected_at DESC);

CREATE TABLE IF NOT EXISTS alert_rules (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT REFERENCES services(id),
    metric_name VARCHAR(255) NOT NULL,
    operator VARCHAR(20) NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    duration_seconds INT NOT NULL DEFAULT 0,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alert_events (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    rule_id BIGINT REFERENCES alert_rules(id),
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_alert_events_service_status
ON alert_events(service_id, status);

CREATE TABLE IF NOT EXISTS alert_silences (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    reason TEXT,
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
