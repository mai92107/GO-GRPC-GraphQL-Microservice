package public

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Logout(ctx *gin.Context) {
	_ = c.service.Logout(ctx.Request.Context(), currentToken(ctx))
	http.SetCookie(ctx.Writer, &http.Cookie{Name: sessionCookie, Path: "/", MaxAge: -1, HttpOnly: true, Secure: c.secureCookie, SameSite: http.SameSiteLaxMode})
	writeData(ctx, http.StatusOK, gin.H{"logged_out": true})
}
