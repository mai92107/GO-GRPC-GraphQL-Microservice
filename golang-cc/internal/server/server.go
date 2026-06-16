package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	admincontroller "github.com/rafa/golang-cc/internal/controllers/admin"
	membercontroller "github.com/rafa/golang-cc/internal/controllers/member"
	publiccontroller "github.com/rafa/golang-cc/internal/controllers/public"
	actuatorrepo "github.com/rafa/golang-cc/internal/repositories/actuator"
	adminrepo "github.com/rafa/golang-cc/internal/repositories/admin"
	memberrepo "github.com/rafa/golang-cc/internal/repositories/member"
	publicrepo "github.com/rafa/golang-cc/internal/repositories/public"
	servermw "github.com/rafa/golang-cc/internal/server/middleware"
	"github.com/rafa/golang-cc/internal/server/routes"
	adminservice "github.com/rafa/golang-cc/internal/services/admin"
	memberservice "github.com/rafa/golang-cc/internal/services/member"
	publicservice "github.com/rafa/golang-cc/internal/services/public"
	"github.com/rafa/golang-cc/internal/utils/email"
)

func New(pool *pgxpool.Pool, sender email.Sender, publicBaseURL string, secureCookie bool, monitoringToken string, frontend http.Handler) http.Handler {
	// 設置 Gin
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	metrics := servermw.NewMetrics(actuatorrepo.NewPoolStatsRepository(pool))
	engine.Use(
		gin.Recovery(), 
		servermw.RequestID(), 
		metrics.Handler(),
	)

	// 註冊路由和控制器
	authService := publicservice.New(publicrepo.New(pool), sender, publicBaseURL)
	publicController := publiccontroller.New(authService, secureCookie)
	adminController := admincontroller.New(adminservice.New(adminrepo.New(pool), authService))
	memberController := membercontroller.New(memberservice.New(memberrepo.New(pool), memberrepo.NewTransactionRepository(pool)))
	routes.RegisterActuator(engine, actuatorrepo.NewHealthRepository(pool), metrics, servermw.ActuatorAuthorization(authService, monitoringToken))
	routes.RegisterPublic(engine, publicController, authService)
	routes.RegisterAdmin(engine, adminController, authService)
	routes.RegisterMember(engine, memberController, authService)
	if frontend != nil {
		engine.NoRoute(gin.WrapH(frontend))
	}
	return engine
}
