package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListMerchants(ctx context.Context) ([]domain.Merchant, error) {
	rows, err := r.pool.Query(ctx, `SELECT m.id,m.name,m.is_active,m.is_system,
		ARRAY(SELECT a.alias FROM merchant_aliases a WHERE a.merchant_id=m.id ORDER BY a.alias),
		ARRAY(SELECT c.category_id FROM merchant_categories c WHERE c.merchant_id=m.id ORDER BY c.category_id)
		FROM merchants m ORDER BY m.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Merchant{}
	for rows.Next() {
		var item domain.Merchant
		if err := rows.Scan(&item.Id, &item.Name, &item.IsActive, &item.IsSystem, &item.Aliases, &item.CategoryIDs); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) CreateMerchant(ctx context.Context, item domain.Merchant) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	if err = tx.QueryRow(ctx, `INSERT INTO merchants(name,is_active,is_system) VALUES($1,$2,false) RETURNING id`, item.Name, item.IsActive).
		Scan(&item.Id); err != nil {
		return "", err
	}
	for _, alias := range item.Aliases {
		if _, err = tx.Exec(ctx, `INSERT INTO merchant_aliases(merchant_id,alias) VALUES($1,$2)`, item.Id, alias); err != nil {
			return "", err
		}
	}
	for _, category := range item.CategoryIDs {
		if _, err = tx.Exec(ctx, `INSERT INTO merchant_categories(merchant_id,category_id) VALUES($1,$2)`, item.Id, category); err != nil {
			return "", err
		}
	}
	err = tx.Commit(ctx)
	return item.Id, err
}

func (r *Repository) UpdateMerchant(ctx context.Context, item domain.Merchant) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE merchants SET name=$2,is_active=$3,updated_at=now() WHERE id=$1`, item.Id, item.Name, item.IsActive)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if _, err = tx.Exec(ctx, `DELETE FROM merchant_aliases WHERE merchant_id=$1`, item.Id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM merchant_categories WHERE merchant_id=$1`, item.Id); err != nil {
		return err
	}
	for _, alias := range item.Aliases {
		if _, err = tx.Exec(ctx, `INSERT INTO merchant_aliases(merchant_id,alias) VALUES($1,$2)`, item.Id, alias); err != nil {
			return err
		}
	}
	for _, category := range item.CategoryIDs {
		if _, err = tx.Exec(ctx, `INSERT INTO merchant_categories(merchant_id,category_id) VALUES($1,$2)`, item.Id, category); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) DeleteMerchant(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM merchants WHERE id=$1 AND is_system=false`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
