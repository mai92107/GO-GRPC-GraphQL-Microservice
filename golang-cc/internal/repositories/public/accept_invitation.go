package public

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) AcceptInvitation(ctx context.Context, tokenHash []byte, now time.Time, user domain.User, passwordHash string, preferenceIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var email string
	err = tx.QueryRow(ctx, `SELECT email FROM invitations WHERE token_hash=$1 AND accepted_at IS NULL AND expires_at>$2 FOR UPDATE`, tokenHash, now).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrInvalidToken
	}
	if err != nil {
		return err
	}
	user.Email = email
	if _, err = tx.Exec(ctx, `INSERT INTO users(id,email,password_hash,display_name,role,status) VALUES($1,$2,$3,$4,$5,$6)`,
		user.ID, user.Email, passwordHash, user.DisplayName, user.Role, user.Status); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE invitations SET accepted_at=$2 WHERE token_hash=$1`, tokenHash, now); err != nil {
		return err
	}
	for _, id := range preferenceIDs {
		if _, err = tx.Exec(ctx, `INSERT INTO reward_preferences(user_id,reward_unit_id,weight) VALUES($1,$2,1)`, user.ID, id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) InvitationEmail(ctx context.Context, tokenHash []byte, now time.Time) (string, error) {
	var email string
	err := r.pool.QueryRow(ctx, `SELECT email FROM invitations WHERE token_hash=$1 AND accepted_at IS NULL AND expires_at>$2`, tokenHash, now).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrInvalidToken
	}
	return email, err
}
