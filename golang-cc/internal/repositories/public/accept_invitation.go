package public

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
	"gorm.io/gorm"
)

func (r *Repository) AcceptInvitation(
	ctx context.Context,
	tokenHash []byte,
	now time.Time,
	user domain.User,
	passwordHash string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var invitation domain.Invitation
		var newUser domain.Users
		err := tx.
			Where("token_hash = ? AND accepted_at IS NULL AND expires_at > ?", tokenHash, now).
			Take(&invitation).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrInvalidToken
		}

		if err != nil {
			return err
		}
		newUser.User = user
		newUser.Email = invitation.Email
		newUser.PasswordHash = passwordHash

		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.Invitation{}).
			Where("token_hash = ?", tokenHash).
			Update("accepted_at", now).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *Repository) InvitationEmail(ctx context.Context, tokenHash []byte, now time.Time) (string, error) {
	var email string
	err := r.pool.QueryRow(ctx, `SELECT email FROM invitations WHERE token_hash=$1 AND accepted_at IS NULL AND expires_at>$2`, tokenHash, now).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrInvalidToken
	}
	return email, err
}
