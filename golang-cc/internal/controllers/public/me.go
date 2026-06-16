package public

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/platform/httpcontext"
)

func (c *Controller) Me(ctx *gin.Context) {
	user, _ := httpcontext.CurrentUser(ctx.Request)
	writeData(ctx, http.StatusOK, gin.H{"user": mapUser(user), "csrf_token": httpcontext.CSRFToken(currentToken(ctx))})
}
