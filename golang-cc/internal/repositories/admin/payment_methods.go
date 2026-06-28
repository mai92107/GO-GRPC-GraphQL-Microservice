package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListPaymentMethods(ctx context.Context) ([]domain.PaymentMethod, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,is_active,is_system FROM payment_methods ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PaymentMethod{}
	for rows.Next() {
		var item domain.PaymentMethod
		if err := rows.Scan(&item.ID, &item.Name, &item.IsActive, &item.IsSystem); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) CreatePaymentMethod(ctx context.Context, item domain.PaymentMethod) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO payment_methods(id,name,is_active,is_system) VALUES($1,$2,$3,false)`, item.ID, item.Name, item.IsActive)
	return err
}

func (r *Repository) UpdatePaymentMethod(ctx context.Context, id, name string, active bool) error {
	tag, err := r.pool.Exec(ctx, `UPDATE payment_methods SET name=$2,is_active=$3,updated_at=now() WHERE id=$1`, id, name, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeletePaymentMethod(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM payment_methods WHERE id=$1 AND is_system=false`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
