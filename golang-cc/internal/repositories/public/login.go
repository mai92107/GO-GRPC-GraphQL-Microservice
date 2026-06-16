package public

import (
	"context"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) FindLoginUser(ctx context.Context, email string) (domain.LoginUser, error) {
	var result domain.LoginUser
	err := r.pool.QueryRow(ctx, `SELECT id,email,display_name,role,status,password_hash FROM users WHERE email=$1`, email).
		Scan(&result.ID, &result.Email, &result.DisplayName, &result.Role, &result.Status, &result.PasswordHash)
	return result, err
}

func (r *Repository) CreateSession(ctx context.Context, id, userID string, tokenHash []byte, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO sessions(id,user_id,token_hash,expires_at) VALUES($1,$2,$3,$4)`, id, userID, tokenHash, expiresAt)
	return err
}
