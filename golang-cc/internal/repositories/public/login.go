package public

import (
	"context"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
	"gorm.io/gorm"
)

func (r *Repository) FindLoginUser(ctx context.Context, email string) (domain.LoginUser, error) {
	var result domain.LoginUser

	err := r.db.WithContext(ctx).
		Model(domain.Users{}).
		Where("email = ?", email).
		Scan(&result).Error

	if err != nil {
		return result, err
	}

	if result.ID == "" {
		return result, gorm.ErrRecordNotFound
	}

	return result, nil
}

func (r *Repository) CreateSession(
	ctx context.Context,
	id, userID string,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	session := domain.Session{
		Id:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	return r.db.WithContext(ctx).Create(&session).Error
}
