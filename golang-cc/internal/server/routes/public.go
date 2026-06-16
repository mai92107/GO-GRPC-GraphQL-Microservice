package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	publiccontroller "github.com/rafa/golang-cc/internal/controllers/public"
	servermw "github.com/rafa/golang-cc/internal/server/middleware"
)

func RegisterPublic(engine *gin.Engine, controller *publiccontroller.Controller, authService servermw.Authenticator) {
	rateLimit := servermw.RateLimit(10, time.Minute)
	public := engine.Group("/api/public/auth")
	public.POST("/login", rateLimit, controller.Login)
	public.POST("/accept-invitation", rateLimit, controller.AcceptInvitation)
	public.POST("/request-password-reset", rateLimit, controller.RequestPasswordReset)
	public.POST("/reset-password", rateLimit, controller.ResetPassword)

	authenticated := public.Group("", servermw.Authentication(authService))
	authenticated.GET("/me", controller.Me)
	authenticated.POST("/logout", servermw.CSRF(), controller.Logout)
}
