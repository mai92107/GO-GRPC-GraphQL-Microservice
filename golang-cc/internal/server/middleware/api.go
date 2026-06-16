package middleware

import (
	"context"
	"crypto/subtle"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/platform/httpcontext"
)

type Authenticator interface {
	Authenticate(context.Context, string) (domain.User, error)
}

func Authentication(authService Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessionCookie)
		if err != nil {
			abort(c, http.StatusUnauthorized, "unauthorized", "請先登入")
			return
		}
		user, err := authService.Authenticate(c.Request.Context(), token)
		if err != nil {
			abort(c, http.StatusUnauthorized, "unauthorized", "登入已失效")
			return
		}
		ctx := httpcontext.WithToken(httpcontext.WithUser(c.Request.Context(), user), token)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func Role(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := httpcontext.CurrentUser(c.Request)
		if !ok || user.Role != role {
			abort(c, http.StatusForbidden, "forbidden", "權限不足")
			return
		}
		c.Next()
	}
}

func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := httpcontext.Token(c.Request)
		if !ok || subtle.ConstantTimeCompare([]byte(httpcontext.CSRFToken(token)), []byte(c.GetHeader("X-CSRF-Token"))) != 1 {
			abort(c, http.StatusForbidden, "csrf_failed", "CSRF token 無效")
			return
		}
		c.Next()
	}
}

type rateEntry struct {
	start time.Time
	count int
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	entries := make(map[string]rateEntry)
	return func(c *gin.Context) {
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			host = c.Request.RemoteAddr
		}
		key, now := host+":"+c.FullPath(), time.Now()
		mu.Lock()
		entry := entries[key]
		if entry.start.IsZero() || now.Sub(entry.start) >= window {
			entry = rateEntry{start: now}
		}
		entry.count++
		entries[key] = entry
		mu.Unlock()
		if entry.count > limit {
			abort(c, http.StatusTooManyRequests, "rate_limited", "請稍後再試")
			return
		}
		c.Next()
	}
}

func abort(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get("request_id")
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{
		"code": code, "message": message, "fields": nil, "request_id": requestID,
	}})
}
