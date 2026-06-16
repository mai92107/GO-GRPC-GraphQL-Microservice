package admin

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListActivities(ctx context.Context) ([]domain.Activity, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,card_product_id,name,start_date::text,end_date::text,is_active,
		COALESCE(source_url,''),verified_at::text,shared_monthly_caps FROM card_activities ORDER BY start_date DESC,name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Activity{}
	for rows.Next() {
		var item domain.Activity
		var caps []byte
		if err := rows.Scan(&item.ID, &item.CardProductID, &item.Name, &item.StartDate, &item.EndDate, &item.IsActive, &item.SourceURL, &item.VerifiedAt, &caps); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(caps, &item.SharedMonthlyCaps); err != nil {
			return nil, err
		}
		item.Benefits, err = loadBenefits(ctx, r.pool, item.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func loadBenefits(ctx context.Context, q queryer, activityID string) ([]domain.ActivityBenefit, error) {
	rows, err := q.Query(ctx, `SELECT b.id,b.reward_unit_id,b.name,b.rate::text,b.monthly_cap::text,b.stack_group,b.priority,
		b.required_account_tiers,b.action_required,b.action_message,
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

func (r *Repository) CreateActivity(ctx context.Context, item domain.Activity) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	caps, _ := json.Marshal(item.SharedMonthlyCaps)
	if _, err = tx.Exec(ctx, `INSERT INTO card_activities(id,card_product_id,name,start_date,end_date,is_active,source_url,verified_at,shared_monthly_caps)
		VALUES($1,$2,$3,$4::date,$5::date,$6,NULLIF($7,''),$8::date,$9::jsonb)`, item.ID, item.CardProductID, item.Name, item.StartDate, item.EndDate, item.IsActive, item.SourceURL, item.VerifiedAt, caps); err != nil {
		return err
	}
	if err = insertBenefits(ctx, tx, item); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) UpdateActivity(ctx context.Context, item domain.Activity) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	caps, _ := json.Marshal(item.SharedMonthlyCaps)
	tag, err := tx.Exec(ctx, `UPDATE card_activities SET card_product_id=$2,name=$3,start_date=$4::date,end_date=$5::date,is_active=$6,
		source_url=NULLIF($7,''),verified_at=$8::date,shared_monthly_caps=$9::jsonb,updated_at=now() WHERE id=$1`,
		item.ID, item.CardProductID, item.Name, item.StartDate, item.EndDate, item.IsActive, item.SourceURL, item.VerifiedAt, caps)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if err = insertBenefits(ctx, tx, item); err != nil {
		return err
	}
	ids := make([]string, 0, len(item.Benefits))
	for _, b := range item.Benefits {
		ids = append(ids, b.ID)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM card_activity_benefits WHERE card_activity_id=$1 AND NOT (id=ANY($2::uuid[]))`, item.ID, ids); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) DeleteActivity(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM card_activities WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func insertBenefits(ctx context.Context, tx pgx.Tx, item domain.Activity) error {
	for _, b := range item.Benefits {
		if _, err := tx.Exec(ctx, `INSERT INTO card_activity_benefits(id,card_activity_id,reward_unit_id,name,rate,monthly_cap,stack_group,priority,required_account_tiers,action_required,action_message)
			VALUES($1,$2,$3,$4,$5::numeric,$6::numeric,$7,$8,$9,$10,$11)
			ON CONFLICT(id) DO UPDATE SET reward_unit_id=EXCLUDED.reward_unit_id,name=EXCLUDED.name,rate=EXCLUDED.rate,monthly_cap=EXCLUDED.monthly_cap,stack_group=EXCLUDED.stack_group,priority=EXCLUDED.priority,required_account_tiers=EXCLUDED.required_account_tiers,action_required=EXCLUDED.action_required,action_message=EXCLUDED.action_message,updated_at=now()`,
			b.ID, item.ID, b.RewardUnitID, b.Name, b.Rate, b.MonthlyCap, b.StackGroup, b.Priority, b.RequiredAccountTiers, b.ActionRequired, b.ActionMessage); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM card_activity_benefit_categories WHERE benefit_id=$1`, b.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM card_activity_benefit_merchants WHERE benefit_id=$1`, b.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM card_activity_benefit_payment_methods WHERE benefit_id=$1`, b.ID); err != nil {
			return err
		}
		for _, c := range b.CategoryCodes {
			if _, err := tx.Exec(ctx, `INSERT INTO card_activity_benefit_categories VALUES($1,$2)`, b.ID, c); err != nil {
				return err
			}
		}
		for _, m := range b.MerchantCodes {
			if _, err := tx.Exec(ctx, `INSERT INTO card_activity_benefit_merchants VALUES($1,$2)`, b.ID, m); err != nil {
				return err
			}
		}
		for _, method := range b.PaymentMethods {
			if _, err := tx.Exec(ctx, `INSERT INTO card_activity_benefit_payment_methods VALUES($1,$2)`, b.ID, method); err != nil {
				return err
			}
		}
	}
	return nil
}
