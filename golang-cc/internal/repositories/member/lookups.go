package member

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) Catalog(ctx context.Context) ([]domain.CatalogCard, error) {
	rows, err := r.pool.Query(ctx, `SELECT cp.id,cp.bank_id,b.name,cp.name,COALESCE(cp.card_image_url,''),COALESCE(cp.primary_color,''),cp.is_active,cp.account_tiers,
		COALESCE(cp.qualified_type,''),COALESCE(cp.selectable_type,'')
		FROM card_products cp JOIN banks b ON b.id=cp.bank_id WHERE cp.is_active AND b.is_active ORDER BY b.name,cp.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CatalogCard{}
	for rows.Next() {
		var x domain.CatalogCard
		if err := rows.Scan(&x.ID, &x.BankID, &x.BankName, &x.Name, &x.CardImageURL, &x.PrimaryColor, &x.IsActive, &x.AccountTiers, &x.QualifiedType, &x.SelectableType); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *Repository) CatalogCard(ctx context.Context, id string) (domain.CatalogCard, error) {
	var x domain.CatalogCard
	err := r.pool.QueryRow(ctx, `SELECT cp.id,cp.bank_id,b.name,cp.name,COALESCE(cp.card_image_url,''),COALESCE(cp.primary_color,''),cp.is_active,cp.account_tiers,
		COALESCE(cp.qualified_type,''),COALESCE(cp.selectable_type,'')
		FROM card_products cp JOIN banks b ON b.id=cp.bank_id WHERE cp.id=$1 AND cp.is_active AND b.is_active`, id).
		Scan(&x.ID, &x.BankID, &x.BankName, &x.Name, &x.CardImageURL, &x.PrimaryColor, &x.IsActive, &x.AccountTiers, &x.QualifiedType, &x.SelectableType)
	if err != nil {
		return x, err
	}
	networkRows, err := r.pool.Query(ctx, `SELECT n.id,n.id,n.name FROM catalog.card_product_networks pn
		JOIN catalog.card_networks n ON n.id=pn.card_network_id
		WHERE pn.card_product_id=$1 AND n.is_active ORDER BY n.name`, x.ID)
	if err != nil {
		return x, err
	}
	for networkRows.Next() {
		var network string
		if err := networkRows.Scan(&network); err != nil {
			networkRows.Close()
			return x, err
		}
		x.Networks = append(x.Networks, network)
	}
	networkRows.Close()
	acts, err := r.catalogActivities(ctx, x.ID)
	if err != nil {
		return x, err
	}
	x.Activities = acts
	return x, nil
}

func (r *Repository) catalogActivities(ctx context.Context, productID string) ([]domain.CardProductActivity, error) {
	rows, err := r.pool.Query(ctx, `SELECT p.id,p.name,
		COALESCE((min(v.effective_from) AT TIME ZONE 'Asia/Taipei')::date::text,'2000-01-01'),
		COALESCE(((max(v.effective_to)-interval '1 microsecond') AT TIME ZONE 'Asia/Taipei')::date::text,'2099-12-31'),
		p.status='published',COALESCE(p.source_url,''),p.verified_at::date::text
		FROM reward.programs p
		LEFT JOIN reward.components c ON c.reward_program_id=p.id
		LEFT JOIN reward.component_versions v ON v.reward_component_id=c.id
		WHERE p.card_product_id=$1 AND p.status<>'archived'
		GROUP BY p.id
		ORDER BY COALESCE((min(v.effective_from) AT TIME ZONE 'Asia/Taipei')::date,DATE '2000-01-01') DESC`, productID)
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
	rows, err := r.pool.Query(ctx, `SELECT c.id,cv.reward_unit_id,cv.name,c.layer,c.display_order,cv.effect_type,cv.reward_value::text,
		cap.limit_value::text,c.stack_group,c.priority,
		COALESCE(req.qualified_names,'{}'),COALESCE(req.action_required,'none'),COALESCE(req.action_message,''),
		COALESCE(req.payment_methods,'{}'),COALESCE(req.categories,'{}'),COALESCE(req.merchants,'{}')
		FROM reward.components c
		JOIN LATERAL (
			SELECT * FROM reward.component_versions v WHERE v.reward_component_id=c.id
			ORDER BY v.effective_from DESC LIMIT 1
		) cv ON true
		LEFT JOIN LATERAL (
			SELECT v.limit_value FROM reward.component_caps cc
			JOIN reward.cap_versions v ON v.reward_cap_id=cc.reward_cap_id
			WHERE cc.reward_component_id=c.id ORDER BY v.effective_from DESC LIMIT 1
		) cap ON true
		LEFT JOIN LATERAL (
			SELECT array_remove(array_agg(DISTINCT qpv.name) FILTER (WHERE cp.plan_type='qualified'),NULL) AS qualified_names,
				CASE WHEN count(*) FILTER (WHERE cp.plan_type='selectable')>0 THEN 'app_switch'
				     WHEN count(rem.value)>0 THEN 'account_setup'
				     ELSE 'none' END AS action_required,
				COALESCE(max(
					CASE WHEN cp.plan_type='selectable' THEN
						concat('需至 APP 切換卡片方案：',spv.name,'，對應優惠：',cv.name,' ',trim(trailing '.' from trim(trailing '0' from cv.reward_value::text)),
							CASE WHEN spv.reminder_text<>'' THEN '；'||spv.reminder_text ELSE '' END)
					END
				), max(rem.value),'') AS action_message,
				array_remove(array_agg(DISTINCT pm.value),NULL) AS payment_methods,
				array_remove(array_agg(DISTINCT cat.value),NULL) AS categories,
				array_remove(array_agg(DISTINCT mer.value),NULL) AS merchants
			FROM reward.requirements rr
			JOIN reward.conditions rc ON rc.id=rr.reward_condition_id AND rc.is_active
			JOIN reward.condition_versions x ON x.reward_condition_id=rr.reward_condition_id
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'card_plan_ids') plan_id ON true
			LEFT JOIN catalog.card_plans cp ON cp.id=plan_id.value::uuid
			LEFT JOIN LATERAL (
				SELECT name FROM catalog.card_plan_versions pv WHERE pv.card_plan_id=cp.id ORDER BY pv.effective_from DESC LIMIT 1
			) qpv ON cp.plan_type='qualified'
			LEFT JOIN LATERAL (
				SELECT name,reminder_text FROM catalog.card_plan_versions pv WHERE pv.card_plan_id=cp.id ORDER BY pv.effective_from DESC LIMIT 1
			) spv ON cp.plan_type='selectable'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'payment_method_ids') pm(value) ON rc.condition_type='payment_method'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'category_ids') cat(value) ON rc.condition_type='category'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'merchant_ids') mer(value) ON rc.condition_type='merchant'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'reminder_messages') rem(value) ON rc.condition_type='channel'
			WHERE rr.reward_component_id=c.id
		) req ON true
		WHERE c.reward_program_id=$1 AND c.is_active ORDER BY c.layer,c.display_order,c.priority,cv.name`, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ActivityBenefit{}
	for rows.Next() {
		var b domain.ActivityBenefit
		if err := rows.Scan(&b.ID, &b.RewardUnitID, &b.Name, &b.Layer, &b.DisplayOrder, &b.EffectType, &b.RewardValue, &b.MonthlyCap, &b.StackGroup, &b.Priority, &b.QualifiedType, &b.ActionRequired, &b.ActionMessage, &b.PaymentMethods, &b.CategoryIDs, &b.MerchantIDs); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
func (r *Repository) RewardUnits(ctx context.Context) ([]domain.MemberRewardUnit, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,symbol,symbol_position,twd_rate::text,precision FROM reward_units ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MemberRewardUnit{}
	for rows.Next() {
		var x domain.MemberRewardUnit
		if err := rows.Scan(&x.ID, &x.Name, &x.Symbol, &x.SymbolPosition, &x.TWDRate, &x.Precision); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Categories(ctx context.Context) ([]domain.MemberCategory, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name FROM categories WHERE id<>'general' AND is_active ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MemberCategory{}
	for rows.Next() {
		var x domain.MemberCategory
		if err := rows.Scan(&x.ID, &x.Name); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *Repository) PaymentMethods(ctx context.Context, userID string) ([]domain.MemberPaymentMethod, error) {
	rows, err := r.pool.Query(ctx, `SELECT p.id,p.name,p.type,
		CASE WHEN p.type IN ('physical_card','online_card') THEN true ELSE COALESCE(up.is_available,false) END
		FROM payment_methods p LEFT JOIN member_profile.user_payment_methods up
		ON up.payment_method_id=p.id AND up.user_id=$1
		WHERE p.is_active ORDER BY p.display_order,p.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MemberPaymentMethod{}
	for rows.Next() {
		var item domain.MemberPaymentMethod
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.IsAvailable); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) Merchants(ctx context.Context, category string) ([]domain.MemberMerchant, error) {
	rows, err := r.pool.Query(ctx, `SELECT DISTINCT m.id,m.name
		FROM merchants m JOIN merchant_categories mc ON mc.merchant_id=m.id
		WHERE m.is_active AND mc.category_id=$1 ORDER BY m.name`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MemberMerchant{}
	for rows.Next() {
		var item domain.MemberMerchant
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
