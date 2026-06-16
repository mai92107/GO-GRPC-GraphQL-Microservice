package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/platform/httpcontext"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := secure.UUID()
		c.Header("X-Request-ID", id)
		c.Set("request_id", id)
		c.Request = c.Request.WithContext(httpcontext.WithRequestID(c.Request.Context(), id))
		c.Next()
	}
}
