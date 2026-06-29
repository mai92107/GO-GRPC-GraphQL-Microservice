package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	telegramcontroller "github.com/rafa/golang-cc/internal/controllers/telegram"
	"github.com/rafa/golang-cc/internal/platform/config"
	memberrepo "github.com/rafa/golang-cc/internal/repositories/member"
	telegramrepo "github.com/rafa/golang-cc/internal/repositories/telegram"
	"github.com/rafa/golang-cc/internal/server"
	telegramserver "github.com/rafa/golang-cc/internal/server/telegram"
	memberservice "github.com/rafa/golang-cc/internal/services/member"
	"github.com/rafa/golang-cc/internal/utils/email"
	"github.com/rafa/golang-cc/web"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RunServer() error {
	// load 基本資料
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/local.json"
	}
	cfg, err := config.LoadFromProject(configPath)
	if err != nil {
		return err
	}

	// 連接資料庫
	pool, err := pgxpool.New(context.Background(), cfg.Database.ConnectionString())
	if err != nil {
		return err
	}
	gormDB, err := gorm.Open(postgres.Open(cfg.Database.ConnectionString()), &gorm.Config{})
	if err != nil {
		return err
	}
	defer pool.Close()

	// 初始化服務和伺服器
	sender := email.SMTP{Addr: cfg.Mail.SMTPAddress, From: cfg.Mail.From}
	handler := server.New(pool, gormDB, sender, cfg.Server.PublicBaseURL, cfg.Server.SecureCookie, cfg.Monitoring.Token, web.Handler())

	if cfg.Telegram.Enabled {
		// 啟動 Telegram Bot
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		memberService := memberservice.New(memberrepo.New(pool, gormDB), memberrepo.NewTransactionRepository(pool))
		tgcontroller := telegramcontroller.New(telegramrepo.New(pool), memberService)
		go telegramserver.NewPoller(cfg.Telegram.BotToken, tgcontroller).Run(ctx)
	}

	// 啟動 HTTP 伺服器
	server := &http.Server{
		Addr: cfg.Server.Address, Handler: handler, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second,
	}
	log.Printf("listening on %s", server.Addr)
	return server.ListenAndServe()
}

func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
