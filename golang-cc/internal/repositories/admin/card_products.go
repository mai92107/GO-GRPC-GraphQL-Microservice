package admin

import (
	"context"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListCardProducts(ctx context.Context) ([]domain.CardProduct, error) {
	rows, err := r.pool.Query(ctx, `SELECT cp.id,cp.bank_id,b.name,cp.name,cp.is_active,cp.account_tiers FROM card_products cp JOIN banks b ON b.id=cp.bank_id ORDER BY b.name,cp.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CardProduct{}
	for rows.Next() {
		var x domain.CardProduct
		if err := rows.Scan(&x.ID, &x.BankID, &x.BankName, &x.Name, &x.IsActive, &x.AccountTiers); err != nil {
			return nil, err
		}
		x.Activities = []domain.CardProductActivity{}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	activities, err := r.ListActivities(ctx)
	if err != nil {
		return nil, err
	}
	for i := range out {
		for _, a := range activities {
			if a.CardProductID == out[i].ID {
				out[i].Activities = append(out[i].Activities, domain.CardProductActivity{ID: a.ID, Name: a.Name, StartDate: a.StartDate, EndDate: a.EndDate, IsActive: a.IsActive, SourceURL: a.SourceURL, VerifiedAt: a.VerifiedAt, Benefits: a.Benefits})
			}
		}
	}
	return out, nil
}
func (r *Repository) CreateCardProduct(ctx context.Context, id string, input domain.CardProductInput) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO card_products(id,bank_id,name,is_active,account_tiers) VALUES($1,$2,$3,$4,$5)`, id, input.BankID, input.Name, input.IsActive, input.AccountTiers)
	return err
}
func (r *Repository) UpdateCardProduct(ctx context.Context, id string, input domain.CardProductInput) error {
	tag, err := r.pool.Exec(ctx, `UPDATE card_products SET bank_id=$2,name=$3,is_active=$4,account_tiers=$5,updated_at=now() WHERE id=$1`, id, input.BankID, input.Name, input.IsActive, input.AccountTiers)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
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
