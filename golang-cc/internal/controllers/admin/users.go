package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ListUsers(ctx *gin.Context) {
	result, err := c.service.ListUsers(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapUsers(result))
}

func (c *Controller) UpdateUserStatus(ctx *gin.Context) {
	var input struct {
		Status string `json:"status" binding:"required,oneof=active disabled"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "狀態無效")
		return
	}
	if err := c.service.UpdateUserStatus(ctx.Request.Context(), ctx.Param("id"), currentUserID(ctx), input.Status); err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到使用者")
		return
	}
	data(ctx, http.StatusOK, gin.H{"status": input.Status})
}

func (c *Controller) SendUserPasswordReset(ctx *gin.Context) {
	if err := c.service.SendUserPasswordReset(ctx.Request.Context(), ctx.Param("id")); err != nil {
		if isNotFound(err) {
			failure(ctx, http.StatusNotFound, "not_found", "找不到使用者")
			return
		}
		println("error sending reset email, error: " + err.Error())
		failure(ctx, http.StatusInternalServerError, "email_failed", "寄送失敗")
		return
	}
	data(ctx, http.StatusOK, gin.H{"sent": true})
}
