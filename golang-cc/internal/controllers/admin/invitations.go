package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) CreateInvitation(ctx *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "Email 格式錯誤")
		return
	}
	if err := c.service.CreateInvitation(ctx.Request.Context(), input.Email, currentUserID(ctx)); err != nil {
		failure(ctx, http.StatusInternalServerError, "email_failed", "無法寄送邀請信")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"sent": true})
}

func (c *Controller) ListInvitations(ctx *gin.Context) {
	result, err := c.service.ListInvitations(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapInvitations(result))
}

func (c *Controller) DeleteInvitation(ctx *gin.Context) {
	if err := c.service.DeleteInvitation(ctx.Request.Context(), ctx.Param("id")); err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到未使用邀請")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}
