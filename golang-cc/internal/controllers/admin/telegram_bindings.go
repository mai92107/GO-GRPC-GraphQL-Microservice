package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type telegramBindingRequest struct {
	ChatID int64  `json:"chat_id" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
}

func (c *Controller) ListTelegramBindings(ctx *gin.Context) {
	result, err := c.service.ListTelegramBindings(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapTelegramBindings(result))
}

func (c *Controller) CreateTelegramBinding(ctx *gin.Context) {
	var input telegramBindingRequest
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "chat_id 與 user_id 必填")
		return
	}
	if err := c.service.CreateTelegramBinding(ctx.Request.Context(), input.ChatID, input.UserID); err != nil {
		if isNotFound(err) {
			failure(ctx, http.StatusNotFound, "member_not_found", "找不到啟用中的會員")
			return
		}
		failure(ctx, http.StatusConflict, "binding_conflict", "Chat 或會員已綁定")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"chat_id": input.ChatID, "user_id": input.UserID})
}

func (c *Controller) DeleteTelegramBinding(ctx *gin.Context) {
	chatID, err := strconv.ParseInt(ctx.Param("chat_id"), 10, 64)
	if err != nil || chatID == 0 {
		failure(ctx, http.StatusBadRequest, "validation_failed", "chat_id 無效")
		return
	}
	if err := c.service.DeleteTelegramBinding(ctx.Request.Context(), chatID); err != nil {
		if isNotFound(err) {
			failure(ctx, http.StatusNotFound, "not_found", "找不到 Telegram 綁定")
			return
		}
		failure(ctx, http.StatusInternalServerError, "internal_error", "刪除失敗")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}
