package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT code,name,is_active FROM categories ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Category{}
	for rows.Next() {
		var item domain.Category
		if err := rows.Scan(&item.Code, &item.Name, &item.IsActive); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateCategory(ctx context.Context, code, name string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO categories(code,name) VALUES($1,$2)`, code, name)
	return err
}

func (r *Repository) UpdateCategory(ctx context.Context, code, name string, active bool) error {
	tag, err := r.pool.Exec(ctx, `UPDATE categories SET name=$2,is_active=$3 WHERE code=$1`, code, name, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, code string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE code=$1 AND code<>'general'`, code)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
