package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProjectTestConfig(t *testing.T) {
	cfg, err := LoadFromProject("configs/test.json")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.ConnectionString() == "" || cfg.Server.Address != ":8080" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestConnectionStringEscapesCredentials(t *testing.T) {
	database := Database{Username: "user@example", Password: "p@ss word", Host: "localhost", Port: 5432, Database: "cards", SSLMode: "disable"}
	got := database.ConnectionString()
	want := "postgres://user%40example:p%40ss%20word@localhost:5432/cards?sslmode=disable"
	if got != want {
		t.Fatalf("connection string = %q, want %q", got, want)
	}
}

func TestTelegramTokenRequiredWhenEnabled(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "configs", "test.json"))
	if err != nil {
		t.Fatal(err)
	}
	enabled := strings.Replace(string(raw), `"enabled": false`, `"enabled": true`, 1)
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(enabled), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected enabled telegram without bot token to fail")
	}
}

func TestTelegramTokenFormatRequiredWhenEnabled(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "configs", "test.json"))
	if err != nil {
		t.Fatal(err)
	}
	enabled := strings.Replace(string(raw), `"enabled": false`, `"enabled": true`, 1)
	invalid := strings.Replace(enabled, `"bot_token": ""`, `"bot_token": "invalid-token"`, 1)
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(invalid), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "telegram.bot_token format is invalid") {
		t.Fatalf("expected invalid telegram token format error, got %v", err)
	}
}

func TestValidTelegramTokenLoads(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "configs", "test.json"))
	if err != nil {
		t.Fatal(err)
	}
	enabled := strings.Replace(string(raw), `"enabled": false`, `"enabled": true`, 1)
	valid := strings.Replace(enabled, `"bot_token": ""`, `"bot_token": "123456789:ABC_def-123"`, 1)
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err != nil {
		t.Fatalf("expected valid telegram token to load, got %v", err)
	}
}
