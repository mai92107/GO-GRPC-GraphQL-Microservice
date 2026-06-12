package bot

import (
	"sync"
	"time"

	"golang-springboot-monitor-bot/internal/model"
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]model.TelegramSession
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]model.TelegramSession)}
}

func (store *SessionStore) Get(chatID string) (model.TelegramSession, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[chatID]
	if ok && time.Now().After(session.ExpiresAt) {
		delete(store.sessions, chatID)
		return model.TelegramSession{}, false
	}
	return session, ok
}

func (store *SessionStore) Put(session model.TelegramSession) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.sessions[session.ChatID] = session
}

func (store *SessionStore) Delete(chatID string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.sessions, chatID)
}
