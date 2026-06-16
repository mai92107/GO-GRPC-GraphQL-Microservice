package public

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ResetPassword(ctx *gin.Context) {
	var input struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=12,max=128"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		writeError(ctx, http.StatusBadRequest, "validation_failed", "輸入格式錯誤")
		return
	}
	if err := c.service.ResetPassword(ctx.Request.Context(), input.Token, input.Password); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid_reset_token", "重設連結無效或已過期")
		return
	}
	writeData(ctx, http.StatusOK, gin.H{"reset": true})
}
