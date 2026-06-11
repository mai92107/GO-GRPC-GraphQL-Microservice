package applog

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang-springboot-monitor-bot/internal/config"
)

func TestLoggerKeepsInfoOffTerminalAndRedactsSecrets(t *testing.T) {
	var terminalOutput bytes.Buffer
	terminal := NewTerminalWriter(&terminalOutput, "Asia/Taipei")
	logger, err := New(config.LogConfig{Directory: t.TempDir(), Level: "debug", RetentionDays: 14, Timezone: "Asia/Taipei"}, terminal)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info(context.Background(), "telegram_sent", Field{Key: "bot_token", Value: "secret"})
	if terminalOutput.Len() != 0 {
		t.Fatalf("info reached terminal: %q", terminalOutput.String())
	}
	logger.Error(context.Background(), "health_failed", errors.New("timeout"), Field{Key: "service", Value: "order-service"})
	if !strings.Contains(terminalOutput.String(), "ERROR ") || !strings.Contains(terminalOutput.String(), "order-service") {
		t.Fatalf("missing terminal error: %q", terminalOutput.String())
	}
	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(logger.cfg.Directory, "monitor-*.log"))
	data, _ := os.ReadFile(files[0])
	if strings.Contains(string(data), "secret") || !strings.Contains(string(data), "[REDACTED]") {
		t.Fatalf("secret was not redacted: %s", data)
	}
}
