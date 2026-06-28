package member

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

type CardWrite struct {
	Nickname      string
	LastFour      string
	IsActive      bool
	StatementDay  *int
	PaymentDueDay *int
	AccountTier   string
	CardNetworkID string
}

func (r *Repository) ListCards(ctx context.Context, userID string) ([]domain.MemberCard, error) {
	rows, err := r.pool.Query(ctx, `SELECT mc.id,mc.card_product_id,COALESCE(NULLIF(mc.nickname,''),cp.name),b.name,
		COALESCE(mc.last_four,''),mc.is_active,mc.statement_day,mc.payment_due_day,COALESCE(mc.account_tier,''),
		COALESCE(cp.card_image_url,''),COALESCE(cp.primary_color,''),COALESCE(cp.qualified_type,''),COALESCE(cp.selectable_type,''),
		COALESCE(n.id::text,''),COALESCE(n.id,''),COALESCE(n.name,'')
		FROM member_cards mc JOIN card_products cp ON cp.id=mc.card_product_id JOIN banks b ON b.id=cp.bank_id
		LEFT JOIN catalog.card_networks n ON n.id=mc.card_network_id
		WHERE mc.user_id=$1 ORDER BY cp.name,mc.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.MemberCard{}
	for rows.Next() {
		var item domain.MemberCard
		if err := rows.Scan(&item.ID, &item.CardProductID, &item.Name, &item.Issuer, &item.LastFour, &item.IsActive, &item.StatementDay, &item.PaymentDueDay, &item.AccountTier, &item.CardImageURL, &item.PrimaryColor, &item.QualifiedType, &item.SelectableType, &item.NetworkID, &item.NetworkCode, &item.NetworkName); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) GetCard(ctx context.Context, userID, id string) (domain.MemberCard, error) {
	var item domain.MemberCard
	err := r.pool.QueryRow(ctx, `SELECT mc.id,mc.card_product_id,COALESCE(NULLIF(mc.nickname,''),cp.name),b.name,COALESCE(mc.last_four,''),mc.is_active,mc.statement_day,mc.payment_due_day,COALESCE(mc.account_tier,''),
		COALESCE(cp.card_image_url,''),COALESCE(cp.primary_color,''),COALESCE(cp.qualified_type,''),COALESCE(cp.selectable_type,''),
		COALESCE(n.id::text,''),COALESCE(n.id,''),COALESCE(n.name,'')
		FROM member_cards mc JOIN card_products cp ON cp.id=mc.card_product_id JOIN banks b ON b.id=cp.bank_id
		LEFT JOIN catalog.card_networks n ON n.id=mc.card_network_id WHERE mc.id=$1 AND mc.user_id=$2`, id, userID).
		Scan(&item.ID, &item.CardProductID, &item.Name, &item.Issuer, &item.LastFour, &item.IsActive, &item.StatementDay, &item.PaymentDueDay, &item.AccountTier, &item.CardImageURL, &item.PrimaryColor, &item.QualifiedType, &item.SelectableType, &item.NetworkID, &item.NetworkCode, &item.NetworkName)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, domain.ErrNotFound
	}
	return item, err
}

func (r *Repository) CreateCard(ctx context.Context, id, userID, productID string, input CardWrite) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO member_cards(id,user_id,card_product_id,card_network_id,nickname,last_four,is_active,statement_day,payment_due_day,account_tier)
		SELECT $1,$2,cp.id,NULLIF($4,'')::uuid,NULLIF($5,''),NULLIF($6,''),$7,$8,$9,NULLIF($10,'') FROM card_products cp WHERE cp.id=$3
		AND ($10='' OR cardinality(cp.account_tiers)=0 OR $10=ANY(cp.account_tiers))`, id, userID, productID, input.CardNetworkID, input.Nickname, input.LastFour, input.IsActive, input.StatementDay, input.PaymentDueDay, input.AccountTier)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInvalidInput
	}
	if input.AccountTier != "" {
		if _, err := tx.Exec(ctx, `INSERT INTO catalog.member_card_qualification_statuses
			(id,member_card_id,card_plan_id,is_qualified,effective_from)
			SELECT md5('member-qualification:'||$1::text||':'||p.id::text)::uuid,$1::uuid,p.id,(pv.name=$2),'2000-01-01 00:00:00+00'
			FROM catalog.card_plans p JOIN catalog.card_plan_versions pv ON pv.card_plan_id=p.id
			WHERE p.card_product_id=$3 AND p.plan_type='qualified'`, id, input.AccountTier, productID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r *Repository) UpdateCard(ctx context.Context, id, userID string, input CardWrite) error {
	tag, err := r.pool.Exec(ctx, `UPDATE member_cards mc SET card_network_id=NULLIF($3,'')::uuid,nickname=NULLIF($4,''),last_four=NULLIF($5,''),is_active=$6,statement_day=$7,payment_due_day=$8,account_tier=NULLIF($9,''),updated_at=now()
		FROM card_products cp WHERE mc.id=$1 AND mc.user_id=$2 AND cp.id=mc.card_product_id AND ($9='' OR cardinality(cp.account_tiers)=0 OR $9=ANY(cp.account_tiers))`, id, userID, input.CardNetworkID, input.Nickname, input.LastFour, input.IsActive, input.StatementDay, input.PaymentDueDay, input.AccountTier)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *Repository) DeleteCard(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM member_cards WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
