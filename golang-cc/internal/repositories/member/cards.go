package member

import (
	"context"
	"errors"
	"time"

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
	Network       string
	CreditLimit   string
}

func (r *Repository) ListCards(ctx context.Context, userID string) ([]domain.MemberCard, error) {
	var result []domain.MemberCard

	err := r.db.WithContext(ctx).
		Table("public.member_cards AS mc").
		Select(`
			mc.id AS member_card_id,
			COALESCE(mc.nickname, cp.name) AS name,
			b.name AS issuer,
			COALESCE(mc.last_four, '') AS last_four,
			COALESCE(mc.credit_limit::text, '') AS credit_limit,
			mc.is_active AS is_active,
			COALESCE(cp.card_image_url, '') AS card_image_url,
			COALESCE(cp.primary_color, '') AS primary_color,
			COALESCE(cp.qualified_type, '') AS qualified_type,
			COALESCE(cp.selectable_type, '') AS selectable_type,
			COALESCE(mc.network, '') AS network
		`).
		Joins("JOIN catalog.card_products AS cp ON cp.id = mc.card_product_id").
		Joins("JOIN public.banks AS b ON b.id = cp.bank_id").
		Where("mc.user_id = ?", userID).
		Order("cp.name, mc.id").
		Scan(&result).Error

	return result, err
}
func (r *Repository) GetCard(ctx context.Context, userID, id string) (domain.MemberCardInfo, error) {
	var result domain.MemberCardInfo
	err := r.db.WithContext(ctx).
		Table("member_cards mc").
		Select(`
			mc.id AS memberCardID,
			COALESCE(NULLIF(mc.nickname, ''), cp.name) AS name,
			COALESCE(mc.nickname, '') AS nickname,
			b.name AS issuer,
			COALESCE(mc.last_four, '') AS last_four,
			mc.is_active AS is_active,
			mc.statement_day AS statement_day,
			mc.payment_due_day AS payment_due_day,
			COALESCE(mc.account_tier, '') AS account_tier,
			COALESCE(mc.credit_limit::text, '') AS credit_limit,
			COALESCE(cp.card_image_url, '') AS card_image_url,
			COALESCE(cp.primary_color, '') AS primary_color,
			COALESCE(cp.qualified_type, '') AS qualified_type,
			COALESCE(cp.selectable_type, '') AS selectable_type,
			COALESCE(mc.network, '') AS network
		`).
		Joins("JOIN card_products cp ON cp.id = mc.card_product_id").
		Joins("JOIN banks b ON b.id = cp.bank_id").
		Where("mc.id = ? AND mc.user_id = ?", id, userID).
		Scan(&result).Error

	return result, err
}

func (r *Repository) CreateCard(ctx context.Context, id, userID, cardID string, input CardWrite) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO member_cards(id,user_id,card_product_id,network,nickname,last_four,is_active,statement_day,payment_due_day,account_tier,credit_limit)
		SELECT $1,$2,cp.id,$4,NULLIF($5,''),NULLIF($6,''),$7,$8,$9,NULLIF($10,''),$11::numeric FROM card_products cp WHERE cp.id=$3
		AND $4 = ANY(catalog.card_product_network_values(cp.networks))
		AND ($10='' OR cardinality(cp.account_tiers)=0 OR $10=ANY(cp.account_tiers))`, id, userID, cardID, input.Network, input.Nickname, input.LastFour, input.IsActive, input.StatementDay, input.PaymentDueDay, input.AccountTier, input.CreditLimit)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInvalidInput
	}
	if _, err := tx.Exec(ctx, `INSERT INTO member_card_credit_limits(member_card_id, credit_limit, effective_from)
		VALUES ($1, $2::numeric, statement_timestamp())`, id, input.CreditLimit); err != nil {
		return err
	}
	if input.AccountTier != "" {
		if _, err := tx.Exec(ctx, `INSERT INTO catalog.member_card_qualification_statuses
			(id,member_card_id,card_plan_id,is_qualified,effective_from)
			SELECT md5('member-qualification:'||$1::text||':'||p.id::text)::uuid,$1::uuid,p.id,(pv.name=$2),'2000-01-01 00:00:00+00'
			FROM catalog.card_plans p JOIN catalog.card_plan_versions pv ON pv.card_plan_id=p.id
			WHERE p.card_product_id=$3 AND p.plan_type='qualified'`, id, input.AccountTier, cardID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r *Repository) UpdateCard(ctx context.Context, id, userID string, input CardWrite) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var limitChanged bool
	if err := tx.QueryRow(ctx, `SELECT mc.credit_limit IS DISTINCT FROM $4::numeric
		FROM member_cards mc
		JOIN card_products cp ON cp.id=mc.card_product_id
		WHERE mc.id=$1 AND mc.user_id=$2
		AND ($3='' OR cardinality(cp.account_tiers)=0 OR $3=ANY(cp.account_tiers))
		FOR UPDATE OF mc`, id, userID, input.AccountTier, input.CreditLimit).
		Scan(&limitChanged); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}

	tag, err := tx.Exec(ctx, `UPDATE member_cards mc SET
		network=$3,
		nickname=NULLIF($4,''),
		last_four=NULLIF($5,''),
		is_active=$6,
		statement_day=$7,
		payment_due_day=$8,
		account_tier=NULLIF($9,''),
		credit_limit=$10::numeric,
		updated_at=now()
		FROM card_products cp
		WHERE mc.id=$1 AND mc.user_id=$2 AND cp.id=mc.card_product_id
		AND $3 = ANY(catalog.card_product_network_values(cp.networks))
		AND ($9='' OR cardinality(cp.account_tiers)=0 OR $9=ANY(cp.account_tiers))`, id, userID, input.Network, input.Nickname, input.LastFour, input.IsActive, input.StatementDay, input.PaymentDueDay, input.AccountTier, input.CreditLimit)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if limitChanged {
		var changedAt time.Time
		if err := tx.QueryRow(ctx, `SELECT statement_timestamp()`).Scan(&changedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE member_card_credit_limits
			SET effective_to=$2 WHERE member_card_id=$1 AND effective_to IS NULL`, id, changedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO member_card_credit_limits(member_card_id, credit_limit, effective_from)
			VALUES ($1, $2::numeric, $3)`, id, input.CreditLimit, changedAt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
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
