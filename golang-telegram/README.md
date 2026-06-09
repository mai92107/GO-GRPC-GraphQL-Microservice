# Golang Spring Boot Monitor Bot

Runnable Go implementation based on `outputs/golang-springboot-monitor-bot-system-design.md`.

The current system monitors Spring Boot Actuator endpoints, stores runtime state in an in-memory repository, evaluates alert rules, and sends notifications through console logging or Telegram Bot API.

## Features

- Load monitored services and alert rules from JSON
- Periodically call `/actuator/health`
- Optionally call `/actuator/prometheus`
- Parse selected Spring Boot Prometheus metrics
- Evaluate health, response-time, and configurable Prometheus metric alerts
- Deduplicate open alerts
- Emit recovery notifications
- Provide local bot command handling for `/list_services`, `/status`, and `/alerts`
- Support Telegram long polling for read-only query commands when Telegram is enabled
- Include PostgreSQL migration schema in `migrations/001_init.sql`

## Run Once

```sh
GOCACHE=$(pwd)/.cache/go-build go run ./cmd/monitor -config configs/demo.json -once
```

## Run Continuously

```sh
GOCACHE=$(pwd)/.cache/go-build go run ./cmd/monitor -config configs/demo.json
```

Stop it with `Ctrl+C`.

## Telegram

Fill `telegram.bot_token` and `telegram.chat_id`, then start:

```sh
GOCACHE=$(pwd)/.cache/go-build go run ./cmd/monitor -config configs/telegram.json
```

Telegram is read-only. It can receive alert and recovery notifications and supports:

```text
/list_services
/status
/status <service_name>
/alerts
```

`/status` includes each server's latest health check and collected metrics.

Hosts and alert rules must be changed in the JSON config and require a process restart.

To forward alert and recovery events to a cron result notification API, add:

```json
"cron_result_notify": {
  "enabled": true,
  "host": "https://notify.example.com",
  "path": "/cron/result",
  "bearer_token": "replace-with-secret",
  "success_response_code": "SUCCESS"
}
```

The API receives `serviceName`, `ruleKey`, `severity`, `status`, `message`, `startedAt`,
and optional `resolvedAt` fields. The bearer token is sent through the `Authorization`
header and is not logged.

## Try Bot Commands Locally

The local command mode runs one monitor pass first, then formats a command response from the repository:

```sh
GOCACHE=$(pwd)/.cache/go-build go run ./cmd/monitor -config configs/demo.json -command "/status"
GOCACHE=$(pwd)/.cache/go-build go run ./cmd/monitor -config configs/demo.json -command "/alerts"
GOCACHE=$(pwd)/.cache/go-build go run ./cmd/monitor -config configs/demo.json -command "/list_services"
```

## Config

Edit `configs/demo.json`:

```json
{
  "app": {
    "name": "springboot-monitor-demo",
    "default_check_interval_seconds": 30,
    "http_timeout_seconds": 5,
    "metrics_enabled": true,
    "response_time_warning_ms": 2000
  },
  "telegram": {
    "enabled": false,
    "bot_token": "",
    "chat_id": ""
  },
  "services": [
    {
      "name": "order-service",
      "environment": "demo",
      "base_url": "http://localhost:8081",
      "health_path": "/actuator/health",
      "metrics_path": "/actuator/prometheus",
      "check_interval_seconds": 10,
      "enabled": true
    }
  ],
  "alert_rules": [
    {
      "key": "high-live-threads",
      "service_name": "order-service",
      "metric_name": "jvm_threads_live_threads",
      "labels": {},
      "operator": ">",
      "threshold": 100,
      "severity": "warning",
      "enabled": true
    },
    {
      "key": "high-process-cpu",
      "service_name": "*",
      "metric_name": "process_cpu_usage",
      "labels": {},
      "operator": ">",
      "threshold": 0.8,
      "severity": "warning",
      "enabled": true
    }
  ]
}
```

Supported operators: `>`, `>=`, `<`, `<=`, `==`, `!=`.

Supported `metric_name` enum values:

```text
http_server_requests_seconds_count
http_server_requests_seconds_sum
http_server_requests_seconds_max
jvm_memory_used_bytes
jvm_memory_max_bytes
jvm_gc_pause_seconds_count
jvm_gc_pause_seconds_sum
jvm_threads_live_threads
hikaricp_connections_active
hikaricp_connections_pending
process_cpu_usage
process_uptime_seconds
```

The application rejects the JSON config during startup when `metric_name` is not one of these values.

Set `telegram.enabled` to `true` and fill `bot_token` / `chat_id` to send alert, recovery, and command response messages through Telegram.

When Telegram is enabled in continuous mode, the monitor also listens for Telegram commands from the configured `chat_id`.

## Tests

```sh
GOCACHE=$(pwd)/.cache/go-build go test ./...
```
