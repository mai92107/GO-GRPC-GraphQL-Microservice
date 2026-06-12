package bot

import (
	"testing"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

func TestSessionStoreRemovesExpiredSession(t *testing.T) {
	store := NewSessionStore()
	store.Put(model.TelegramSession{
		ChatID:    "123",
		ExpiresAt: time.Now().Add(-time.Second),
	})

	if _, ok := store.Get("123"); ok {
		t.Fatal("expired session must not be returned")
	}
	if len(store.sessions) != 0 {
		t.Fatal("expired session must be removed")
	}
}
