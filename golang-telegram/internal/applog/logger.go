package applog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang-springboot-monitor-bot/internal/config"
)

type Field struct {
	Key   string
	Value any
}

type ApplicationLogger interface {
	Debug(context.Context, string, ...Field)
	Info(context.Context, string, ...Field)
	Warn(context.Context, string, ...Field)
	Error(context.Context, string, error, ...Field)
	Close() error
}

type TerminalStatusWriter interface {
	Monitoring(string)
	Error(string, error)
}

type TerminalWriter struct {
	out io.Writer
	loc *time.Location
	mu  sync.Mutex
}

func NewTerminalWriter(out io.Writer, timezone string) *TerminalWriter {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.Local
	}
	return &TerminalWriter{out: out, loc: loc}
}

func (writer *TerminalWriter) Monitoring(serviceName string) {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	fmt.Fprintf(writer.out, "監控 %s 中\n", serviceName)
}

func (writer *TerminalWriter) Error(serviceName string, err error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	fmt.Fprintf(writer.out, "ERROR %s %s %s\n", time.Now().In(writer.loc).Format(time.RFC3339), serviceName, redactText(err.Error()))
}

type Logger struct {
	mu       sync.Mutex
	cfg      config.LogConfig
	loc      *time.Location
	terminal TerminalStatusWriter
	file     *os.File
	date     string
}

type FallbackLogger struct {
	terminal TerminalStatusWriter
}

func NewFallback(terminal TerminalStatusWriter) ApplicationLogger {
	return &FallbackLogger{terminal: terminal}
}

func (logger *FallbackLogger) Debug(context.Context, string, ...Field) {}
func (logger *FallbackLogger) Info(context.Context, string, ...Field)  {}
func (logger *FallbackLogger) Warn(context.Context, string, ...Field)  {}
func (logger *FallbackLogger) Error(_ context.Context, _ string, err error, fields ...Field) {
	if err == nil {
		err = fmt.Errorf("application error")
	}
	if logger.terminal != nil {
		logger.terminal.Error(fieldString(fields, "service"), err)
	}
}
func (logger *FallbackLogger) Close() error { return nil }

func New(cfg config.LogConfig, terminal TerminalStatusWriter) (*Logger, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, err
	}
	logger := &Logger{cfg: cfg, loc: loc, terminal: terminal}
	if err := logger.rotate(time.Now()); err != nil {
		return nil, err
	}
	return logger, nil
}

func (logger *Logger) Debug(ctx context.Context, event string, fields ...Field) {
	logger.write("DEBUG", event, nil, fields...)
}
func (logger *Logger) Info(ctx context.Context, event string, fields ...Field) {
	logger.write("INFO", event, nil, fields...)
}
func (logger *Logger) Warn(ctx context.Context, event string, fields ...Field) {
	logger.write("WARN", event, nil, fields...)
}
func (logger *Logger) Error(ctx context.Context, event string, err error, fields ...Field) {
	if err == nil {
		err = fmt.Errorf("%s", event)
	}
	logger.write("ERROR", event, err, fields...)
	service := fieldString(fields, "service")
	if logger.terminal != nil {
		logger.terminal.Error(service, err)
	}
}

func (logger *Logger) write(level, event string, eventErr error, fields ...Field) {
	if !enabled(logger.cfg.Level, level) {
		return
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	now := time.Now().In(logger.loc)
	if err := logger.rotate(now); err != nil {
		if logger.terminal != nil {
			logger.terminal.Error("logger", err)
		}
		return
	}
	record := map[string]any{"time": now.Format(time.RFC3339Nano), "level": level, "event": event}
	for _, field := range fields {
		if sensitive(field.Key) {
			record[field.Key] = "[REDACTED]"
		} else {
			record[field.Key] = field.Value
		}
	}
	if eventErr != nil {
		record["error"] = redactText(eventErr.Error())
	}
	data, _ := json.Marshal(record)
	if _, err := logger.file.Write(append(data, '\n')); err != nil && logger.terminal != nil {
		logger.terminal.Error("logger", err)
	}
}

func (logger *Logger) rotate(now time.Time) error {
	date := now.In(logger.loc).Format("2006-01-02")
	if logger.file != nil && logger.date == date {
		return nil
	}
	if logger.file != nil {
		_ = logger.file.Sync()
		_ = logger.file.Close()
	}
	if err := os.MkdirAll(logger.cfg.Directory, 0o700); err != nil {
		return err
	}
	path := filepath.Join(logger.cfg.Directory, "monitor-"+date+".log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	logger.file, logger.date = file, date
	return logger.cleanup(now)
}

func (logger *Logger) cleanup(now time.Time) error {
	entries, err := os.ReadDir(logger.cfg.Directory)
	if err != nil {
		return err
	}
	cutoff := now.In(logger.loc).AddDate(0, 0, -(logger.cfg.RetentionDays - 1))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "monitor-") || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		dateText := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), "monitor-"), ".log")
		date, err := time.ParseInLocation("2006-01-02", dateText, logger.loc)
		if err == nil && date.Before(cutoff) {
			_ = os.Remove(filepath.Join(logger.cfg.Directory, entry.Name()))
		}
	}
	return nil
}

func (logger *Logger) Close() error {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	if logger.file == nil {
		return nil
	}
	if err := logger.file.Sync(); err != nil {
		return err
	}
	return logger.file.Close()
}

func enabled(configured, level string) bool {
	ranks := map[string]int{"debug": 0, "info": 1, "warn": 2, "error": 3}
	return ranks[strings.ToLower(level)] >= ranks[strings.ToLower(configured)]
}
func fieldString(fields []Field, key string) string {
	for _, field := range fields {
		if field.Key == key {
			return fmt.Sprint(field.Value)
		}
	}
	return ""
}
func sensitive(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "token") || strings.Contains(key, "authorization") || strings.Contains(key, "credential") || strings.Contains(key, "telegram_update")
}
func redactText(value string) string {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "bearer ") || strings.Contains(lower, "/bot") {
		return "[REDACTED]"
	}
	return value
}
