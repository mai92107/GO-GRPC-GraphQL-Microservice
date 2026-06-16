package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListRewardUnits(ctx context.Context) ([]domain.RewardUnit, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,code,name,symbol,symbol_position,twd_rate::text,precision FROM reward_units ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.RewardUnit{}
	for rows.Next() {
		var item domain.RewardUnit
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Symbol, &item.SymbolPosition, &item.TWDRate, &item.Precision); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateRewardUnit(ctx context.Context, id string, input domain.RewardUnitInput) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO reward_units(id,code,name,symbol,symbol_position,twd_rate,precision,is_system) VALUES($1,$2,$3,$4,$5,$6::numeric,$7,false)`, id, input.Code, input.Name, input.Symbol, input.SymbolPosition, input.TWDRate, input.Precision)
	return err
}

func (r *Repository) UpdateRewardUnit(ctx context.Context, id string, input domain.RewardUnitInput) error {
	tag, err := r.pool.Exec(ctx, `UPDATE reward_units SET code=$2,name=$3,symbol=$4,symbol_position=$5,twd_rate=$6::numeric,precision=$7,updated_at=now() WHERE id=$1`, id, input.Code, input.Name, input.Symbol, input.SymbolPosition, input.TWDRate, input.Precision)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteRewardUnit(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM reward_units WHERE id=$1 AND is_system=false`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
