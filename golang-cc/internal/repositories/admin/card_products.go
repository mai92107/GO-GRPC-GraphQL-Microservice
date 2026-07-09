package admin

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
	"gorm.io/gorm"
)

func (r *Repository) ListCards(ctx context.Context) ([]domain.Card, error) {
	var out []domain.Card
	err := r.db.WithContext(ctx).
		Table("catalog.card_products AS cp").
		Select(`
			cp.id AS id,
			cp.bank_id AS bank_id,
			b.name AS bank_name,
			cp.name AS name,
			cp.is_active AS is_active,
			COALESCE(cp.qualified_type, '') AS qualified_type,
			COALESCE(cp.selectable_type, '') AS selectable_type,
			COALESCE(cp.networks, '') AS networks
		`).
		Joins("JOIN banks AS b ON b.id = cp.bank_id").
		Group("cp.id,b.name").
		Order("b.name, cp.name").
		Scan(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetCardInfo(
	ctx context.Context,
	id string,
) (domain.CardInfo, error) {

	var x domain.CardInfo
	err := r.db.WithContext(ctx).
		Table("catalog.card_products AS cp").
		Select(`
			cp.id,
			cp.bank_id,
			b.name AS bank_name,
			cp.name,
			cp.is_active,
			COALESCE(cp.qualified_type, '') AS qualified_type,
			COALESCE(cp.selectable_type, '') AS selectable_type,
			COALESCE(string_agg(n.name, ',' ORDER BY n.name), '') AS networks
		`).
		Joins("JOIN banks AS b ON b.id = cp.bank_id").
		Joins("LEFT JOIN catalog.card_product_networks pn ON pn.card_product_id=cp.id").
		Joins("LEFT JOIN catalog.card_networks n ON n.id=pn.card_network_id").
		Where("cp.id = ?", id).
		Group("cp.id,b.name").
		Scan(&x).Error

	if err != nil {
		return x, err
	}
	if x.ID == "" {
		return x, domain.ErrNotFound
	}
	return x, nil
}

func (r *Repository) ListNetworks(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT name FROM catalog.card_networks WHERE is_active ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	networks := []string{}
	for rows.Next() {
		var network string
		if err := rows.Scan(&network); err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	return networks, rows.Err()
}

func (r *Repository) CreateCard(ctx context.Context, id string, input domain.CardInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("catalog.card_products").Create(map[string]any{
			"id":              id,
			"bank_id":         input.BankID,
			"name":            input.Name,
			"is_active":       input.IsActive,
			"qualified_type":  nullString(input.QualifiedType),
			"selectable_type": nullString(input.SelectableType),
		}).Error; err != nil {
			return err
		}
		return replaceCardNetworks(tx, id, input.Networks)
	})
}

func (r *Repository) UpdateCard(ctx context.Context, id string, input domain.CardInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"bank_id":         input.BankID,
			"name":            input.Name,
			"is_active":       input.IsActive,
			"qualified_type":  nullString(input.QualifiedType),
			"selectable_type": nullString(input.SelectableType),
			"updated_at":      time.Now(),
		}

		result := tx.
			Model(&domain.CardProduct{}).
			Where("id = ?", id).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return replaceCardNetworks(tx, id, input.Networks)
	})
}

func (r *Repository) DeleteCard(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM card_products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func syncQualifiedCardPlans(ctx context.Context, tx pgx.Tx, cardProductID, qualifiedType string) error {
	qualifiedTypes := domain.SplitCardType(qualifiedType)
	if len(qualifiedTypes) == 0 {
		_, err := tx.Exec(ctx, `UPDATE catalog.card_plans SET is_active=false
			WHERE card_product_id=$1 AND plan_type='qualified'`, cardProductID)
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE catalog.card_plans SET is_active=false
		WHERE card_product_id=$1 AND plan_type='qualified'`, cardProductID); err != nil {
		return err
	}
	for order, cardType := range qualifiedTypes {
		if _, err := tx.Exec(ctx, `INSERT INTO catalog.card_plans(id,card_product_id,plan_type,is_active,display_order)
			VALUES(md5('qualified-plan:'||$1::text||':'||$2)::uuid,$1::uuid,'qualified',true,$3)
			ON CONFLICT(id) DO UPDATE SET is_active=true,display_order=EXCLUDED.display_order`,
			cardProductID, cardType, order+1); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO catalog.card_plan_versions(
			id,card_plan_id,name,description,effective_from,published_at)
			SELECT md5('qualified-plan-version:'||plan.id::text)::uuid,plan.id,$2,'由會員自行確認是否符合銀行資格','2000-01-01 00:00:00+00',now()
			FROM (SELECT md5('qualified-plan:'||$1::text||':'||$2)::uuid AS id) plan
			ON CONFLICT DO NOTHING`, cardProductID, cardType); err != nil {
			return err
		}
	}
	return nil
}

func nullString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func replaceCardNetworks(tx *gorm.DB, cardProductID string, networks []string) error {
	if err := tx.Exec(`DELETE FROM catalog.card_product_networks WHERE card_product_id=?`, cardProductID).Error; err != nil {
		return err
	}
	for _, network := range networks {
		if err := tx.Exec(`INSERT INTO catalog.card_product_networks(card_product_id,card_network_id)
			SELECT ?::uuid,id FROM catalog.card_networks WHERE name=? AND is_active`, cardProductID, network).Error; err != nil {
			return err
		}
	}
	return nil
}
