package httpcontext

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/rafa/golang-cc/internal/domain"
)

type key string

const (
	userKey      key = "user"
	tokenKey     key = "session-token"
	requestIDKey key = "request-id"
)

func WithUser(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func CurrentUser(r *http.Request) (domain.User, bool) {
	user, ok := r.Context().Value(userKey).(domain.User)
	return user, ok
}

func Token(r *http.Request) (string, bool) {
	token, ok := r.Context().Value(tokenKey).(string)
	return token, ok
}

func CSRFToken(session string) string {
	sum := sha256.Sum256([]byte("csrf:" + session))
	return hex.EncodeToString(sum[:])
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{
		"code": code, "message": message, "fields": fields, "request_id": r.Context().Value(requestIDKey),
	}})
}
