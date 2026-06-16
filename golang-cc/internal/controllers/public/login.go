package public

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/platform/httpcontext"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (c *Controller) Login(ctx *gin.Context) {
	var input loginRequest
	if ctx.ShouldBindJSON(&input) != nil {
		writeError(ctx, http.StatusBadRequest, "validation_failed", "輸入格式錯誤")
		return
	}
	user, token, err := c.service.Login(ctx.Request.Context(), strings.TrimSpace(input.Email), input.Password)
	if err != nil {
		writeError(ctx, http.StatusUnauthorized, "invalid_credentials", "Email 或密碼錯誤")
		return
	}
	http.SetCookie(ctx.Writer, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: c.secureCookie, SameSite: http.SameSiteLaxMode, MaxAge: int(30 * 24 * 60 * 60)})
	writeData(ctx, http.StatusOK, gin.H{"user": mapUser(user), "csrf_token": httpcontext.CSRFToken(token)})
}
