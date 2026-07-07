package member

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
)

func (r *Repository) UpdatePaymentMethods(ctx context.Context, userID string, codes []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE member_profile.user_payment_methods SET is_available=false,updated_at=now() WHERE user_id=$1`, userID); err != nil {
		return err
	}
	for _, code := range codes {
		tag, err := tx.Exec(ctx, `INSERT INTO member_profile.user_payment_methods(id,user_id,payment_method_id,is_available)
			SELECT md5('user-payment:'||$1::uuid||':'||p.id)::uuid,$1,p.id,true
			FROM payment_methods p WHERE p.id=$2 AND p.is_active AND p.type IN ('mobile_payment','electronic_ticket')
			ON CONFLICT(user_id,payment_method_id) DO UPDATE SET is_available=true,updated_at=now()`, userID, code)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrInvalidInput
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) SetQualificationStatus(ctx context.Context, userID, memberCardID, planID string, qualified bool, effectiveFrom time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM member_cards mc
		JOIN card_products cp ON cp.id=mc.card_product_id
		JOIN catalog.card_plans p ON p.card_product_id=mc.card_product_id
		JOIN LATERAL (
			SELECT * FROM catalog.card_plan_versions v WHERE v.card_plan_id=p.id
			ORDER BY v.effective_from DESC LIMIT 1
		) pv ON true
		WHERE mc.id=$1 AND mc.user_id=$2 AND p.id=$3 AND p.plan_type='qualified'
		AND pv.name=ANY(string_to_array(COALESCE(cp.qualified_type,''), ','))
	)`, memberCardID, userID, planID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM catalog.member_card_qualification_statuses
		WHERE member_card_id=$1 AND card_plan_id=$2 AND effective_from >= $3`, memberCardID, planID, effectiveFrom); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE catalog.member_card_qualification_statuses
		SET effective_to=$3 WHERE member_card_id=$1 AND card_plan_id=$2
		AND effective_from < $3 AND (effective_to IS NULL OR effective_to > $3)`, memberCardID, planID, effectiveFrom); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO catalog.member_card_qualification_statuses
		(id,member_card_id,card_plan_id,is_qualified,effective_from,updated_by_user_at)
		VALUES ($1,$2,$3,$4,$5,now())`, newUUID(), memberCardID, planID, qualified, effectiveFrom); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) RewardOverview(ctx context.Context, userID, memberCardID string, at time.Time) (domain.MemberCardRewardOverview, error) {

	// // 取得指定卡片資訊
	// card, err := r.getMemberCardForOverview(ctx, userID, memberCardID)
	// if err != nil {
	// 	return domain.MemberCardRewardOverview{}, err
	// }

	// // 取得卡片當下的資格方案類型
	// qualifiedTypes := domain.SplitCardType(card.QualifiedType)

	// // 取得指定資格的方案
	// qualifiedPlans, err := r.listQualifiedPlans(ctx, memberCardID, card.CardProductID, qualifiedTypes, at)
	// if err != nil {
	// 	return domain.MemberCardRewardOverview{}, err
	// }

	// // 取得卡片回饋內容
	// components, err := r.listRewardComponents(ctx, card.CardProductID, memberCardID, qualifiedTypes, at)
	// if err != nil {
	// 	return domain.MemberCardRewardOverview{}, err
	// }

	// componentIDs := make([]string, 0, len(components))
	// for _, c := range components {
	// 	componentIDs = append(componentIDs, c.ComponentID)
	// }

	// // 取得回饋條件與提醒文字
	// requirements, reminders, err := r.listRewardRequirementTexts(ctx, componentIDs, at)
	// if err != nil {
	// 	return domain.MemberCardRewardOverview{}, err
	// }

	// // 取得回饋上限資訊
	// caps, err := r.listRewardCaps(ctx, componentIDs, at)
	// if err != nil {
	// 	return domain.MemberCardRewardOverview{}, err
	// }

	// // 組合回饋資訊
	// for i := range components {
	// 	id := components[i].ComponentID

	// 	components[i].Layer = components[i].Layer
	// 	components[i].Requirements = requirements[id]
	// 	components[i].Reminders = reminders[id]

	// 	if components[i].Requirements == nil {
	// 		components[i].Requirements = []string{}
	// 	}
	// 	if components[i].Reminders == nil {
	// 		components[i].Reminders = []string{}
	// 	}

	// 	if cap, ok := caps[id]; ok {
	// 		components[i].Cap = &cap
	// 	}
	// }

	// return domain.MemberCardRewardOverview{
	// 	Card:           card,
	// 	QualifiedPlans: qualifiedPlans,
	// 	RewardGroups:   components,
	// }, nil
	return domain.MemberCardRewardOverview{}, nil
}

func (r *Repository) getMemberCardForOverview(
	ctx context.Context,
	userID string,
	memberCardID string,
) (domain.MemberCard, error) {
	// card, err := r.GetCard(ctx, userID, memberCardID)
	// if err != nil {
	// 	return domain.MemberCard{}, err
	// }
	// if card.NetworkID != "" {
	// 	card.Network = &domain.CardNetwork{
	// 		ID:   card.NetworkID,
	// 		Code: card.NetworkCode,
	// 		Name: card.NetworkName,
	// 	}
	// }

	return domain.MemberCard{}, nil
}
func (r *Repository) listRewardComponents(
	ctx context.Context,
	cardProductID string,
	memberCardID string,
	qualifiedTypes []string,
	at time.Time,
) ([]domain.RewardGroupOverview, error) {
	// 取得卡片回饋內容
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id,
			cv.name,
			cv.reward_value::text,
			c.layer,
			c.display_order,
			cv.effect_type,
			cv.reward_value::text,
			prev.reward_value::text,
			nextv.reward_value::text,
			(
				CASE
					WHEN nextv.id IS NOT NULL THEN nextv.effective_from
					WHEN prev.id IS NOT NULL THEN cv.effective_from
				END
			)::text,
			(
				prev.id IS NOT NULL
				AND cv.display_change_until IS NOT NULL
				AND cv.display_change_until >= $2
			)
		FROM reward.programs p
		JOIN reward.components c
			ON c.reward_program_id = p.id
			AND c.is_active
		JOIN LATERAL (
			SELECT *
			FROM reward.component_versions v
			WHERE v.reward_component_id = c.id
			  AND v.effective_from <= $2
			  AND (v.effective_to IS NULL OR v.effective_to > $2)
			ORDER BY v.effective_from DESC
			LIMIT 1
		) cv ON true
		LEFT JOIN reward.component_versions prev
			ON prev.id = cv.supersedes_version_id
		LEFT JOIN LATERAL (
			SELECT *
			FROM reward.component_versions v
			WHERE v.reward_component_id = c.id
			  AND v.effective_from > $2
			ORDER BY v.effective_from
			LIMIT 1
		) nextv ON true
		LEFT JOIN LATERAL (
			SELECT
				count(*) > 0
				AND count(*) = count(*) FILTER (
					WHERE plan_version.name = ANY($4::text[])
				) AS has_qualified_requirement,
				bool_or(
					plan_version.name = ANY($4::text[])
					AND COALESCE(status.is_qualified, false)
				) AS matches_qualified_requirement
			FROM reward.requirements requirement
			JOIN reward.condition_versions condition_version
				ON condition_version.reward_condition_id = requirement.reward_condition_id
				AND condition_version.effective_from <= $2
				AND (condition_version.effective_to IS NULL OR condition_version.effective_to > $2)
			CROSS JOIN LATERAL jsonb_array_elements_text(
				condition_version.configuration_json->'card_plan_ids'
			) plan_id
			JOIN catalog.card_plans plan
				ON plan.id = plan_id.value::uuid
				AND plan.plan_type = 'qualified'
			JOIN LATERAL (
				SELECT v.name
				FROM catalog.card_plan_versions v
				WHERE v.card_plan_id = plan.id
				  AND v.effective_from <= $2
				  AND (v.effective_to IS NULL OR v.effective_to > $2)
				ORDER BY v.effective_from DESC
				LIMIT 1
			) plan_version ON true
			LEFT JOIN LATERAL (
				SELECT q.is_qualified
				FROM catalog.member_card_qualification_statuses q
				WHERE q.member_card_id = $3
				  AND q.card_plan_id = plan.id
				  AND q.effective_from <= $2
				  AND (q.effective_to IS NULL OR q.effective_to > $2)
				ORDER BY q.effective_from DESC
				LIMIT 1
			) status ON true
			WHERE requirement.reward_component_id = c.id
			  AND requirement.effective_from <= $2
			  AND (requirement.effective_to IS NULL OR requirement.effective_to > $2)
		) qualified_gate ON true
		WHERE p.card_product_id = $1
		  AND p.status = 'published'
		  AND (
			NOT COALESCE(qualified_gate.has_qualified_requirement, false)
			OR COALESCE(qualified_gate.matches_qualified_requirement, false)
		  )
		ORDER BY c.layer, c.display_order, c.priority, cv.name
	`, cardProductID, at, memberCardID, qualifiedTypes)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []domain.RewardGroupOverview{}

	for rows.Next() {
		item, err := scanRewardGroupOverview(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *Repository) listQualifiedPlans(
	ctx context.Context,
	memberCardID string,
	cardProductID string,
	qualifiedTypes []string,
	at time.Time,
) ([]domain.QualificationStatus, error) {

	result := []domain.QualificationStatus{}

	if len(qualifiedTypes) == 0 {
		return result, nil
	}
	// 取得指定資格的方案
	rows, err := r.pool.Query(ctx, `
		SELECT
			p.id,
			pv.name,
			COALESCE(s.is_qualified, false),
			COALESCE(s.effective_from, pv.effective_from)
		FROM catalog.card_plans p
		JOIN LATERAL (
			SELECT *
			FROM catalog.card_plan_versions v
			WHERE v.card_plan_id = p.id
			  AND v.effective_from <= $3
			  AND (v.effective_to IS NULL OR v.effective_to > $3)
			ORDER BY v.effective_from DESC
			LIMIT 1
		) pv ON true
		LEFT JOIN LATERAL (
			SELECT *
			FROM catalog.member_card_qualification_statuses q
			WHERE q.member_card_id = $1
			  AND q.card_plan_id = p.id
			  AND q.effective_from <= $3
			  AND (q.effective_to IS NULL OR q.effective_to > $3)
			ORDER BY q.effective_from DESC
			LIMIT 1
		) s ON true
		WHERE p.card_product_id = $2
		  AND p.plan_type = 'qualified'
		  AND p.is_active
		  AND pv.name = ANY($4::text[])
		ORDER BY p.display_order, pv.name
	`, memberCardID, cardProductID, at, qualifiedTypes)

	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.QualificationStatus
		var effectiveFrom time.Time

		if err := rows.Scan(
			&item.PlanID,
			&item.Name,
			&item.IsQualified,
			&effectiveFrom,
		); err != nil {
			return result, err
		}

		item.EffectiveFrom = effectiveFrom.Format(time.RFC3339)
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

func scanRewardGroupOverview(rows pgx.Rows) (domain.RewardGroupOverview, error) {
	var item domain.RewardGroupOverview
	var previousRate *string
	var nextRate *string
	var changeEffectiveAt *string

	err := rows.Scan(
		&item.ComponentID,
		&item.Name,
		&item.CurrentRate,
		&item.Layer,
		&item.DisplayOrder,
		&item.EffectType,
		&item.RewardValue,
		&previousRate,
		&nextRate,
		&changeEffectiveAt,
		&item.ShowPreviousAsStrikethrough,
	)

	if err != nil {
		return domain.RewardGroupOverview{}, err
	}

	item.PreviousRate = previousRate
	item.NextRate = nextRate
	item.ChangeEffectiveAt = changeEffectiveAt
	item.Requirements = []string{}
	item.Reminders = []string{}

	return item, nil
}
func (r *Repository) listRewardRequirementTexts(
	ctx context.Context,
	componentIDs []string,
	at time.Time,
) (
	map[string][]string,
	map[string][]string,
	error,
) {
	requirements := map[string][]string{}
	reminders := map[string][]string{}

	if len(componentIDs) == 0 {
		return requirements, reminders, nil
	}
	// 取得回饋條件與提醒文字
	rows, err := r.pool.Query(ctx, `
		SELECT
			rr.reward_component_id,
			condition.condition_type,
			x.description,
			category_text.value,
			payment_text.value,
			merchant_text.value,
			cp.id,
			cp.plan_type,
			pv.name,
			pv.reminder_text,
			component_version.name,
			component_version.reward_value::text,
			reminder_text.value
		FROM reward.requirements rr
		JOIN reward.conditions condition
			ON condition.id = rr.reward_condition_id
		JOIN reward.condition_versions x
			ON x.reward_condition_id = rr.reward_condition_id
			AND x.effective_from <= $2
			AND (x.effective_to IS NULL OR x.effective_to > $2)
		LEFT JOIN LATERAL jsonb_array_elements_text(
			x.configuration_json->'card_plan_ids'
		) plan_id ON true
		LEFT JOIN catalog.card_plans cp
			ON cp.id = plan_id.value::uuid
		LEFT JOIN LATERAL (
			SELECT *
			FROM catalog.card_plan_versions v
			WHERE v.card_plan_id = cp.id
			  AND v.effective_from <= $2
			  AND (v.effective_to IS NULL OR v.effective_to > $2)
			ORDER BY v.effective_from DESC
			LIMIT 1
		) pv ON true
		LEFT JOIN LATERAL (
			SELECT name, reward_value
			FROM reward.component_versions v
			WHERE v.reward_component_id = rr.reward_component_id
			  AND v.effective_from <= $2
			  AND (v.effective_to IS NULL OR v.effective_to > $2)
			ORDER BY v.effective_from DESC
			LIMIT 1
		) component_version ON true
		LEFT JOIN LATERAL (
			SELECT string_agg(category.name, '、' ORDER BY category.name) AS value
			FROM jsonb_array_elements_text(x.configuration_json->'category_ids') category_id
			JOIN categories category
				ON category.id = category_id.value
		) category_text ON condition.condition_type = 'category'
		LEFT JOIN LATERAL (
			SELECT string_agg(method.name, '、' ORDER BY method.name) AS value
			FROM jsonb_array_elements_text(x.configuration_json->'payment_method_ids') method_code
			JOIN payment_methods method
				ON method.id = method_code.value
		) payment_text ON condition.condition_type = 'payment_method'
		LEFT JOIN LATERAL (
			SELECT string_agg(merchant.name, '、' ORDER BY merchant.name) AS value
			FROM jsonb_array_elements_text(x.configuration_json->'merchant_ids') merchant_id
			JOIN merchants merchant
				ON merchant.id = merchant_id.value::bigint
		) merchant_text ON condition.condition_type = 'merchant'
		LEFT JOIN LATERAL jsonb_array_elements_text(
			x.configuration_json->'reminder_messages'
		) reminder_text(value)
			ON condition.condition_type = 'channel'
		WHERE rr.reward_component_id = ANY($1::uuid[])
		  AND rr.effective_from <= $2
		  AND (rr.effective_to IS NULL OR rr.effective_to > $2)
	`, componentIDs, at)

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var componentID string
		var conditionType string
		var description *string
		var categoryText *string
		var paymentText *string
		var merchantText *string
		var cardPlanID *string
		var planType *string
		var planName *string
		var planReminder *string
		var benefitName *string
		var rewardValue *string
		var channelReminder *string

		if err := rows.Scan(
			&componentID,
			&conditionType,
			&description,
			&categoryText,
			&paymentText,
			&merchantText,
			&cardPlanID,
			&planType,
			&planName,
			&planReminder,
			&benefitName,
			&rewardValue,
			&channelReminder,
		); err != nil {
			return nil, nil, err
		}

		if channelReminder != nil && *channelReminder != "" {
			appendUnique(&reminders, componentID, *channelReminder)
		}

		if planType != nil &&
			*planType == "selectable" &&
			planName != nil &&
			benefitName != nil &&
			rewardValue != nil {
			appendUnique(&reminders, componentID, selectablePlanReminder(*planName, *benefitName, *rewardValue, stringValue(planReminder)))
		}

		if cardPlanID != nil || channelReminder != nil {
			continue
		}

		text := firstNonEmpty(categoryText, paymentText, merchantText, description)
		if text != "" {
			appendUnique(&requirements, componentID, text)
		}
	}

	return requirements, reminders, rows.Err()
}

func (r *Repository) listRewardCaps(
	ctx context.Context,
	componentIDs []string,
	at time.Time,
) (map[string]domain.RewardCapOverview, error) {

	result := map[string]domain.RewardCapOverview{}

	if len(componentIDs) == 0 {
		return result, nil
	}
	// 取得回饋上限資訊
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id,
			cap.cap_type,
			COALESCE(cap.limit_value::text, cap.limit_formula),
			cap.period_type,
			CASE
				WHEN cap.cap_type = 'reward_amount'
				     AND component_version.reward_value > 0
				     AND cap.limit_value IS NOT NULL
				THEN (cap.limit_value / component_version.reward_value)::text
				ELSE NULL
			END
		FROM reward.components c
		JOIN LATERAL (
			SELECT v.reward_value
			FROM reward.component_versions v
			WHERE v.reward_component_id = c.id
			  AND v.effective_from <= $2
			  AND (v.effective_to IS NULL OR v.effective_to > $2)
			ORDER BY v.effective_from DESC
			LIMIT 1
		) component_version ON true
		JOIN LATERAL (
			SELECT
				cv.cap_type,
				cv.limit_value,
				cv.limit_formula,
				cv.period_type
			FROM reward.component_caps cc
			JOIN reward.cap_versions cv
				ON cv.reward_cap_id = cc.reward_cap_id
			WHERE cc.reward_component_id = c.id
			  AND cc.effective_from <= $2
			  AND (cc.effective_to IS NULL OR cc.effective_to > $2)
			  AND cv.effective_from <= $2
			  AND (cv.effective_to IS NULL OR cv.effective_to > $2)
			ORDER BY cv.effective_from DESC
			LIMIT 1
		) cap ON true
		WHERE c.id = ANY($1::uuid[])
	`, componentIDs, at)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var componentID string
		var cap domain.RewardCapOverview
		var spendable *string

		if err := rows.Scan(
			&componentID,
			&cap.Type,
			&cap.Limit,
			&cap.Period,
			&spendable,
		); err != nil {
			return nil, err
		}

		if spendable != nil {
			cap.Spendable = *spendable
		}

		result[componentID] = cap
	}

	return result, rows.Err()
}

func (r *Repository) InsertOutboxEvent(ctx context.Context, tx pgx.Tx, aggregateType, aggregateID, eventType string, payload []byte) error {
	if !json.Valid(payload) {
		return fmt.Errorf("invalid outbox payload")
	}
	_, err := tx.Exec(ctx, `INSERT INTO integration.outbox_events(id,aggregate_type,aggregate_id,event_type,payload_json)
		VALUES ($1,$2,$3,$4,$5)`, newUUID(), aggregateType, aggregateID, eventType, payload)
	return err
}

func (r *Repository) RewardCalculations(ctx context.Context, userID, transactionID string) ([]domain.RewardCalculationSnapshot, error) {
	rows, err := r.pool.Query(ctx, `SELECT c.id,c.calculation_no,c.calculation_type,c.transaction_at::text,
		c.total_effective_rate::text,c.total_reward_value::text,c.engine_version,c.status,
		c.input_snapshot_json,c.calculated_at::text
		FROM "transaction".reward_calculations c JOIN transactions t ON t.id=c.transaction_id
		WHERE c.transaction_id=$1 AND t.user_id=$2 ORDER BY c.calculation_no DESC`, transactionID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RewardCalculationSnapshot{}
	for rows.Next() {
		var item domain.RewardCalculationSnapshot
		if err := rows.Scan(&item.ID, &item.CalculationNo, &item.CalculationType, &item.TransactionAt,
			&item.TotalEffectiveRate, &item.TotalRewardValue, &item.EngineVersion, &item.Status,
			&item.InputSnapshot, &item.CalculatedAt); err != nil {
			return nil, err
		}
		componentRows, err := r.pool.Query(ctx, `SELECT reward_component_id::text,reward_component_version_id::text,reward_unit_id,
			suggested_card_plan_id::text,qualified_card_plan_id::text,member_qualification_status_id::text,
			effect_type,reward_value::text,reward_rate::text,uncapped_reward::text,allocated_reward::text,cap_used_before::text,
			cap_remaining_before::text,condition_snapshot_json,reminder_snapshot_json
			FROM "transaction".reward_calculation_components WHERE reward_calculation_id=$1 ORDER BY id`, item.ID)
		if err != nil {
			return nil, err
		}
		item.Components = []domain.RewardCalculationComponentSnapshot{}
		for componentRows.Next() {
			var component domain.RewardCalculationComponentSnapshot
			if err := componentRows.Scan(&component.ComponentID, &component.ComponentVersionID, &component.RewardUnitID,
				&component.SuggestedCardPlanID, &component.QualifiedCardPlanID, &component.QualificationStatusID,
				&component.EffectType, &component.RewardValue, &component.RewardRate, &component.UncappedReward, &component.AllocatedReward, &component.CapUsedBefore,
				&component.CapRemainingBefore, &component.ConditionSnapshot, &component.ReminderSnapshot); err != nil {
				componentRows.Close()
				return nil, err
			}
			item.Components = append(item.Components, component)
		}
		componentRows.Close()
		out = append(out, item)
	}
	if len(out) == 0 {
		var exists bool
		if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM transactions WHERE id=$1 AND user_id=$2)`, transactionID, userID).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrNotFound
		}
	}
	return out, rows.Err()
}

func appendUnique(target *map[string][]string, key string, value string) {
	if value == "" {
		return
	}
	for _, existing := range (*target)[key] {
		if existing == value {
			return
		}
	}
	(*target)[key] = append((*target)[key], value)
}

func firstNonEmpty(values ...*string) string {
	for _, value := range values {
		if value != nil && *value != "" {
			return *value
		}
	}
	return ""
}

func selectablePlanReminder(planName, benefitName, rewardValue, reminderText string) string {
	base := fmt.Sprintf(
		"需至 APP 切換卡片方案：%s，對應優惠：%s %s",
		planName,
		benefitName,
		trimNumericText(rewardValue),
	)
	reminderText = strings.TrimSpace(reminderText)
	if reminderText == "" || reminderText == base {
		return base
	}
	return base + "；" + reminderText
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func trimNumericText(value string) string {
	value = strings.TrimRight(strings.TrimRight(value, "0"), ".")
	if value == "" {
		return "0"
	}
	return value
}
