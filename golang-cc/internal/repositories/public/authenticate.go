package public

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) Authenticate(ctx context.Context, tokenHash []byte, now time.Time) (domain.User, error) {
	var user domain.User
	err := r.pool.QueryRow(ctx, `SELECT u.id,u.email,u.display_name,u.role,u.status
		FROM sessions s JOIN users u ON u.id=s.user_id
		WHERE s.token_hash=$1 AND s.expires_at>$2 AND u.status='active'`, tokenHash, now).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if err == nil {
		_, _ = r.pool.Exec(ctx, `UPDATE sessions SET last_seen_at=$2 WHERE token_hash=$1`, tokenHash, now)
	}
	return user, err
}
