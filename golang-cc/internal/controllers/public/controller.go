package public

import (
	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/platform/httpcontext"
	service "github.com/rafa/golang-cc/internal/services/public"
)

const sessionCookie = "cc_session"

type Controller struct {
	service      *service.Service
	secureCookie bool
}

func New(service *service.Service, secureCookie bool) *Controller {
	return &Controller{service: service, secureCookie: secureCookie}
}

func writeData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func writeError(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get("request_id")
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message, "fields": nil, "request_id": requestID}})
}

type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

func mapUser(user domain.User) userResponse {
	return userResponse(user)
}

func currentToken(c *gin.Context) string {
	token, _ := httpcontext.Token(c.Request)
	return token
}
