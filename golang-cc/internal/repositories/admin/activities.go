package admin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) ListActivities(ctx context.Context) ([]domain.Activity, error) {
	return r.listActivities(ctx, "")
}

func (r *Repository) listActivities(ctx context.Context, cardProductID string) ([]domain.Activity, error) {
	rows, err := r.pool.Query(ctx, `SELECT p.id,p.card_product_id,p.name,
		COALESCE((min(v.effective_from) AT TIME ZONE 'Asia/Taipei')::date::text,'2000-01-01'),
		COALESCE(((max(v.effective_to)-interval '1 microsecond') AT TIME ZONE 'Asia/Taipei')::date::text,'2099-12-31'),
		p.status='published',COALESCE(p.source_url,''),p.verified_at::date::text
		FROM reward.programs p
		LEFT JOIN reward.components c ON c.reward_program_id=p.id
		LEFT JOIN reward.component_versions v ON v.reward_component_id=c.id
		WHERE p.status<>'archived' AND ($1='' OR p.card_product_id=$1::uuid)
		GROUP BY p.id
		ORDER BY COALESCE((min(v.effective_from) AT TIME ZONE 'Asia/Taipei')::date,DATE '2000-01-01') DESC,p.name`, cardProductID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Activity{}
	for rows.Next() {
		var item domain.Activity
		if err := rows.Scan(&item.ID, &item.CardProductID, &item.Name, &item.StartDate, &item.EndDate, &item.IsActive, &item.SourceURL, &item.VerifiedAt); err != nil {
			return nil, err
		}
		item.SharedMonthlyCaps = map[string]string{}
		item.NetworkIDs, err = loadProgramNetworkIDs(ctx, r.pool, item.ID)
		if err != nil {
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

func loadBenefits(ctx context.Context, q queryer, programID string) ([]domain.ActivityBenefit, error) {
	rows, err := q.Query(ctx, `SELECT c.id,cv.reward_unit_id,cv.name,
		c.layer,c.display_order,cv.effect_type,cv.reward_value::text,
		cap.limit_value::text,c.stack_group,c.priority,
		COALESCE(req.qualified_names,'{}'),COALESCE(req.selectable_name,''),
		COALESCE(req.action_required,'none'),COALESCE(req.action_message,''),
		COALESCE(req.payment_methods,'{}'),COALESCE(req.categories,'{}'),COALESCE(req.merchants,'{}')
		FROM reward.components c
		JOIN LATERAL (
			SELECT * FROM reward.component_versions v WHERE v.reward_component_id=c.id
			ORDER BY v.effective_from DESC LIMIT 1
		) cv ON true
		LEFT JOIN LATERAL (
			SELECT v.limit_value FROM reward.component_caps cc
			JOIN reward.cap_versions v ON v.reward_cap_id=cc.reward_cap_id
			WHERE cc.reward_component_id=c.id
			ORDER BY v.effective_from DESC LIMIT 1
		) cap ON true
		LEFT JOIN LATERAL (
			SELECT
				array_remove(array_agg(DISTINCT qpv.name) FILTER (WHERE cp.plan_type='qualified'),NULL) AS qualified_names,
				min(spv.name) FILTER (WHERE cp.plan_type='selectable') AS selectable_name,
				CASE WHEN count(*) FILTER (WHERE cp.plan_type='selectable')>0 THEN 'app_switch'
				     WHEN count(rem.value)>0 THEN 'account_setup'
				     ELSE 'none' END AS action_required,
				COALESCE(max(spv.reminder_text) FILTER (WHERE cp.plan_type='selectable' AND spv.reminder_text<>''), max(rem.value),'') AS action_message,
				array_remove(array_agg(DISTINCT pm.value),NULL) AS payment_methods,
				array_remove(array_agg(DISTINCT cat.value),NULL) AS categories,
				array_remove(array_agg(DISTINCT mer.value),NULL) AS merchants
			FROM reward.requirements rr
			JOIN reward.conditions rc ON rc.id=rr.reward_condition_id AND rc.is_active
			JOIN reward.condition_versions x ON x.reward_condition_id=rr.reward_condition_id
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'card_plan_ids') plan_id ON true
			LEFT JOIN catalog.card_plans cp ON cp.id=plan_id.value::uuid
			LEFT JOIN LATERAL (
				SELECT name FROM catalog.card_plan_versions pv WHERE pv.card_plan_id=cp.id
				ORDER BY pv.effective_from DESC LIMIT 1
			) qpv ON cp.plan_type='qualified'
			LEFT JOIN LATERAL (
				SELECT name,reminder_text FROM catalog.card_plan_versions pv WHERE pv.card_plan_id=cp.id
				ORDER BY pv.effective_from DESC LIMIT 1
			) spv ON cp.plan_type='selectable'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'payment_method_ids') pm(value) ON rc.condition_type='payment_method'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'category_ids') cat(value) ON rc.condition_type='category'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'merchant_ids') mer(value) ON rc.condition_type='merchant'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'reminder_messages') rem(value) ON rc.condition_type='channel'
			WHERE rr.reward_component_id=c.id
		) req ON true
		WHERE c.reward_program_id=$1 AND c.is_active
		ORDER BY c.layer,c.display_order,c.priority,cv.name`, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ActivityBenefit{}
	for rows.Next() {
		var b domain.ActivityBenefit
		if err := rows.Scan(&b.ID, &b.RewardUnitID, &b.Name, &b.Layer, &b.DisplayOrder, &b.EffectType, &b.RewardValue, &b.MonthlyCap, &b.StackGroup, &b.Priority, &b.QualifiedType, &b.SelectableType, &b.ActionRequired, &b.ActionMessage, &b.PaymentMethods, &b.CategoryIDs, &b.MerchantIDs); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func loadProgramNetworkIDs(ctx context.Context, q queryer, programID string) ([]string, error) {
	rows, err := q.Query(ctx, `SELECT DISTINCT value::text
		FROM reward.components c
		JOIN reward.requirements rr ON rr.reward_component_id=c.id
		JOIN reward.conditions rc ON rc.id=rr.reward_condition_id AND rc.condition_type='card_network'
		JOIN reward.condition_versions cv ON cv.reward_condition_id=rc.id
		CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'card_network_ids') value
		WHERE c.reward_program_id=$1
		ORDER BY value::text`, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *Repository) CreateActivity(ctx context.Context, item domain.Activity) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := validateActivityDateRange(ctx, tx, item); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO reward.programs(id,card_product_id,name,source_url,verified_at,status)
		VALUES($1::uuid,$2::uuid,$3,NULLIF($4,''),$5::date,
		CASE WHEN $6 THEN 'published' ELSE 'draft' END)`,
		item.ID, item.CardProductID, item.Name, item.SourceURL, item.VerifiedAt, item.IsActive); err != nil {
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
	if err := validateActivityDateRange(ctx, tx, item); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE reward.programs SET card_product_id=$2,name=$3,source_url=NULLIF($4,''),
		verified_at=$5::date,status=CASE WHEN $6 THEN 'published' ELSE 'draft' END,updated_at=now()
		WHERE id=$1`, item.ID, item.CardProductID, item.Name, item.SourceURL, item.VerifiedAt, item.IsActive)
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
	if _, err = tx.Exec(ctx, `UPDATE reward.components SET is_active=false
		WHERE reward_program_id=$1 AND NOT (id=ANY($2::uuid[]))`, item.ID, ids); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) DeleteActivity(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE reward.programs SET status='archived',updated_at=now() WHERE id=$1`, id)
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
		if b.ActionRequired == "app_switch" {
			if err := validateSelectableBenefit(ctx, tx, item, b); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reward.components(id,reward_program_id,stack_group,priority,is_active,layer,display_order)
			VALUES($1::uuid,$2::uuid,$3,$4,$5,true,$6,$7,$8)
			ON CONFLICT(id) DO UPDATE SET reward_program_id=EXCLUDED.reward_program_id,stack_group=EXCLUDED.stack_group,
				priority=EXCLUDED.priority,is_active=true,
				layer=EXCLUDED.layer,display_order=EXCLUDED.display_order`,
			b.ID, item.ID, b.StackGroup, b.Priority, b.Layer, b.DisplayOrder); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reward.component_versions(
			id,reward_component_id,reward_unit_id,name,effect_type,reward_value,effective_from,effective_to,published_at)
			VALUES(md5('component-version:'||$1::text)::uuid,$1::uuid,$2::uuid,$3,$4,$5::numeric,
				$6::date::timestamp AT TIME ZONE 'Asia/Taipei',
				($7::date+1)::timestamp AT TIME ZONE 'Asia/Taipei',now())
			ON CONFLICT(id) DO UPDATE SET reward_unit_id=EXCLUDED.reward_unit_id,name=EXCLUDED.name,
				effect_type=EXCLUDED.effect_type,reward_value=EXCLUDED.reward_value,
				effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to`,
			b.ID, b.RewardUnitID, b.Name, b.EffectType, b.RewardValue, item.StartDate, item.EndDate); err != nil {
			return err
		}
		if err := syncBenefitConditions(ctx, tx, item, b); err != nil {
			return err
		}
		if err := syncBenefitCap(ctx, tx, item, b); err != nil {
			return err
		}
	}
	return nil
}

func syncBenefitCap(ctx context.Context, tx pgx.Tx, item domain.Activity, benefit domain.ActivityBenefit) error {
	if benefit.MonthlyCap == nil || *benefit.MonthlyCap == "" {
		_, err := tx.Exec(ctx, `UPDATE reward.caps SET is_active=false WHERE id=md5('cap:'||$1::text)::uuid`, benefit.ID)
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward.caps(id,reward_program_id,name,scope,is_active)
		VALUES(md5('cap:'||$1::text)::uuid,$2::uuid,$3||'每月上限','component',true)
		ON CONFLICT(id) DO UPDATE SET reward_program_id=EXCLUDED.reward_program_id,name=EXCLUDED.name,is_active=true`,
		benefit.ID, item.ID, benefit.Name); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward.cap_versions(id,reward_cap_id,cap_type,limit_value,reward_unit_id,period_type,effective_from,effective_to)
		VALUES(md5('cap-version:'||$1::text)::uuid,md5('cap:'||$1::text)::uuid,'reward_amount',$2::numeric,$3::uuid,'calendar_month',
			$4::date::timestamp AT TIME ZONE 'Asia/Taipei',($5::date+1)::timestamp AT TIME ZONE 'Asia/Taipei')
		ON CONFLICT(id) DO UPDATE SET limit_value=EXCLUDED.limit_value,reward_unit_id=EXCLUDED.reward_unit_id,
			effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to`,
		benefit.ID, *benefit.MonthlyCap, benefit.RewardUnitID, item.StartDate, item.EndDate); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO reward.component_caps(reward_component_id,reward_cap_id,effective_from,effective_to)
		VALUES($1::uuid,md5('cap:'||$1::text)::uuid,$2::date::timestamp AT TIME ZONE 'Asia/Taipei',($3::date+1)::timestamp AT TIME ZONE 'Asia/Taipei')
		ON CONFLICT(reward_component_id,reward_cap_id,effective_from) DO UPDATE SET effective_to=EXCLUDED.effective_to`,
		benefit.ID, item.StartDate, item.EndDate)
	return err
}

func validateActivityDateRange(ctx context.Context, tx pgx.Tx, item domain.Activity) error {
	var duplicate bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM reward.programs existing
		JOIN reward.components c ON c.reward_program_id=existing.id
		JOIN reward.component_versions v ON v.reward_component_id=c.id
		WHERE existing.card_product_id=$1 AND existing.id<>$2 AND existing.status='published'
		AND tstzrange(v.effective_from,v.effective_to,'[)') &&
			tstzrange($3::date::timestamp AT TIME ZONE 'Asia/Taipei',($4::date+1)::timestamp AT TIME ZONE 'Asia/Taipei','[)')
	)`, item.CardProductID, item.ID, item.StartDate, item.EndDate).Scan(&duplicate); err != nil {
		return err
	}
	if duplicate {
		return domain.ErrInvalidInput
	}
	return nil
}

func validateSelectableBenefit(ctx context.Context, tx pgx.Tx, item domain.Activity, benefit domain.ActivityBenefit) error {
	selectableName := strings.TrimSpace(benefit.SelectableType)
	if selectableName == "" {
		selectableName = benefit.Name
	}
	var duplicate bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1
		FROM catalog.card_plans p
		JOIN catalog.card_plan_versions pv ON pv.card_plan_id=p.id
		WHERE p.card_product_id=$1 AND p.plan_type='selectable' AND p.is_active
		AND p.id<>md5('selectable-plan:'||$2::text)::uuid
		AND pv.name=$3
		AND tstzrange(pv.effective_from,pv.effective_to,'[)') &&
			tstzrange($4::date::timestamp AT TIME ZONE 'Asia/Taipei',($5::date+1)::timestamp AT TIME ZONE 'Asia/Taipei','[)')
	)`, item.CardProductID, benefit.ID, selectableName, item.StartDate, item.EndDate).Scan(&duplicate); err != nil {
		return err
	}
	if duplicate {
		return domain.ErrInvalidInput
	}
	return nil
}

func syncBenefitConditions(ctx context.Context, tx pgx.Tx, item domain.Activity, benefit domain.ActivityBenefit) error {
	for _, kind := range []string{"category", "payment", "merchant", "network", "qualified", "selectable", "reminder"} {
		if _, err := tx.Exec(ctx, `DELETE FROM reward.requirements WHERE id=md5($1||'-requirement:'||$2::text)::uuid`, kind, benefit.ID); err != nil {
			return err
		}
	}
	if err := syncArrayCondition(ctx, tx, benefit.ID, "category", "category", "category_ids", benefit.Name+"／消費分類", benefit.CategoryIDs, item.StartDate, item.EndDate, 10); err != nil {
		return err
	}
	if err := syncArrayCondition(ctx, tx, benefit.ID, "payment", "payment_method", "payment_method_ids", benefit.Name+"／支付方式", benefit.PaymentMethods, item.StartDate, item.EndDate, 20); err != nil {
		return err
	}
	if err := syncArrayCondition(ctx, tx, benefit.ID, "merchant", "merchant", "merchant_ids", benefit.Name+"／指定店家", benefit.MerchantIDs, item.StartDate, item.EndDate, 25); err != nil {
		return err
	}
	if err := syncArrayCondition(ctx, tx, benefit.ID, "network", "card_network", "card_network_ids", item.Name+"／卡組織", item.NetworkIDs, item.StartDate, item.EndDate, 15); err != nil {
		return err
	}
	if err := syncPlanRequirements(ctx, tx, item, benefit); err != nil {
		return err
	}
	if benefit.ActionRequired != "none" && benefit.ActionRequired != "app_switch" && strings.TrimSpace(benefit.ActionMessage) != "" {
		if err := syncReminderCondition(ctx, tx, item, benefit); err != nil {
			return err
		}
	}
	return nil
}

func syncArrayCondition(ctx context.Context, tx pgx.Tx, componentID, key, conditionType, jsonKey, name string, values []string, startDate, endDate string, order int) error {
	if len(values) == 0 {
		_, err := tx.Exec(ctx, `UPDATE reward.conditions SET is_active=false WHERE id=md5($1||'-condition:'||$2::text)::uuid`, key, componentID)
		return err
	}
	configuration, err := json.Marshal(map[string]any{jsonKey: values})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward.conditions(id,condition_type,name,is_active)
		VALUES(md5($1||'-condition:'||$2::text)::uuid,$3,$4,true)
		ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,is_active=true`, key, componentID, conditionType, name); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
		VALUES(md5($1||'-condition-version:'||$2::text)::uuid,md5($1||'-condition:'||$2::text)::uuid,'in',$3::jsonb,$4,
			$5::date::timestamp AT TIME ZONE 'Asia/Taipei',($6::date+1)::timestamp AT TIME ZONE 'Asia/Taipei')
		ON CONFLICT(id) DO UPDATE SET configuration_json=EXCLUDED.configuration_json,description=EXCLUDED.description,
			effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to`, key, componentID, configuration, joinLabels(values), startDate, endDate); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
		VALUES(md5($1||'-requirement:'||$2::text)::uuid,$2::uuid,md5($1||'-condition:'||$2::text)::uuid,
			$3::date::timestamp AT TIME ZONE 'Asia/Taipei',($4::date+1)::timestamp AT TIME ZONE 'Asia/Taipei',$5)
		ON CONFLICT(id) DO UPDATE SET effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to,display_order=EXCLUDED.display_order`,
		key, componentID, startDate, endDate, order)
	return err
}

func syncPlanRequirements(ctx context.Context, tx pgx.Tx, item domain.Activity, benefit domain.ActivityBenefit) error {
	qualifiedNames := []string{benefit.QualifiedType}
	selectableNames := []string{benefit.SelectableType}
	if len(qualifiedNames) > 0 {
		if err := syncQualifiedRequirement(ctx, tx, item, benefit, qualifiedNames); err != nil {
			return err
		}
	}
	if len(selectableNames) > 0 {
		if err := syncSelectableRequirement(ctx, tx, item, benefit, selectableNames[0]); err != nil {
			return err
		}
	}
	return nil
}

func splitTypeSet(ctx context.Context, tx pgx.Tx, cardProductID string) (map[string]bool, map[string]bool) {
	var qualified, selectable string
	_ = tx.QueryRow(ctx, `SELECT COALESCE(qualified_type,''),COALESCE(selectable_type,'') FROM card_products WHERE id=$1`, cardProductID).Scan(&qualified, &selectable)
	return labelSet(qualified), labelSet(selectable)
}

func labelSet(value string) map[string]bool {
	out := map[string]bool{}
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			out[item] = true
		}
	}
	return out
}

func syncQualifiedRequirement(ctx context.Context, tx pgx.Tx, item domain.Activity, benefit domain.ActivityBenefit, names []string) error {
	planIDs := []string{}
	for order, name := range names {
		if _, err := tx.Exec(ctx, `INSERT INTO catalog.card_plans(id,card_product_id,plan_type,is_active,display_order)
			VALUES(md5('qualified-plan:'||$1::text||':'||$2)::uuid,$1::uuid,'qualified',true,$3)
			ON CONFLICT(id) DO UPDATE SET is_active=true,display_order=EXCLUDED.display_order`,
			item.CardProductID, name, order+1); err != nil {
			return err
		}
		var planID string
		if err := tx.QueryRow(ctx, `SELECT id::text FROM catalog.card_plans WHERE id=md5('qualified-plan:'||$1::text||':'||$2)::uuid`, item.CardProductID, name).Scan(&planID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO catalog.card_plan_versions(id,card_plan_id,name,description,effective_from,published_at)
			VALUES(md5('qualified-plan-version:'||$1::text)::uuid,$1::uuid,$2,'由會員自行確認是否符合銀行資格','2000-01-01 00:00:00+00',now())
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description`, planID, name); err != nil {
			return err
		}
		planIDs = append(planIDs, planID)
	}
	return syncPlanCondition(ctx, tx, benefit.ID, "qualified", "會員資格", planIDs, "需符合："+joinLabels(names), item.StartDate, item.EndDate, 30)
}

