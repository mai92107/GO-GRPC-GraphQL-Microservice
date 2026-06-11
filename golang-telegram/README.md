# Golang Spring Boot Monitor Bot

Single-process Go monitor for fewer than ten Spring Boot Actuator services.
Detailed behavior and architecture are defined in
`outputs/golang-springboot-monitor-bot-detailed-system-design.md`.

## Capabilities

- Poll `/actuator/health` and `/actuator/prometheus` concurrently per service
- Evaluate health, response-time, and supported Prometheus metric alerts
- Deduplicate alerts, acknowledge open alerts, and send recovery notifications
- Manage services and alert rules through a single authorized Telegram chat
- Validate, back up, atomically replace, and immediately apply JSON config
- Keep supported metric trends in memory for eight hours
- Generate 1h, 4h, and 8h PNG trend charts with threshold lines
- Write structured daily JSONL logs while keeping normal terminal output limited
  to `監控 <service> 中`
- Keep optional `cron_result_notify` notification integration

There is no mute, unmute, or notification-suppression capability. JSON is the
only persistent configuration source. Runtime state and trends reset at restart.

## Run

```powershell
go run ./cmd/monitor -config configs/demo.json
go run ./cmd/monitor -config configs/demo.json -once
go test ./...
```

Enable Telegram by setting `telegram.enabled`, `telegram.bot_token`, and the
single administrator `telegram.chat_id`.

## Telegram Commands

```text
/menu
/list_services
/status
/status <service_name>
/metric <service_name>
/alerts
/check [service_name]
/trend <service_name> <metric_name> <1h|4h|8h>
```

`/menu` opens the Inline Keyboard management interface. Service and alert-rule
mutations collect required values, display a summary, require confirmation,
write the JSON configuration safely, and immediately update the running monitor.
The bot registers a clickable Telegram command menu at startup, so users can
select common commands from the input area instead of typing them.
Trend charts also use a button flow: select service, metric group, metric
subgroup, exact metric, and then the 1h, 4h, or 8h range.

## Logging

Default configuration:

```json
{
  "logging": {
    "directory": "logs",
    "level": "info",
    "retention_days": 14,
    "timezone": "Asia/Taipei"
  }
}
```

All `INFO`, `WARN`, `ERROR`, and optionally `DEBUG` events are written to
`logs/monitor-YYYY-MM-DD.log`. Errors additionally appear on the terminal.
Tokens, authorization values, credentials, and Telegram update payloads are
redacted.
