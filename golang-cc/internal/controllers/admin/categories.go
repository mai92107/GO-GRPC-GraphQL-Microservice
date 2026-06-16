package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ListCategories(ctx *gin.Context) {
	result, err := c.service.ListCategories(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapCategories(result))
}
func (c *Controller) CreateCategory(ctx *gin.Context) {
	var input struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	if err := c.service.CreateCategory(ctx.Request.Context(), input.Code, input.Name); err != nil {
		failure(ctx, http.StatusConflict, "conflict", "類別重複")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"code": input.Code})
}
func (c *Controller) UpdateCategory(ctx *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		IsActive *bool  `json:"is_active" binding:"required"`
	}
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	if err := c.service.UpdateCategory(ctx.Request.Context(), ctx.Param("code"), input.Name, *input.IsActive); err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到類別")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}
func (c *Controller) DeleteCategory(ctx *gin.Context) {
	err := c.service.DeleteCategory(ctx.Request.Context(), ctx.Param("code"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到可刪除類別")
		return
	}
	if err != nil {
		failure(ctx, http.StatusConflict, "category_in_use", "類別使用中，請停用")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}