func syncSelectableRequirement(ctx context.Context, tx pgx.Tx, item domain.Activity, benefit domain.ActivityBenefit, name string) error {
	if _, err := tx.Exec(ctx, `INSERT INTO catalog.card_plans(id,card_product_id,plan_type,is_active,display_order)
		VALUES(md5('selectable-plan:'||$1::text)::uuid,$2::uuid,'selectable',true,$3)
		ON CONFLICT(id) DO UPDATE SET card_product_id=EXCLUDED.card_product_id,is_active=true,display_order=EXCLUDED.display_order`,
		benefit.ID, item.CardProductID, benefit.Priority); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO catalog.card_plan_versions(id,card_plan_id,name,description,reminder_text,effective_from,effective_to,published_at)
		VALUES(md5('selectable-plan-version:'||$1::text)::uuid,md5('selectable-plan:'||$1::text)::uuid,$2,$3,$3,
			$4::date::timestamp AT TIME ZONE 'Asia/Taipei',($5::date+1)::timestamp AT TIME ZONE 'Asia/Taipei',now())
		ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description,reminder_text=EXCLUDED.reminder_text,
			effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to`,
		benefit.ID, name, benefit.ActionMessage, item.StartDate, item.EndDate); err != nil {
		return err
	}
	planID := "md5"
	if err := tx.QueryRow(ctx, `SELECT (md5('selectable-plan:'||$1::text)::uuid)::text`, benefit.ID).Scan(&planID); err != nil {
		return err
	}
	return syncPlanCondition(ctx, tx, benefit.ID, "selectable", "切換方案", []string{planID}, benefit.ActionMessage, item.StartDate, item.EndDate, 40)
}

func syncPlanCondition(ctx context.Context, tx pgx.Tx, componentID, key, suffix string, planIDs []string, description, startDate, endDate string, order int) error {
	configuration, err := json.Marshal(map[string]any{"card_plan_ids": planIDs})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward.conditions(id,condition_type,name,is_active)
		VALUES(md5($1||'-condition:'||$2::text)::uuid,'card_plan',$3,true)
		ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,is_active=true`, key, componentID, suffix); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
		VALUES(md5($1||'-condition-version:'||$2::text)::uuid,md5($1||'-condition:'||$2::text)::uuid,'in',$3::jsonb,$4,
			$5::date::timestamp AT TIME ZONE 'Asia/Taipei',($6::date+1)::timestamp AT TIME ZONE 'Asia/Taipei')
		ON CONFLICT(id) DO UPDATE SET configuration_json=EXCLUDED.configuration_json,description=EXCLUDED.description,
			effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to`, key, componentID, configuration, description, startDate, endDate); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
		VALUES(md5($1||'-requirement:'||$2::text)::uuid,$2::uuid,md5($1||'-condition:'||$2::text)::uuid,
			$3::date::timestamp AT TIME ZONE 'Asia/Taipei',($4::date+1)::timestamp AT TIME ZONE 'Asia/Taipei',$5)
		ON CONFLICT(id) DO UPDATE SET effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to,display_order=EXCLUDED.display_order`,
		key, componentID, startDate, endDate, order)
	return err
}

