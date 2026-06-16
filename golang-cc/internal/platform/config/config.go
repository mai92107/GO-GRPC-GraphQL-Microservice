package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

var telegramBotTokenPattern = regexp.MustCompile(`^[0-9]+:[A-Za-z0-9_-]+$`)

type Config struct {
	Database   Database   `json:"database"`
	Mail       Mail       `json:"mail"`
	Server     Server     `json:"server"`
	Monitoring Monitoring `json:"monitoring"`
	Telegram   Telegram   `json:"telegram"`
}

type Database struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

func (d Database) ConnectionString() string {
	result := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.Username, d.Password),
		Host:   net.JoinHostPort(d.Host, fmt.Sprint(d.Port)),
		Path:   d.Database,
	}
	query := result.Query()
	query.Set("sslmode", d.SSLMode)
	result.RawQuery = query.Encode()
	return result.String()
}

type Mail struct {
	SMTPAddress string `json:"smtp_address"`
	From        string `json:"from"`
	WebAddress  string `json:"web_address"`
}

type Server struct {
	Address       string `json:"address"`
	PublicBaseURL string `json:"public_base_url"`
	SecureCookie  bool   `json:"secure_cookie"`
}

type Monitoring struct {
	Token string `json:"token"`
}

type Telegram struct {
	Enabled  bool   `json:"enabled"`
	BotToken string `json:"bot_token"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Database.Username == "" || cfg.Database.Password == "" || cfg.Database.Host == "" ||
		cfg.Database.Port <= 0 || cfg.Database.Database == "" || cfg.Database.SSLMode == "" ||
		cfg.Mail.SMTPAddress == "" || cfg.Mail.From == "" ||
		cfg.Server.Address == "" || cfg.Server.PublicBaseURL == "" ||
		cfg.Monitoring.Token == "" || (cfg.Telegram.Enabled && cfg.Telegram.BotToken == "") {
		return Config{}, fmt.Errorf("config %s is missing required values", path)
	}
	if cfg.Telegram.Enabled && !telegramBotTokenPattern.MatchString(cfg.Telegram.BotToken) {
		return Config{}, fmt.Errorf("config %s telegram.bot_token format is invalid", path)
	}
	return cfg, nil
}

func LoadFromProject(path string) (Config, error) {
	if filepath.IsAbs(path) {
		return Load(path)
	}
	dir, err := os.Getwd()
	if err != nil {
		return Config{}, err
	}
	for {
		candidate := filepath.Join(dir, path)
		if _, err := os.Stat(candidate); err == nil {
			return Load(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Config{}, fmt.Errorf("config %s not found from project directory", path)
		}
		dir = parent
	}
}
