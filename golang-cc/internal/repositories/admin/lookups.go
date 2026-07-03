package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListRegions(ctx context.Context) ([]domain.LookupItem, error) {
	return r.listLookupItems(ctx, "catalog.regions")
}

func (r *Repository) CreateRegion(ctx context.Context, item domain.LookupItem) error {
	return r.createLookupItem(ctx, "catalog.regions", item)
}

func (r *Repository) UpdateRegion(ctx context.Context, item domain.LookupItem) error {
	return r.updateLookupItem(ctx, "catalog.regions", item)
}

func (r *Repository) DeleteRegion(ctx context.Context, id string) error {
	return r.deleteLookupItem(ctx, "catalog.regions", id)
}

func (r *Repository) ListUserQualifications(ctx context.Context) ([]domain.LookupItem, error) {
	return r.listLookupItems(ctx, "catalog.user_qualifications")
}

func (r *Repository) CreateUserQualification(ctx context.Context, item domain.LookupItem) error {
	return r.createLookupItem(ctx, "catalog.user_qualifications", item)
}

func (r *Repository) UpdateUserQualification(ctx context.Context, item domain.LookupItem) error {
	return r.updateLookupItem(ctx, "catalog.user_qualifications", item)
}

func (r *Repository) DeleteUserQualification(ctx context.Context, id string) error {
	return r.deleteLookupItem(ctx, "catalog.user_qualifications", id)
}

func (r *Repository) listLookupItems(ctx context.Context, table string) ([]domain.LookupItem, error) {
	rows, err := r.pool.Query(ctx, "SELECT id,name,is_active FROM "+table+" ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.LookupItem{}
	for rows.Next() {
		var item domain.LookupItem
		if err := rows.Scan(&item.ID, &item.Name, &item.IsActive); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) createLookupItem(ctx context.Context, table string, item domain.LookupItem) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO "+table+"(id,name,is_active) VALUES($1,$2,$3)", item.ID, item.Name, item.IsActive)
	return err
}

func (r *Repository) updateLookupItem(ctx context.Context, table string, item domain.LookupItem) error {
	tag, err := r.pool.Exec(ctx, "UPDATE "+table+" SET name=$2,is_active=$3,updated_at=now() WHERE id=$1", item.ID, item.Name, item.IsActive)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) deleteLookupItem(ctx context.Context, table, id string) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM "+table+" WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
