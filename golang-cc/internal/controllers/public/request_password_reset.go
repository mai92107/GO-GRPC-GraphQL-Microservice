package public

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (c *Controller) RequestPasswordReset(ctx *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if ctx.ShouldBindJSON(&input) == nil {
		_ = c.service.RequestPasswordReset(ctx.Request.Context(), strings.TrimSpace(input.Email))
	}
	writeData(ctx, http.StatusOK, gin.H{"accepted": true})
}
