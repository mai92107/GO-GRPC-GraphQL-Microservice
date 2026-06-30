package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Dashboard(ctx *gin.Context) {
	result, err := c.service.Dashboard(ctx.Request.Context())
	if err != nil {
		println("error getting dashboard data, error: ", err.Error())
		failure(ctx, http.StatusInternalServerError, "internal_error", "查詢失敗")
		return
	}
	data(ctx, http.StatusOK, mapDashboard(result))
}