func syncReminderCondition(ctx context.Context, tx pgx.Tx, item domain.Activity, benefit domain.ActivityBenefit) error {
	message := strings.TrimSpace(benefit.ActionMessage)
	configuration, err := json.Marshal(map[string]any{
		"reminder_messages": []string{message},
		"action_required":   benefit.ActionRequired,
	})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward.conditions(id,condition_type,name,is_active)
		VALUES(md5('reminder-condition:'||$1::text)::uuid,'channel',$2,true)
		ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,is_active=true`, benefit.ID, benefit.Name+"／操作提醒"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
		VALUES(md5('reminder-condition-version:'||$1::text)::uuid,md5('reminder-condition:'||$1::text)::uuid,'equals',$2::jsonb,$3,
			$4::date::timestamp AT TIME ZONE 'Asia/Taipei',($5::date+1)::timestamp AT TIME ZONE 'Asia/Taipei')
		ON CONFLICT(id) DO UPDATE SET configuration_json=EXCLUDED.configuration_json,description=EXCLUDED.description,
			effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to`,
		benefit.ID, configuration, message, item.StartDate, item.EndDate); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
		VALUES(md5('reminder-requirement:'||$1::text)::uuid,$1::uuid,md5('reminder-condition:'||$1::text)::uuid,
			$2::date::timestamp AT TIME ZONE 'Asia/Taipei',($3::date+1)::timestamp AT TIME ZONE 'Asia/Taipei',50)
		ON CONFLICT(id) DO UPDATE SET effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to`,
		benefit.ID, item.StartDate, item.EndDate)
	return err
}

func joinLabels(values []string) string {
	out := ""
	for i, value := range values {
		if i > 0 {
			out += " / "
		}
		out += value
	}
	return out
}
