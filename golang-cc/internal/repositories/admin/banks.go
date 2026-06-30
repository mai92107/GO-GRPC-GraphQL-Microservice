package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListBanks(ctx context.Context) ([]domain.Bank, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,is_active FROM banks ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Bank{}
	for rows.Next() {
		var item domain.Bank
		if err := rows.Scan(&item.ID, &item.Name, &item.IsActive); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateBank(ctx context.Context, id string, input domain.BankInput) error {
	bank := domain.Bank{
		ID:         id,
		Name:       input.Name,
		WebsiteURL: nullString(input.WebsiteURL),
		IsActive:   input.IsActive,
	}

	return r.db.WithContext(ctx).Create(&bank).Error
}

func (r *Repository) UpdateBank(ctx context.Context, id string, input domain.BankInput) error {
	tag, err := r.pool.Exec(ctx, `UPDATE banks SET name=$2,website_url=NULLIF($3,''),is_active=$4,updated_at=now() WHERE id=$1`, id, input.Name, input.WebsiteURL, input.IsActive)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteBank(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM banks WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
