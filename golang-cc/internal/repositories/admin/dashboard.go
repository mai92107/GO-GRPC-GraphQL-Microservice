package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) Dashboard(ctx context.Context) (domain.Dashboard, error) {
	var result domain.Dashboard
	err := r.pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM users WHERE role='member'),
		(SELECT count(*) FROM banks),
		(SELECT count(*) FROM card_products),
		(SELECT count(*) FROM card_activities WHERE is_active)`).
		Scan(&result.Members, &result.Banks, &result.Cards, &result.ActiveActivities)
	return result, err
}
