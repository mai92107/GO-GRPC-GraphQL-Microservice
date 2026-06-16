package member

import (
	"context"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) Catalog(ctx context.Context) ([]domain.CatalogCard, error) {
	rows, err := r.pool.Query(ctx, `SELECT cp.id,cp.bank_id,b.name,cp.name,cp.is_active,cp.account_tiers FROM card_products cp JOIN banks b ON b.id=cp.bank_id WHERE cp.is_active AND b.is_active ORDER BY b.name,cp.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CatalogCard{}
	for rows.Next() {
		var x domain.CatalogCard
		if err := rows.Scan(&x.ID, &x.BankID, &x.BankName, &x.Name, &x.IsActive, &x.AccountTiers); err != nil {
			return nil, err
		}
		x.Activities = []domain.CardProductActivity{}
		out = append(out, x)
	}
	for i := range out {
		acts, err := r.catalogActivities(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Activities = acts
	}
	return out, rows.Err()
}
func (r *Repository) catalogActivities(ctx context.Context, productID string) ([]domain.CardProductActivity, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,start_date::text,end_date::text,is_active,COALESCE(source_url,''),verified_at::text FROM card_activities WHERE card_product_id=$1 ORDER BY start_date DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CardProductActivity{}
	for rows.Next() {
		var a domain.CardProductActivity
		if err := rows.Scan(&a.ID, &a.Name, &a.StartDate, &a.EndDate, &a.IsActive, &a.SourceURL, &a.VerifiedAt); err != nil {
			return nil, err
		}
		bs, err := r.catalogBenefits(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		a.Benefits = bs
		out = append(out, a)
	}
	return out, rows.Err()
}
func (r *Repository) catalogBenefits(ctx context.Context, activityID string) ([]domain.ActivityBenefit, error) {
	rows, err := r.pool.Query(ctx, `SELECT b.id,b.reward_unit_id,b.name,b.rate::text,b.monthly_cap::text,b.stack_group,b.priority,b.required_account_tiers,b.action_required,b.action_message,
		ARRAY(SELECT p.payment_method_code FROM card_activity_benefit_payment_methods p WHERE p.benefit_id=b.id ORDER BY p.payment_method_code),
		ARRAY(SELECT c.category_code FROM card_activity_benefit_categories c WHERE c.benefit_id=b.id ORDER BY c.category_code),
		ARRAY(SELECT m.merchant_code FROM card_activity_benefit_merchants m WHERE m.benefit_id=b.id ORDER BY m.merchant_code)
		FROM card_activity_benefits b WHERE b.card_activity_id=$1 AND b.is_active ORDER BY b.priority,b.name`, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ActivityBenefit{}
	for rows.Next() {
		var b domain.ActivityBenefit
		if err := rows.Scan(&b.ID, &b.RewardUnitID, &b.Name, &b.Rate, &b.MonthlyCap, &b.StackGroup, &b.Priority, &b.RequiredAccountTiers, &b.ActionRequired, &b.ActionMessage, &b.PaymentMethods, &b.CategoryCodes, &b.MerchantCodes); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
func (r *Repository) RewardUnits(ctx context.Context) ([]domain.MemberRewardUnit, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,code,name,symbol,symbol_position,twd_rate::text,precision FROM reward_units ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MemberRewardUnit{}
	for rows.Next() {
		var x domain.MemberRewardUnit
		if err := rows.Scan(&x.ID, &x.Code, &x.Name, &x.Symbol, &x.SymbolPosition, &x.TWDRate, &x.Precision); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Categories(ctx context.Context) ([]domain.MemberCategory, error) {
	rows, err := r.pool.Query(ctx, `SELECT code,name FROM categories WHERE code<>'general' AND is_active ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MemberCategory{}
	for rows.Next() {
		var x domain.MemberCategory
		if err := rows.Scan(&x.Code, &x.Name); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *Repository) PaymentMethods(ctx context.Context) ([]domain.MemberPaymentMethod, error) {
	rows, err := r.pool.Query(ctx, `SELECT code,name FROM payment_methods WHERE is_active ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MemberPaymentMethod{}
	for rows.Next() {
		var item domain.MemberPaymentMethod
		if err := rows.Scan(&item.Code, &item.Name); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) Merchants(ctx context.Context, category string) ([]domain.MemberMerchant, error) {
	rows, err := r.pool.Query(ctx, `SELECT DISTINCT m.code,m.name
		FROM merchants m JOIN merchant_categories mc ON mc.merchant_code=m.code
		WHERE m.is_active AND mc.category_code=$1 ORDER BY m.name`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MemberMerchant{}
	for rows.Next() {
		var item domain.MemberMerchant
		if err := rows.Scan(&item.Code, &item.Name); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
