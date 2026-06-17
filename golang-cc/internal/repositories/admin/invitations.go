package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListInvitings(ctx context.Context) ([]domain.Invitation, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,email,expires_at,accepted_at,created_at FROM invitations WHERE accepted_at IS NULL ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Invitation{}
	for rows.Next() {
		var item domain.Invitation
		if err := rows.Scan(&item.ID, &item.Email, &item.ExpiresAt, &item.AcceptedAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) DeleteInvitation(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM invitations WHERE id=$1 AND accepted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
