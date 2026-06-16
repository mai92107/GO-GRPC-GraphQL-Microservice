package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListMerchants(ctx context.Context) ([]domain.Merchant, error) {
	rows, err := r.pool.Query(ctx, `SELECT m.code,m.name,m.is_active,m.is_system,
		ARRAY(SELECT a.alias FROM merchant_aliases a WHERE a.merchant_code=m.code ORDER BY a.alias),
		ARRAY(SELECT c.category_code FROM merchant_categories c WHERE c.merchant_code=m.code ORDER BY c.category_code)
		FROM merchants m ORDER BY m.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Merchant{}
	for rows.Next() {
		var item domain.Merchant
		if err := rows.Scan(&item.Code, &item.Name, &item.IsActive, &item.IsSystem, &item.Aliases, &item.CategoryCodes); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) CreateMerchant(ctx context.Context, item domain.Merchant) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `INSERT INTO merchants(code,name,is_active,is_system) VALUES($1,$2,true,false)`, item.Code, item.Name); err != nil {
		return err
	}
	for _, alias := range item.Aliases {
		if _, err = tx.Exec(ctx, `INSERT INTO merchant_aliases(merchant_code,alias) VALUES($1,$2)`, item.Code, alias); err != nil {
			return err
		}
	}
	for _, category := range item.CategoryCodes {
		if _, err = tx.Exec(ctx, `INSERT INTO merchant_categories(merchant_code,category_code) VALUES($1,$2)`, item.Code, category); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) UpdateMerchant(ctx context.Context, item domain.Merchant) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE merchants SET name=$2,is_active=$3,updated_at=now() WHERE code=$1`, item.Code, item.Name, item.IsActive)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if _, err = tx.Exec(ctx, `DELETE FROM merchant_aliases WHERE merchant_code=$1`, item.Code); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM merchant_categories WHERE merchant_code=$1`, item.Code); err != nil {
		return err
	}
	for _, alias := range item.Aliases {
		if _, err = tx.Exec(ctx, `INSERT INTO merchant_aliases(merchant_code,alias) VALUES($1,$2)`, item.Code, alias); err != nil {
			return err
		}
	}
	for _, category := range item.CategoryCodes {
		if _, err = tx.Exec(ctx, `INSERT INTO merchant_categories(merchant_code,category_code) VALUES($1,$2)`, item.Code, category); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) DeleteMerchant(ctx context.Context, code string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM merchants WHERE code=$1 AND is_system=false`, code)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
