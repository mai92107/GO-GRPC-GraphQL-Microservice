package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,is_active FROM categories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Category{}
	for rows.Next() {
		var item domain.Category
		if err := rows.Scan(&item.ID, &item.Name, &item.IsActive); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateCategory(ctx context.Context, id, name string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO categories(id,name) VALUES($1,$2)`, id, name)
	return err
}

func (r *Repository) UpdateCategory(ctx context.Context, id, name string, active bool) error {
	tag, err := r.pool.Exec(ctx, `UPDATE categories SET name=$2,is_active=$3 WHERE id=$1`, id, name, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id=$1 AND id<>'general'`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
