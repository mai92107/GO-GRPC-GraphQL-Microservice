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
}

func (r *Repository) ListCards(ctx context.Context, userID string) ([]domain.MemberCard, error) {
	rows, err := r.pool.Query(ctx, `SELECT mc.id,mc.card_product_id,COALESCE(NULLIF(mc.nickname,''),cp.name),b.name,
		COALESCE(mc.last_four,''),mc.is_active,mc.statement_day,mc.payment_due_day,COALESCE(mc.account_tier,'')
		FROM member_cards mc JOIN card_products cp ON cp.id=mc.card_product_id JOIN banks b ON b.id=cp.bank_id
		WHERE mc.user_id=$1 ORDER BY cp.name,mc.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.MemberCard{}
	for rows.Next() {
		var item domain.MemberCard
		if err := rows.Scan(&item.ID, &item.CardProductID, &item.Name, &item.Issuer, &item.LastFour, &item.IsActive, &item.StatementDay, &item.PaymentDueDay, &item.AccountTier); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) GetCard(ctx context.Context, userID, id string) (domain.MemberCard, error) {
	var item domain.MemberCard
	err := r.pool.QueryRow(ctx, `SELECT mc.id,mc.card_product_id,COALESCE(NULLIF(mc.nickname,''),cp.name),b.name,COALESCE(mc.last_four,''),mc.is_active,mc.statement_day,mc.payment_due_day,COALESCE(mc.account_tier,'')
		FROM member_cards mc JOIN card_products cp ON cp.id=mc.card_product_id JOIN banks b ON b.id=cp.bank_id WHERE mc.id=$1 AND mc.user_id=$2`, id, userID).
		Scan(&item.ID, &item.CardProductID, &item.Name, &item.Issuer, &item.LastFour, &item.IsActive, &item.StatementDay, &item.PaymentDueDay, &item.AccountTier)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, domain.ErrNotFound
	}
	return item, err
}

func (r *Repository) CreateCard(ctx context.Context, id, userID, productID string, input CardWrite) error {
	tag, err := r.pool.Exec(ctx, `INSERT INTO member_cards(id,user_id,card_product_id,nickname,last_four,is_active,statement_day,payment_due_day,account_tier)
		SELECT $1,$2,cp.id,NULLIF($4,''),NULLIF($5,''),$6,$7,$8,NULLIF($9,'') FROM card_products cp WHERE cp.id=$3
		AND (cardinality(cp.account_tiers)=0 OR $9=ANY(cp.account_tiers))`, id, userID, productID, input.Nickname, input.LastFour, input.IsActive, input.StatementDay, input.PaymentDueDay, input.AccountTier)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInvalidInput
	}
	return nil
}
func (r *Repository) UpdateCard(ctx context.Context, id, userID string, input CardWrite) error {
	tag, err := r.pool.Exec(ctx, `UPDATE member_cards mc SET nickname=NULLIF($3,''),last_four=NULLIF($4,''),is_active=$5,statement_day=$6,payment_due_day=$7,account_tier=NULLIF($8,''),updated_at=now()
		FROM card_products cp WHERE mc.id=$1 AND mc.user_id=$2 AND cp.id=mc.card_product_id AND (cardinality(cp.account_tiers)=0 OR $8=ANY(cp.account_tiers))`, id, userID, input.Nickname, input.LastFour, input.IsActive, input.StatementDay, input.PaymentDueDay, input.AccountTier)
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
