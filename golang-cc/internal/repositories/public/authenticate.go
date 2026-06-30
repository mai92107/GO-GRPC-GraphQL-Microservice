package public

import (
	"context"
	"errors"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
	"gorm.io/gorm"
)

func (r *Repository) Authenticate(
	ctx context.Context,
	tokenHash []byte,
	now time.Time,
) (domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).
		Table("identity.sessions AS s").
		Select(`
			u.id,
			u.email,
			u.display_name,
			u.role,
			u.status
		`).
		Joins("JOIN identity.users AS u ON u.id = s.user_id").
		Where("s.token_hash = ? AND s.expires_at > ? AND u.status = ?", tokenHash, now, "active").
		Take(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	if err != nil {
		return domain.User{}, err
	}

	_ = r.db.WithContext(ctx).
		Model(domain.Session{}).
		Where("token_hash = ?", tokenHash).
		Update("last_seen_at", now).Error

	return user, nil
}
