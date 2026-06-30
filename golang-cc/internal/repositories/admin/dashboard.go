package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) Dashboard(ctx context.Context) (domain.Dashboard, error) {
	var result domain.Dashboard
	err := r.pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM identity.users WHERE role='member'),
		(SELECT count(*) FROM banks),
		(SELECT count(*) FROM catalog.card_products),
		(SELECT count(*) FROM reward.programs WHERE status='published')`).
		Scan(&result.Members, &result.Banks, &result.Cards, &result.ActiveActivities)
	return result, err
}
