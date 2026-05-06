package router

import (
	"batchLog/api/home"

	"github.com/gin-gonic/gin"
)


func RegisterRoutes(r *gin.Engine) {
	// TODO: 未來需要檢查ip body header 在路徑後加上middleware檢查
	// 依類別分組
	homeGroup := r.Group("/home")
	{
		homeGroup.GET("/say_hello", home.SayHello)
	}
}