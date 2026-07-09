package member

import (
	"context"
	"encoding/json"
	"fmt"
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
	return domain.MemberCardRewardOverview{}, nil
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
		componentRows, err := r.pool.Query(ctx, `SELECT reward_component_id::text,NULL::text,reward_unit_id,
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
