package admin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListUsers(ctx context.Context) ([]domain.AdminUser, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,email,display_name,role,status,created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.AdminUser{}
	for rows.Next() {
		var user domain.AdminUser
		if err := rows.Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	return result, rows.Err()
}

func (r *Repository) UpdateUserStatus(ctx context.Context, id, currentUserID, status string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET status=$3,updated_at=now() WHERE id=$1 AND id<>$2`, id, currentUserID, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if status == "disabled" {
		_, err = r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, id)
	}
	return err
}

func (r *Repository) UserEmail(ctx context.Context, id string) (string, error) {
	var email string
	err := r.pool.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, id).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	return email, err
}
