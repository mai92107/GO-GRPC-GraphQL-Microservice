package member

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

type PreferenceWrite struct{ RewardUnitID, Weight string }

func (r *Repository) Preferences(ctx context.Context, userID string) ([]domain.RewardPreference, error) {
	rows, err := r.pool.Query(ctx, `SELECT u.id,u.code,u.name,u.symbol,COALESCE(p.weight,1)::text FROM reward_units u LEFT JOIN reward_preferences p ON p.reward_unit_id=u.id AND p.user_id=$1 ORDER BY u.code`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.RewardPreference{}
	for rows.Next() {
		var x domain.RewardPreference
		if err := rows.Scan(&x.RewardUnitID, &x.Code, &x.Name, &x.Symbol, &x.Weight); err != nil {
			return nil, err
		}
		result = append(result, x)
	}
	return result, rows.Err()
}
func (r *Repository) UpdatePreferences(ctx context.Context, userID string, items []PreferenceWrite) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, x := range items {
		if _, err = tx.Exec(ctx, `INSERT INTO reward_preferences(user_id,reward_unit_id,weight) VALUES($1,$2,$3::numeric) ON CONFLICT(user_id,reward_unit_id) DO UPDATE SET weight=EXCLUDED.weight,updated_at=now()`, userID, x.RewardUnitID, x.Weight); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
