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
		(SELECT COUNT(*) FROM reward.activities WHERE is_active AND effective_from <= now() AND (effective_to IS NULL OR effective_to > now()))`).
		Scan(&result.Members, &result.Banks, &result.Cards, &result.ActiveActivities)
	return result, err
}
