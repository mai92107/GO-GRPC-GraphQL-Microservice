package admin

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListUsers(ctx context.Context) ([]domain.AdminUser, error) {
	var result []domain.AdminUser

	err := r.db.WithContext(ctx).
		Model(domain.Users{}).
		Order("created_at DESC").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) UpdateUserStatus(ctx context.Context, id, currentUserID, status string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Users{}).
		Where("id = ? AND id <> ?", id, currentUserID).
		Updates(map[string]any{
			"status":     status,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	if status != "disabled" {
		return nil
	}

	err := r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("user_id = ?", id).
		Updates(map[string]any{
			"expires_at": time.Now(),
		}).Error

	return err
}

func (r *Repository) UserEmail(ctx context.Context, id string) (string, error) {
	var email string
	err := r.pool.QueryRow(ctx, `SELECT email FROM identity.users WHERE id=$1`, id).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	return email, err
}
