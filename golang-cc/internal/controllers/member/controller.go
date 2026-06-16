package member

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/platform/httpcontext"
	service "github.com/rafa/golang-cc/internal/services/member"
)

type Controller struct{ service *service.Service }

func New(service *service.Service) *Controller { return &Controller{service: service} }

func userID(c *gin.Context) string {
	user, _ := httpcontext.CurrentUser(c.Request)
	return user.ID
}

func data(c *gin.Context, status int, value any) { c.JSON(status, gin.H{"data": value}) }

func failure(c *gin.Context, status int, code, message string) {
	id, _ := c.Get("request_id")
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message, "fields": nil, "request_id": id}})
}

func notFound(err error) bool { return errors.Is(err, domain.ErrNotFound) }
