package routes

import (
	"github.com/gin-gonic/gin"
	membercontroller "github.com/rafa/golang-cc/internal/controllers/member"
	servermw "github.com/rafa/golang-cc/internal/server/middleware"
)

func RegisterMember(engine *gin.Engine, c *membercontroller.Controller, authService servermw.Authenticator) {
	member := engine.Group("/api/member", servermw.Authentication(authService), servermw.Role("member"))
	member.GET("/catalog/cards", c.Catalog)
	member.GET("/cards", c.ListCards)
	member.GET("/cards/:id", c.GetCard)
	member.GET("/reward-units", c.RewardUnits)
	member.GET("/categories", c.Categories)
	member.GET("/payment-methods", c.PaymentMethods)
	member.GET("/merchants", c.Merchants)
	member.GET("/reward-preferences", c.Preferences)
	member.GET("/transactions", c.ListTransactions)
	member.GET("/transactions/:id", c.GetTransaction)

	write := member.Group("", servermw.CSRF())
	write.POST("/cards", c.CreateCard)
	write.PATCH("/cards/:id", c.UpdateCard)
	write.DELETE("/cards/:id", c.DeleteCard)
	write.PATCH("/reward-preferences", c.UpdatePreferences)
	write.POST("/recommendations", c.Recommend)
	write.POST("/transactions", c.CreateTransaction)
	write.PATCH("/transactions/:id", c.UpdateTransaction)
	write.DELETE("/transactions/:id", c.DeleteTransaction)
}
