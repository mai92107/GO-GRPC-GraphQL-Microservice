package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagerUpdatePersistsAndApplies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	initial := []byte(`{"services":[{"name":"order-service","base_url":"http://localhost:8081","enabled":true}]}`)
	if err := os.WriteFile(path, initial, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	applied := false
	manager := NewManager(path, cfg, func(Config) { applied = true })
	err = manager.Update(func(candidate *Config) error {
		candidate.Services[0].BaseURL = "http://localhost:9091"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !applied {
		t.Fatal("expected runtime apply callback")
	}
	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Services[0].BaseURL != "http://localhost:9091" {
		t.Fatalf("unexpected URL: %s", reloaded.Services[0].BaseURL)
	}
	backups, _ := filepath.Glob(path + ".backup-*")
	if len(backups) != 1 {
		t.Fatalf("expected one backup, got %d", len(backups))
	}
}

func TestManagerInvalidUpdateLeavesCurrentAndFileUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	initial := []byte(`{"services":[{"name":"order-service","base_url":"http://localhost:8081","enabled":true}]}`)
	if err := os.WriteFile(path, initial, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, _ := Load(path)
	manager := NewManager(path, cfg, nil)
	err := manager.Update(func(candidate *Config) error {
		candidate.Services[0].BaseURL = ""
		return nil
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	data, _ := os.ReadFile(path)
	if string(data) != string(initial) {
		t.Fatal("invalid update changed config file")
	}
	if manager.Current().Services[0].BaseURL == "" {
		t.Fatal("invalid update changed active config")
	}
}
