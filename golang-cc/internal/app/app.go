package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	telegramcontroller "github.com/rafa/golang-cc/internal/controllers/telegram"
	"github.com/rafa/golang-cc/internal/platform/config"
	memberrepo "github.com/rafa/golang-cc/internal/repositories/member"
	publicrepo "github.com/rafa/golang-cc/internal/repositories/public"
	telegramrepo "github.com/rafa/golang-cc/internal/repositories/telegram"
	"github.com/rafa/golang-cc/internal/server"
	telegramserver "github.com/rafa/golang-cc/internal/server/telegram"
	memberservice "github.com/rafa/golang-cc/internal/services/member"
	publicservice "github.com/rafa/golang-cc/internal/services/public"
	"github.com/rafa/golang-cc/internal/utils/email"
	"github.com/rafa/golang-cc/web"
)

func RunServer() error {
	cfg, err := config.LoadFromProject(config.DefaultPath)
	if err != nil {
		return err
	}
	pool, err := pgxpool.New(context.Background(), cfg.Database.ConnectionString())
	if err != nil {
		return err
	}
	defer pool.Close()
	sender := email.SMTP{Addr: cfg.Mail.SMTPAddress, From: cfg.Mail.From}
	authService := publicservice.New(publicrepo.New(pool), sender, cfg.Server.PublicBaseURL)
	handler := server.New(pool, authService, cfg.Server.SecureCookie, cfg.Monitoring.Token, web.Handler())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if cfg.Telegram.Enabled {
		memberService := memberservice.New(memberrepo.New(pool), memberrepo.NewTransactionRepository(pool))
		controller := telegramcontroller.New(telegramrepo.New(pool), memberService)
		go telegramserver.NewPoller(cfg.Telegram.BotToken, controller).Run(ctx)
	}
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
