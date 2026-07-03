package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type lookupRequest struct {
	ID       string `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	IsActive *bool  `json:"is_active"`
}

func (c *Controller) ListRequirementTypes(ctx *gin.Context) {
	items, err := c.service.ListRequirementTypes(ctx.Request.Context())
	if err != nil {
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapRequirementTypeOptions(items))
}

func (c *Controller) ListRegions(ctx *gin.Context) {
	items, err := c.service.ListRegions(ctx.Request.Context())
	if err != nil {
		println("Error listing regions:", err)
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, items)
}

func (c *Controller) CreateRegion(ctx *gin.Context) {
	var input lookupRequest
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	if err := c.service.CreateRegion(ctx.Request.Context(), input.ID, input.Name); err != nil {
		failure(ctx, http.StatusConflict, "conflict", "地區資料無效或重複")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"id": input.ID})
}

func (c *Controller) UpdateRegion(ctx *gin.Context) {
	var input lookupRequest
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	if err := c.service.UpdateRegion(ctx.Request.Context(), ctx.Param("id"), input.Name, *input.IsActive); err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到地區")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}

func (c *Controller) DeleteRegion(ctx *gin.Context) {
	err := c.service.DeleteRegion(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到地區")
		return
	}
	if err != nil {
		failure(ctx, http.StatusConflict, "region_in_use", "地區使用中，請停用")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}

func (c *Controller) ListUserQualifications(ctx *gin.Context) {
	items, err := c.service.ListUserQualifications(ctx.Request.Context())
	if err != nil {
		println("error listing user qualifications:", err.Error())
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, items)
}

func (c *Controller) CreateUserQualification(ctx *gin.Context) {
	var input lookupRequest
	if ctx.ShouldBindJSON(&input) != nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	if err := c.service.CreateUserQualification(ctx.Request.Context(), input.ID, input.Name); err != nil {
		failure(ctx, http.StatusConflict, "conflict", "會員資格資料無效或重複")
		return
	}
	data(ctx, http.StatusCreated, gin.H{"id": input.ID})
}

func (c *Controller) UpdateUserQualification(ctx *gin.Context) {
	var input lookupRequest
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		failure(ctx, http.StatusBadRequest, "validation_failed", "資料無效")
		return
	}
	if err := c.service.UpdateUserQualification(ctx.Request.Context(), ctx.Param("id"), input.Name, *input.IsActive); err != nil {
		failure(ctx, http.StatusNotFound, "not_found", "找不到會員資格")
		return
	}
	data(ctx, http.StatusOK, gin.H{"updated": true})
}

func (c *Controller) DeleteUserQualification(ctx *gin.Context) {
	err := c.service.DeleteUserQualification(ctx.Request.Context(), ctx.Param("id"))
	if isNotFound(err) {
		failure(ctx, http.StatusNotFound, "not_found", "找不到會員資格")
		return
	}
	if err != nil {
		failure(ctx, http.StatusConflict, "qualification_in_use", "會員資格使用中，請停用")
		return
	}
	data(ctx, http.StatusOK, gin.H{"deleted": true})
}
