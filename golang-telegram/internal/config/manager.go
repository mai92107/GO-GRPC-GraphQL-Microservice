package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type ApplyFunc func(Config)

type Manager struct {
	mu      sync.RWMutex
	path    string
	current Config
	apply   ApplyFunc
}

func NewManager(path string, current Config, apply ApplyFunc) *Manager {
	return &Manager{path: path, current: current, apply: apply}
}

func (manager *Manager) Current() Config {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return clone(manager.current)
}

func (manager *Manager) Update(mutate func(*Config) error) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	candidate := clone(manager.current)
	if err := mutate(&candidate); err != nil {
		return err
	}
	applyDefaults(&candidate)
	if err := validate(candidate); err != nil {
		return err
	}
	if err := manager.persist(candidate); err != nil {
		return err
	}
	manager.current = candidate
	if manager.apply != nil {
		manager.apply(clone(candidate))
	}
	return nil
}

func (manager *Manager) persist(candidate Config) error {
	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(manager.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	temp, err := os.CreateTemp(dir, filepath.Base(manager.path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create config temporary file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return fmt.Errorf("chmod config temporary file: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("write config temporary file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sync config temporary file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close config temporary file: %w", err)
	}

	backup := manager.path + ".backup-" + time.Now().Format("20060102T150405.000000000")
	if err := copyFile(manager.path, backup); err != nil {
		return fmt.Errorf("backup config: %w", err)
	}
	if err := replaceFile(tempPath, manager.path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return manager.pruneBackups()
}

func replaceFile(source, target string) error {
	if err := os.Rename(source, target); err == nil {
		return nil
	}
	old := target + ".replacing"
	_ = os.Remove(old)
	if err := os.Rename(target, old); err != nil {
		return err
	}
	if err := os.Rename(source, target); err != nil {
		_ = os.Rename(old, target)
		return err
	}
	return os.Remove(old)
}

func copyFile(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o600)
}

func (manager *Manager) pruneBackups() error {
	pattern := manager.path + ".backup-*"
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	sort.Strings(paths)
	for len(paths) > 10 {
		if err := os.Remove(paths[0]); err != nil {
			return err
		}
		paths = paths[1:]
	}
	return nil
}

func clone(cfg Config) Config {
	data, _ := json.Marshal(cfg)
	var result Config
	_ = json.Unmarshal(data, &result)
	return result
}

func FindService(cfg Config, name string) (int, bool) {
	for i := range cfg.Services {
		if strings.EqualFold(cfg.Services[i].Name, name) {
			return i, true
		}
	}
	return -1, false
}

func FindAlertRule(cfg Config, key string) (int, bool) {
	for i := range cfg.AlertRules {
		if strings.EqualFold(cfg.AlertRules[i].Key, key) {
			return i, true
		}
	}
	return -1, false
}
