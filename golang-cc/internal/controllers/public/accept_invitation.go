package public

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type acceptInvitationRequest struct {
	Token       string `json:"token" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Password    string `json:"password" binding:"required,min=12,max=128"`
}

func (c *Controller) AcceptInvitation(ctx *gin.Context) {
	var input acceptInvitationRequest
	if ctx.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.DisplayName) == "" {
		writeError(ctx, http.StatusBadRequest, "validation_failed", "輸入格式錯誤")
		return
	}
	user, err := c.service.AcceptInvitation(ctx.Request.Context(), input.Token, strings.TrimSpace(input.DisplayName), input.Password)
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid_invitation", "邀請連結無效或已過期")
		return
	}
	writeData(ctx, http.StatusCreated, mapUser(user))
}
