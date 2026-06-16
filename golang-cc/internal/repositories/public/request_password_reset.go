package public

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ActiveUserByEmail(ctx context.Context, email string) (string, string, error) {
	var id, actualEmail string
	err := r.pool.QueryRow(ctx, `SELECT id,email FROM users WHERE email=$1 AND status='active'`, email).Scan(&id, &actualEmail)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", domain.ErrNotFound
	}
	return id, actualEmail, err
}

func (r *Repository) CreatePasswordReset(ctx context.Context, id, userID string, tokenHash []byte, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO password_reset_tokens(id,user_id,token_hash,expires_at) VALUES($1,$2,$3,$4)`, id, userID, tokenHash, expiresAt)
	return err
}
