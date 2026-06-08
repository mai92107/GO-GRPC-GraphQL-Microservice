package bot

import (
	"strings"
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/repository"
)

func TestHandleStatusCommand(t *testing.T) {
	repo := repository.NewMemoryRepository([]model.Service{{Name: "order-service", Enabled: true}})
	repo.SaveHealthCheck(model.HealthCheck{
		ServiceName:  "order-service",
		Status:       "UP",
		ResponseTime: 120 * time.Millisecond,
		CheckedAt:    time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC),
	})

	response := NewHandler(repo).HandleCommand("/status order-service")
	if !strings.Contains(response, "order-service: UP") {
		t.Fatalf("unexpected response: %q", response)
	}
}

func TestModificationCommandsAreUnsupported(t *testing.T) {
	repo := repository.NewMemoryRepository(nil)
	response := NewHandler(repo).HandleCommand("/set_host order-service http://10.0.0.12:8081")

	if response != "unknown command" {
		t.Fatalf("unexpected response: %q", response)
	}
	if len(repo.Services()) != 0 {
		t.Fatal("telegram command must not modify services")
	}
}
