package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const sessionCookie = "cc_session"

func ActuatorAuthorization(authService Authenticator, monitoringToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearer := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if bearer != "" && subtle.ConstantTimeCompare([]byte(bearer), []byte(monitoringToken)) == 1 {
			c.Next()
			return
		}

		token, err := c.Cookie(sessionCookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "需要管理員登入或監控 Token"}})
			return
		}
		user, err := authService.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "登入已失效"}})
			return
		}
		if user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "需要管理員權限"}})
			return
		}
		c.Next()
	}
}
