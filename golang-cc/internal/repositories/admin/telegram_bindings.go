package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListTelegramBindings(ctx context.Context) ([]domain.TelegramBinding, error) {
	rows, err := r.pool.Query(ctx, `SELECT b.chat_id,b.user_id,u.email,u.display_name,b.created_at,b.updated_at
		FROM telegram_chat_bindings b JOIN users u ON u.id=b.user_id ORDER BY b.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.TelegramBinding{}
	for rows.Next() {
		var item domain.TelegramBinding
		if err := rows.Scan(&item.ChatID, &item.UserID, &item.Email, &item.DisplayName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateTelegramBinding(ctx context.Context, chatID int64, userID string) error {
	tag, err := r.pool.Exec(ctx, `INSERT INTO telegram_chat_bindings(chat_id,user_id)
		SELECT $1,id FROM users WHERE id=$2 AND role='member' AND status='active'`, chatID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteTelegramBinding(ctx context.Context, chatID int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM telegram_chat_bindings WHERE chat_id=$1`, chatID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
