package telegram

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafa/golang-cc/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) UserIDByChat(ctx context.Context, chatID int64) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `SELECT u.id FROM telegram_chat_bindings b
		JOIN users u ON u.id=b.user_id
		WHERE b.chat_id=$1 AND u.role='member' AND u.status='active'`, chatID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	return userID, err
}
