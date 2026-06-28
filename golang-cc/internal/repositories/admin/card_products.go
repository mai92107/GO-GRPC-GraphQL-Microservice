package admin

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListCardProducts(ctx context.Context) ([]domain.CardProduct, error) {
	rows, err := r.pool.Query(ctx, `SELECT cp.id,cp.bank_id,b.name,cp.name,cp.is_active,cp.account_tiers,
		COALESCE(cp.qualified_type,''),COALESCE(cp.selectable_type,'')
		FROM card_products cp JOIN banks b ON b.id=cp.bank_id ORDER BY b.name,cp.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CardProduct{}
	for rows.Next() {
		var x domain.CardProduct
		if err := rows.Scan(&x.ID, &x.BankID, &x.BankName, &x.Name, &x.IsActive, &x.AccountTiers, &x.QualifiedType, &x.SelectableType); err != nil {
			return nil, err
		}
		x.Activities = []domain.CardProductActivity{}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		networks, err := r.cardProductNetworks(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Networks = networks
	}
	return out, nil
}

func (r *Repository) GetCardProduct(ctx context.Context, id string, includeActivities bool) (domain.CardProduct, error) {
	var x domain.CardProduct
	err := r.pool.QueryRow(ctx, `SELECT cp.id,cp.bank_id,b.name,cp.name,cp.is_active,cp.account_tiers,
		COALESCE(cp.qualified_type,''),COALESCE(cp.selectable_type,'')
		FROM card_products cp JOIN banks b ON b.id=cp.bank_id WHERE cp.id=$1`, id).
		Scan(&x.ID, &x.BankID, &x.BankName, &x.Name, &x.IsActive, &x.AccountTiers, &x.QualifiedType, &x.SelectableType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return x, domain.ErrNotFound
		}
		return x, err
	}
	networks, err := r.cardProductNetworks(ctx, x.ID)
	if err != nil {
		return x, err
	}
	x.Networks = networks
	x.Activities = []domain.CardProductActivity{}
	if !includeActivities {
		return x, nil
	}
	activities, err := r.listActivities(ctx, x.ID)
	if err != nil {
		return x, err
	}
	for _, a := range activities {
		x.Activities = append(x.Activities, domain.CardProductActivity{ID: a.ID, Name: a.Name, StartDate: a.StartDate, EndDate: a.EndDate, IsActive: a.IsActive, SourceURL: a.SourceURL, VerifiedAt: a.VerifiedAt, NetworkIDs: a.NetworkIDs, Benefits: a.Benefits})
	}
	return x, nil
}

func (r *Repository) ListCardNetworks(ctx context.Context) ([]domain.CardNetwork, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,id,name FROM catalog.card_networks WHERE is_active ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CardNetwork{}
	for rows.Next() {
		var network domain.CardNetwork
		if err := rows.Scan(&network.ID, &network.Code, &network.Name); err != nil {
			return nil, err
		}
		out = append(out, network)
	}
	return out, rows.Err()
}

func (r *Repository) cardProductNetworks(ctx context.Context, cardProductID string) ([]domain.CardNetwork, error) {
	rows, err := r.pool.Query(ctx, `SELECT n.id,n.id,n.name FROM catalog.card_product_networks pn
		JOIN catalog.card_networks n ON n.id=pn.card_network_id
		WHERE pn.card_product_id=$1 AND n.is_active ORDER BY n.name`, cardProductID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CardNetwork{}
	for rows.Next() {
		var network domain.CardNetwork
		if err := rows.Scan(&network.ID, &network.Code, &network.Name); err != nil {
			return nil, err
		}
		out = append(out, network)
	}
	return out, rows.Err()
}

func (r *Repository) CreateCardProduct(ctx context.Context, id string, input domain.CardProductInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO card_products(id,bank_id,name,is_active,account_tiers,qualified_type,selectable_type)
		VALUES($1,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,''))`, id, input.BankID, input.Name, input.IsActive, input.AccountTiers, input.QualifiedType, input.SelectableType); err != nil {
		return err
	}
	if err := syncQualifiedCardPlans(ctx, tx, id, input.QualifiedType); err != nil {
		return err
	}
	if err := syncCardProductNetworks(ctx, tx, id, input.NetworkIDs); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *Repository) UpdateCardProduct(ctx context.Context, id string, input domain.CardProductInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE card_products SET bank_id=$2,name=$3,is_active=$4,account_tiers=$5,
		qualified_type=NULLIF($6,''),selectable_type=NULLIF($7,''),updated_at=now() WHERE id=$1`,
		id, input.BankID, input.Name, input.IsActive, input.AccountTiers, input.QualifiedType, input.SelectableType)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if err := syncQualifiedCardPlans(ctx, tx, id, input.QualifiedType); err != nil {
		return err
	}
	if err := syncCardProductNetworks(ctx, tx, id, input.NetworkIDs); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *Repository) DeleteCardProduct(ctx context.Context, id string) error {
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

func syncCardProductNetworks(ctx context.Context, tx pgx.Tx, cardProductID string, networkIDs []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM catalog.card_product_networks WHERE card_product_id=$1 AND NOT (card_network_id=ANY($2::uuid[]))`, cardProductID, networkIDs); err != nil {
		return err
	}
	for _, networkID := range networkIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO catalog.card_product_networks(card_product_id,card_network_id)
			SELECT $1::uuid,n.id FROM catalog.card_networks n WHERE n.id=$2::uuid AND n.is_active
			ON CONFLICT DO NOTHING`, cardProductID, networkID); err != nil {
			return err
		}
	}
	return nil
}
