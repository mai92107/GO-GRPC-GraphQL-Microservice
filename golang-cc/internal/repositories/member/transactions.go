package member

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (s *TransactionRepository) List(ctx context.Context, userID, id string) ([]domain.TransactionSummary, error) {
	rows, err := s.pool.Query(ctx, `SELECT t.id,t.card_id,t.amount_minor,t.category_id,COALESCE(t.merchant_id::text,''),t.merchant_name,t.payment_method_id,p.name,t.transaction_date::text,COALESCE(t.note,'')
		FROM transactions t JOIN payment_methods p ON p.id=t.payment_method_id
		WHERE t.user_id=$1 AND ($2='' OR t.id::text=$2) ORDER BY t.transaction_date DESC,t.created_at DESC`, userID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.TransactionSummary{}
	for rows.Next() {
		var item domain.TransactionSummary
		if err := rows.Scan(&item.ID, &item.CardID, &item.AmountMinor, &item.CategoryID, &item.MerchantID, &item.MerchantName, &item.PaymentMethodID, &item.PaymentMethodName, &item.TransactionDate, &item.Note); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (s *TransactionRepository) Recommend(ctx context.Context, userID recommendations.ID, amountMinor int64, category, merchantID, merchantName string, date recommendations.LocalDate) (recommendations.Result, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return recommendations.Result{}, err
	}
	defer tx.Rollback(ctx)
	if merchantID != "" {
		if _, err := loadMerchantName(ctx, tx, merchantID); err != nil {
			return recommendations.Result{}, err
		}
	}
	rows, err := tx.Query(ctx, `SELECT mc.id, mc.user_id,
		CASE WHEN NULLIF(mc.nickname,'') IS NULL THEN cp.name ELSE mc.nickname || '（' || cp.name || '）' END,
		COALESCE(mc.account_tier,''),COALESCE(mc.card_network_id::text,''),COALESCE(limit_at_date.credit_limit, mc.credit_limit::text),mc.is_active,
		ARRAY(SELECT q.card_plan_id FROM catalog.member_card_qualification_statuses q
			WHERE q.member_card_id=mc.id AND q.is_qualified
			AND q.effective_from <= $2 AND (q.effective_to IS NULL OR q.effective_to > $2))
		FROM member_cards mc JOIN card_products cp ON cp.id=mc.card_product_id
		LEFT JOIN LATERAL (
			SELECT h.credit_limit::text AS credit_limit
			FROM member_card_credit_limits h
			WHERE h.member_card_id=mc.id
			AND (h.effective_from AT TIME ZONE 'Asia/Taipei')::date <= $2::date
			AND (h.effective_to IS NULL OR (h.effective_to AT TIME ZONE 'Asia/Taipei')::date > $2::date)
			ORDER BY h.effective_from DESC, h.created_at DESC
			LIMIT 1
		) limit_at_date ON true
		WHERE mc.user_id = $1 AND mc.is_active AND cp.is_active`, userID, date.Time)
	if err != nil {
		return recommendations.Result{}, err
	}
	var cards []recommendations.Card
	for rows.Next() {
		var card recommendations.Card
		var creditLimit *string
		if err := rows.Scan(&card.ID, &card.UserID, &card.Name, &card.AccountTier, &card.NetworkID, &creditLimit, &card.IsActive, &card.QualifiedCardPlanIDs); err != nil {
			rows.Close()
			return recommendations.Result{}, err
		}
		if creditLimit != nil {
			value := recommendations.MustDecimal(*creditLimit)
			card.CreditLimit = &value
		}
		cards = append(cards, card)
	}
	rows.Close()
	var rules []recommendations.RewardRule
	for _, card := range cards {
		cardRules, err := loadRules(ctx, tx, userID, card.ID, date)
		if err != nil {
			return recommendations.Result{}, err
		}
		rules = append(rules, cardRules...)
	}
	preferences, err := loadPreferences(ctx, tx, userID)
	if err != nil {
		return recommendations.Result{}, err
	}
	usage, err := loadMonthlyUsage(ctx, tx, userID, date)
	if err != nil {
		return recommendations.Result{}, err
	}
	methods, err := loadPaymentMethods(ctx, tx, userID)
	if err != nil {
		return recommendations.Result{}, err
	}
	return recommendations.Recommend(recommendations.RecommendationInput{
		UserID: userID, AmountMinor: amountMinor, CategoryID: category, MerchantID: merchantID, MerchantName: merchantName, PaymentMethods: methods, Date: date,
		Cards: cards, Rules: rules, Preferences: preferences, MonthlyUsage: usage, ActivityMonthlyUsage: loadActivityMonthlyUsage(ctx, tx, userID, date),
	})
}

func (s *TransactionRepository) Create(ctx context.Context, userID recommendations.ID, input WriteInput) (Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Transaction{}, fmt.Errorf("begin create transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if input.MerchantID != "" {
		input.MerchantName, err = loadMerchantName(ctx, tx, input.MerchantID)
		if err != nil {
			return Transaction{}, err
		}
	}

	if err := lockKeys(ctx, tx, []string{lockKey(userID, input.CardID, input.TransactionDate)}); err != nil {
		return Transaction{}, err
	}
	result, err := buildRecommendation(ctx, tx, userID, input)
	if err != nil {
		return Transaction{}, err
	}
	paymentMethodName, err := loadPaymentMethodName(ctx, tx, input.PaymentMethodID)
	if err != nil {
		return Transaction{}, err
	}
	transactionID := newUUID()
	if _, err := tx.Exec(ctx, `INSERT INTO transactions
		(id, user_id, card_id, card_network_id, amount_minor, category_id, merchant_id, merchant_name, payment_method_id, transaction_date, note)
		VALUES ($1, $2, $3, (SELECT card_network_id FROM member_cards WHERE id=$3), $4, $5, NULLIF($6,'')::uuid, $7, $8, $9, NULLIF($10, ''))`,
		transactionID, userID, input.CardID, input.AmountMinor, input.CategoryID,
		input.MerchantID, input.MerchantName, input.PaymentMethodID, input.TransactionDate.Time, input.Note); err != nil {
		return Transaction{}, fmt.Errorf("insert transaction: %w", err)
	}
	if err := insertAllocations(ctx, tx, transactionID, result.Allocations); err != nil {
		return Transaction{}, err
	}
	if err := insertRewardCalculation(ctx, tx, transactionID, input, result.Allocations, "original"); err != nil {
		return Transaction{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, fmt.Errorf("commit create transaction: %w", err)
	}
	return toTransaction(transactionID, userID, input, paymentMethodName, result.Allocations), nil
}

func (s *TransactionRepository) Update(ctx context.Context, userID, transactionID recommendations.ID, input WriteInput) (Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Transaction{}, fmt.Errorf("begin update transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if input.MerchantID != "" {
		input.MerchantName, err = loadMerchantName(ctx, tx, input.MerchantID)
		if err != nil {
			return Transaction{}, err
		}
	}

	var oldCardID recommendations.ID
	var oldDate time.Time
	if err := tx.QueryRow(ctx, `SELECT card_id, transaction_date FROM transactions
		WHERE id = $1 AND user_id = $2 FOR UPDATE`, transactionID, userID).Scan(&oldCardID, &oldDate); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, ErrNotFound
		}
		return Transaction{}, fmt.Errorf("load transaction for update: %w", err)
	}
	keys := []string{
		lockKey(userID, oldCardID, recommendations.LocalDate{Time: oldDate}),
		lockKey(userID, input.CardID, input.TransactionDate),
	}
	if err := lockKeys(ctx, tx, keys); err != nil {
		return Transaction{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM reward_allocations WHERE transaction_id = $1`, transactionID); err != nil {
		return Transaction{}, fmt.Errorf("delete old allocations: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE transactions SET
		card_id = $3, amount_minor = $4, category_id = $5, merchant_id=NULLIF($6,'')::uuid, merchant_name = $7, payment_method_id = $8, transaction_date = $9,
		note = NULLIF($10, ''), updated_at = now()
		WHERE id = $1 AND user_id = $2`,
		transactionID, userID, input.CardID, input.AmountMinor, input.CategoryID,
		input.MerchantID, input.MerchantName, input.PaymentMethodID, input.TransactionDate.Time, input.Note); err != nil {
		return Transaction{}, fmt.Errorf("update transaction: %w", err)
	}
	result, err := buildRecommendation(ctx, tx, userID, input)
	if err != nil {
		return Transaction{}, err
	}
	paymentMethodName, err := loadPaymentMethodName(ctx, tx, input.PaymentMethodID)
	if err != nil {
		return Transaction{}, err
	}
	if err := insertAllocations(ctx, tx, transactionID, result.Allocations); err != nil {
		return Transaction{}, err
	}
	if err := insertRewardCalculation(ctx, tx, transactionID, input, result.Allocations, "manual_recalculation"); err != nil {
		return Transaction{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, fmt.Errorf("commit update transaction: %w", err)
	}
	return toTransaction(transactionID, userID, input, paymentMethodName, result.Allocations), nil
}

func (s *TransactionRepository) Delete(ctx context.Context, userID, transactionID recommendations.ID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var cardID recommendations.ID
	var date time.Time
	if err := tx.QueryRow(ctx, `SELECT card_id, transaction_date FROM transactions
		WHERE id = $1 AND user_id = $2 FOR UPDATE`, transactionID, userID).Scan(&cardID, &date); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("load transaction for delete: %w", err)
	}
	if err := lockKeys(ctx, tx, []string{lockKey(userID, cardID, recommendations.LocalDate{Time: date})}); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM transactions WHERE id = $1 AND user_id = $2`, transactionID, userID); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete transaction: %w", err)
	}
	return nil
}

func buildRecommendation(ctx context.Context, tx pgx.Tx, userID recommendations.ID, input WriteInput) (recommendations.CardRecommendation, error) {
	paymentMethodName, err := loadPaymentMethodName(ctx, tx, input.PaymentMethodID)
	if err != nil {
		return recommendations.CardRecommendation{}, err
	}
	cards, err := loadCard(ctx, tx, userID, input.CardID, input.TransactionDate)
	if err != nil {
		return recommendations.CardRecommendation{}, err
	}
	rules, err := loadRules(ctx, tx, userID, input.CardID, input.TransactionDate)
	if err != nil {
		return recommendations.CardRecommendation{}, err
	}
	preferences, err := loadPreferences(ctx, tx, userID)
	if err != nil {
		return recommendations.CardRecommendation{}, err
	}
	usage, err := loadMonthlyUsage(ctx, tx, userID, input.TransactionDate)
	if err != nil {
		return recommendations.CardRecommendation{}, err
	}
	result, err := recommendations.Recommend(recommendations.RecommendationInput{
		UserID: userID, AmountMinor: input.AmountMinor, CategoryID: input.CategoryID, MerchantID: input.MerchantID, MerchantName: input.MerchantName,
		PaymentMethods: []recommendations.PaymentMethod{{ID: input.PaymentMethodID, Name: paymentMethodName}},
		Date:           input.TransactionDate, Cards: cards, Rules: rules,
		Preferences: preferences, MonthlyUsage: usage, ActivityMonthlyUsage: loadActivityMonthlyUsage(ctx, tx, userID, input.TransactionDate),
	})
	if err != nil {
		return recommendations.CardRecommendation{}, err
	}
	if len(result.Recommendations) == 0 {
		return recommendations.CardRecommendation{
			CardID: input.CardID, CardName: cards[0].Name,
			Allocations: []recommendations.RuleEvaluation{},
		}, nil
	}
	return result.Recommendations[0], nil
}

func loadCard(ctx context.Context, tx pgx.Tx, userID, cardID recommendations.ID, date recommendations.LocalDate) ([]recommendations.Card, error) {
	var card recommendations.Card
	var creditLimit *string
	if err := tx.QueryRow(ctx, `SELECT mc.id, mc.user_id,
		CASE WHEN NULLIF(mc.nickname,'') IS NULL THEN cp.name ELSE mc.nickname || '（' || cp.name || '）' END,
		COALESCE(mc.account_tier,''),COALESCE(mc.card_network_id::text,''),COALESCE(limit_at_date.credit_limit, mc.credit_limit::text),mc.is_active,
		ARRAY(SELECT q.card_plan_id FROM catalog.member_card_qualification_statuses q
			WHERE q.member_card_id=mc.id AND q.is_qualified
			AND q.effective_from <= $3 AND (q.effective_to IS NULL OR q.effective_to > $3))
		FROM member_cards mc JOIN card_products cp ON cp.id=mc.card_product_id
		LEFT JOIN LATERAL (
			SELECT h.credit_limit::text AS credit_limit
			FROM member_card_credit_limits h
			WHERE h.member_card_id=mc.id
			AND (h.effective_from AT TIME ZONE 'Asia/Taipei')::date <= $3::date
			AND (h.effective_to IS NULL OR (h.effective_to AT TIME ZONE 'Asia/Taipei')::date > $3::date)
			ORDER BY h.effective_from DESC, h.created_at DESC
			LIMIT 1
		) limit_at_date ON true
		WHERE mc.id = $1 AND mc.user_id = $2 AND mc.is_active AND cp.is_active`, cardID, userID, date.Time).
		Scan(&card.ID, &card.UserID, &card.Name, &card.AccountTier, &card.NetworkID, &creditLimit, &card.IsActive, &card.QualifiedCardPlanIDs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("active card not found")
		}
		return nil, fmt.Errorf("load card: %w", err)
	}
	if creditLimit != nil {
		value := recommendations.MustDecimal(*creditLimit)
		card.CreditLimit = &value
	}
	return []recommendations.Card{card}, nil
}

func loadRules(ctx context.Context, tx pgx.Tx, userID, cardID recommendations.ID, date recommendations.LocalDate) ([]recommendations.RewardRule, error) {
	rows, err := tx.Query(ctx, `SELECT c.id,p.id,p.name,$1::uuid,mc.id,cv.name,
		cap.limit_value::text,
		COALESCE(cap.limit_formula,''),
		COALESCE((cv.effective_from AT TIME ZONE 'Asia/Taipei')::date,DATE '2000-01-01'),
		CASE WHEN cv.effective_to IS NULL THEN NULL ELSE ((cv.effective_to-interval '1 microsecond') AT TIME ZONE 'Asia/Taipei')::date END,
		(p.status='published' AND c.is_active),c.stack_group,c.priority,
		c.layer,c.display_order,cv.effect_type,cv.reward_value::text,
		COALESCE(req.qualified_names,'{}'),
		COALESCE(req.action_required,'none'),COALESCE(req.action_message,''),
		COALESCE(req.payment_methods,'{}'),
		NULL::text,
		u.id, u.id, u.name, u.symbol, u.symbol_position, u.twd_rate::text, u.precision,
		COALESCE(req.card_network_ids,'{}'),COALESCE(req.categories,'{}'),COALESCE(req.merchants,'{}')
		FROM member_cards mc
		JOIN reward.programs p ON p.card_product_id=mc.card_product_id
		JOIN reward.components c ON c.reward_program_id=p.id
		JOIN LATERAL (
			SELECT v.* FROM reward.component_versions v WHERE v.reward_component_id=c.id
			AND v.effective_from <= $3 AND (v.effective_to IS NULL OR v.effective_to > $3)
			ORDER BY v.effective_from DESC LIMIT 1
		) cv ON true
		JOIN reward_units u ON u.id=cv.reward_unit_id
		LEFT JOIN LATERAL (
			SELECT v.limit_value,v.limit_formula FROM reward.component_caps cc
			JOIN reward.cap_versions v ON v.reward_cap_id=cc.reward_cap_id
			WHERE cc.reward_component_id=c.id AND cc.effective_from <= $3
			AND (cc.effective_to IS NULL OR cc.effective_to > $3)
			AND v.effective_from <= $3 AND (v.effective_to IS NULL OR v.effective_to > $3)
			ORDER BY v.effective_from DESC LIMIT 1
		) cap ON true
		LEFT JOIN LATERAL (
			SELECT
				array_remove(array_agg(DISTINCT qpv.name) FILTER (WHERE cp.plan_type='qualified'),NULL) AS qualified_names,
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
				array_remove(array_agg(DISTINCT cn.value::uuid),NULL) AS card_network_ids,
				array_remove(array_agg(DISTINCT cat.value),NULL) AS categories,
				array_remove(array_agg(DISTINCT mer.value),NULL) AS merchants
			FROM reward.requirements rr
			JOIN reward.conditions rc ON rc.id=rr.reward_condition_id AND rc.is_active
			JOIN reward.condition_versions x ON x.reward_condition_id=rr.reward_condition_id
				AND x.effective_from <= $3 AND (x.effective_to IS NULL OR x.effective_to > $3)
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'card_plan_ids') plan_id ON true
			LEFT JOIN catalog.card_plans cp ON cp.id=plan_id.value::uuid
			LEFT JOIN LATERAL (
				SELECT name FROM catalog.card_plan_versions pv WHERE pv.card_plan_id=cp.id
				AND pv.effective_from <= $3 AND (pv.effective_to IS NULL OR pv.effective_to > $3)
				ORDER BY pv.effective_from DESC LIMIT 1
			) qpv ON cp.plan_type='qualified'
			LEFT JOIN LATERAL (
				SELECT name,reminder_text FROM catalog.card_plan_versions pv WHERE pv.card_plan_id=cp.id
				AND pv.effective_from <= $3 AND (pv.effective_to IS NULL OR pv.effective_to > $3)
				ORDER BY pv.effective_from DESC LIMIT 1
			) spv ON cp.plan_type='selectable'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'payment_method_ids') pm(value) ON rc.condition_type='payment_method'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'card_network_ids') cn(value) ON rc.condition_type='card_network'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'category_ids') cat(value) ON rc.condition_type='category'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'merchant_ids') mer(value) ON rc.condition_type='merchant'
			LEFT JOIN LATERAL jsonb_array_elements_text(x.configuration_json->'reminder_messages') rem(value) ON rc.condition_type='channel'
			WHERE rr.reward_component_id=c.id AND rr.effective_from <= $3
			AND (rr.effective_to IS NULL OR rr.effective_to > $3)
		) req ON true
		WHERE mc.user_id = $1 AND mc.id = $2 AND p.status='published' AND c.is_active`, userID, cardID, date.Time)
	if err != nil {
		return nil, fmt.Errorf("query rules: %w", err)
	}
	defer rows.Close()

	var rules []recommendations.RewardRule
	for rows.Next() {
		var rule recommendations.RewardRule
		var cap *string
		var sharedCap *string
		var twdRate string
		var rewardValue string
		var start, end *time.Time
		if err := rows.Scan(&rule.ID, &rule.ActivityID, &rule.ActivityName, &rule.UserID, &rule.CardID, &rule.Name,
			&cap, &rule.MonthlyCapFormula, &start, &end, &rule.IsActive, &rule.StackGroup, &rule.Priority,
			&rule.Layer, &rule.DisplayOrder, &rule.EffectType, &rewardValue,
			&rule.QualifiedType, &rule.ActionRequired, &rule.ActionMessage, &rule.PaymentMethods, &sharedCap, &rule.RewardUnit.ID, &rule.RewardUnit.ID,
			&rule.RewardUnit.Name, &rule.RewardUnit.Symbol, &rule.RewardUnit.SymbolPosition, &twdRate, &rule.RewardUnit.Precision,
			&rule.CardNetworkIDs, &rule.CategoryID, &rule.MerchantIDs); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}
		rule.RewardValue = recommendations.MustDecimal(rewardValue)
		rule.RewardUnit.TWDRate = recommendations.MustDecimal(twdRate)
		if cap != nil {
			value := recommendations.MustDecimal(*cap)
			rule.MonthlyCap = &value
		}
		if sharedCap != nil {
			value := recommendations.MustDecimal(*sharedCap)
			rule.SharedMonthlyCap = &value
		}
		if start != nil {
			value := recommendations.LocalDate{Time: *start}
			rule.StartDate = &value
		}
		if end != nil {
			value := recommendations.LocalDate{Time: *end}
			rule.EndDate = &value
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	for i := range rules {
		if err := loadVersionedRuleConditions(ctx, tx, &rules[i], date); err != nil {
			return nil, err
		}
	}
	return rules, nil
}

func loadVersionedRuleConditions(ctx context.Context, tx pgx.Tx, rule *recommendations.RewardRule, date recommendations.LocalDate) error {
	at := date.Time
	var categories, merchants, paymentMethods []string
	if err := tx.QueryRow(ctx, `SELECT
		COALESCE(ARRAY(
			SELECT value::uuid FROM reward.requirements rr
			JOIN reward.condition_versions cv ON cv.reward_condition_id=rr.reward_condition_id
			CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'card_network_ids') value
			WHERE rr.reward_component_id=$1 AND rr.effective_from <= $2 AND (rr.effective_to IS NULL OR rr.effective_to > $2)
			AND cv.effective_from <= $2 AND (cv.effective_to IS NULL OR cv.effective_to > $2)
		),'{}'),
		COALESCE(ARRAY(
			SELECT value::uuid FROM reward.requirements rr
			JOIN reward.condition_versions cv ON cv.reward_condition_id=rr.reward_condition_id
			CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'card_plan_ids') value
			JOIN catalog.card_plans p ON p.id=value::uuid AND p.plan_type='qualified'
			WHERE rr.reward_component_id=$1 AND rr.effective_from <= $2 AND (rr.effective_to IS NULL OR rr.effective_to > $2)
			AND cv.effective_from <= $2 AND (cv.effective_to IS NULL OR cv.effective_to > $2)
		),'{}'),
		COALESCE((
			SELECT p.id FROM reward.requirements rr
			JOIN reward.condition_versions cv ON cv.reward_condition_id=rr.reward_condition_id
			CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'card_plan_ids') value
			JOIN catalog.card_plans p ON p.id=value::uuid AND p.plan_type='selectable'
			WHERE rr.reward_component_id=$1 AND rr.effective_from <= $2 AND (rr.effective_to IS NULL OR rr.effective_to > $2)
			AND cv.effective_from <= $2 AND (cv.effective_to IS NULL OR cv.effective_to > $2)
			LIMIT 1
		)::text,''),
		COALESCE((
			SELECT pv.name FROM reward.requirements rr
			JOIN reward.condition_versions cv ON cv.reward_condition_id=rr.reward_condition_id
			CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'card_plan_ids') value
			JOIN catalog.card_plans p ON p.id=value::uuid AND p.plan_type='selectable'
			JOIN catalog.card_plan_versions pv ON pv.card_plan_id=p.id
				AND pv.effective_from <= $2 AND (pv.effective_to IS NULL OR pv.effective_to > $2)
			WHERE rr.reward_component_id=$1 AND rr.effective_from <= $2 AND (rr.effective_to IS NULL OR rr.effective_to > $2)
			AND cv.effective_from <= $2 AND (cv.effective_to IS NULL OR cv.effective_to > $2)
			LIMIT 1
		),''),
		COALESCE(ARRAY(
			SELECT value FROM reward.requirements rr
			JOIN reward.condition_versions cv ON cv.reward_condition_id=rr.reward_condition_id
			CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'category_ids') value
			WHERE rr.reward_component_id=$1 AND rr.effective_from <= $2 AND (rr.effective_to IS NULL OR rr.effective_to > $2)
			AND cv.effective_from <= $2 AND (cv.effective_to IS NULL OR cv.effective_to > $2)
		),'{}'),
		COALESCE(ARRAY(
			SELECT value FROM reward.requirements rr
			JOIN reward.condition_versions cv ON cv.reward_condition_id=rr.reward_condition_id
			CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'merchant_ids') value
			WHERE rr.reward_component_id=$1 AND rr.effective_from <= $2 AND (rr.effective_to IS NULL OR rr.effective_to > $2)
			AND cv.effective_from <= $2 AND (cv.effective_to IS NULL OR cv.effective_to > $2)
		),'{}'),
		COALESCE(ARRAY(
			SELECT value FROM reward.requirements rr
			JOIN reward.condition_versions cv ON cv.reward_condition_id=rr.reward_condition_id
			CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'payment_method_ids') value
			WHERE rr.reward_component_id=$1 AND rr.effective_from <= $2 AND (rr.effective_to IS NULL OR rr.effective_to > $2)
			AND cv.effective_from <= $2 AND (cv.effective_to IS NULL OR cv.effective_to > $2)
		),'{}')`, rule.ID, at).Scan(&rule.CardNetworkIDs, &rule.QualifiedCardPlanIDs, &rule.SuggestedCardPlanID, &rule.SuggestedPlanName, &categories, &merchants, &paymentMethods); err != nil {
		return fmt.Errorf("load versioned rule conditions: %w", err)
	}
	if len(categories) > 0 {
		rule.CategoryID = categories
	}
	if len(merchants) > 0 {
		rule.MerchantIDs = merchants
	}
	if len(paymentMethods) > 0 {
		rule.PaymentMethods = paymentMethods
	}
	return nil
}

func loadPreferences(ctx context.Context, tx pgx.Tx, userID recommendations.ID) (map[recommendations.ID]recommendations.Decimal, error) {
	rows, err := tx.Query(ctx, `SELECT u.id, COALESCE(p.weight, 1)::text
		FROM reward_units u LEFT JOIN reward_preferences p
		ON p.reward_unit_id = u.id AND p.user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("query preferences: %w", err)
	}
	defer rows.Close()
	result := make(map[recommendations.ID]recommendations.Decimal)
	for rows.Next() {
		var id recommendations.ID
		var weight string
		if err := rows.Scan(&id, &weight); err != nil {
			return nil, fmt.Errorf("scan preference: %w", err)
		}
		result[id] = recommendations.MustDecimal(weight)
	}
	return result, rows.Err()
}

func loadMonthlyUsage(ctx context.Context, tx pgx.Tx, userID recommendations.ID, date recommendations.LocalDate) (map[recommendations.ID]recommendations.Decimal, error) {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	rows, err := tx.Query(ctx, `SELECT a.benefit_id, SUM(a.allocated_reward)::text
		FROM reward_allocations a JOIN transactions t ON t.id = a.transaction_id
		WHERE t.user_id = $1 AND t.transaction_date >= $2 AND t.transaction_date < $3
		GROUP BY a.benefit_id`, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query monthly usage: %w", err)
	}
	defer rows.Close()
	result := make(map[recommendations.ID]recommendations.Decimal)
	for rows.Next() {
		var id recommendations.ID
		var value string
		if err := rows.Scan(&id, &value); err != nil {
			return nil, fmt.Errorf("scan monthly usage: %w", err)
		}
		result[id] = recommendations.MustDecimal(value)
	}
	return result, rows.Err()
}

func insertAllocations(ctx context.Context, tx pgx.Tx, transactionID recommendations.ID, allocations []recommendations.RuleEvaluation) error {
	for _, allocation := range allocations {
		if _, err := tx.Exec(ctx, `INSERT INTO reward_allocations
			(id, transaction_id, benefit_id, reward_unit_id, uncapped_reward,
			allocated_reward, preference_weight, score)
			VALUES ($1, $2, $3, $4, $5::numeric, $6::numeric, $7::numeric, $8::numeric)`,
			newUUID(), transactionID, allocation.RuleID, allocation.RewardUnit.ID,
			allocation.UncappedReward.String(), allocation.AllocatedReward.String(),
			allocation.PreferenceWeight.String(), allocation.Score.String()); err != nil {
			return fmt.Errorf("insert allocation: %w", err)
		}
	}
	return nil
}

func insertRewardCalculation(ctx context.Context, tx pgx.Tx, transactionID recommendations.ID, input WriteInput, allocations []recommendations.RuleEvaluation, calculationType string) error {
	var previousID *recommendations.ID
	var calculationNo int
	if err := tx.QueryRow(ctx, `SELECT id,calculation_no FROM "transaction".reward_calculations
		WHERE transaction_id=$1 AND status='active' FOR UPDATE`, transactionID).Scan(&previousID, &calculationNo); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("load active reward calculation: %w", err)
	}
	if previousID != nil {
		if _, err := tx.Exec(ctx, `UPDATE "transaction".reward_calculations SET status='superseded' WHERE id=$1`, *previousID); err != nil {
			return err
		}
	}
	calculationNo++
	totalRate := recommendations.Decimal{}
	totalValue := recommendations.Decimal{}
	for _, allocation := range allocations {
		totalRate = totalRate.Add(allocation.RewardRate)
		twdRate := allocation.RewardUnit.TWDRate
		if twdRate.Sign() <= 0 {
			twdRate = recommendations.MustDecimal("1")
		}
		totalValue = totalValue.Add(allocation.AllocatedReward.Mul(twdRate))
	}
	snapshot, err := json.Marshal(map[string]any{
		"card_id": input.CardID, "amount_minor": input.AmountMinor, "category_id": input.CategoryID,
		"merchant_id": input.MerchantID, "merchant_name": input.MerchantName,
		"payment_method_id": input.PaymentMethodID, "transaction_date": input.TransactionDate.String(),
		"currency_id": "TWD", "country_id": "TW", "channel": "physical",
	})
	if err != nil {
		return err
	}
	calculationID := newUUID()
	if _, err := tx.Exec(ctx, `INSERT INTO "transaction".reward_calculations
		(id,transaction_id,calculation_no,calculation_type,transaction_at,total_effective_rate,total_reward_value,
		 engine_version,status,supersedes_calculation_id,input_snapshot_json)
		VALUES ($1,$2,$3,$4,$5,$6::numeric,$7::numeric,'reward-catalog-v1','active',$8,$9)`,
		calculationID, transactionID, calculationNo, calculationType, input.TransactionDate.Time,
		totalRate.String(), totalValue.String(), previousID, snapshot); err != nil {
		return fmt.Errorf("insert reward calculation: %w", err)
	}
	for _, allocation := range allocations {
		var versionID *recommendations.ID
		conditionSnapshot := []byte("[]")
		_ = tx.QueryRow(ctx, `SELECT v.id,COALESCE((
			SELECT jsonb_agg(jsonb_build_object(
				'type',c.condition_type,'operator',cv.operator,'configuration',cv.configuration_json,'description',cv.description
			) ORDER BY rr.display_order)
			FROM reward.requirements rr JOIN reward.conditions c ON c.id=rr.reward_condition_id
			JOIN reward.condition_versions cv ON cv.reward_condition_id=c.id
			WHERE rr.reward_component_id=$1 AND rr.effective_from <= $2 AND (rr.effective_to IS NULL OR rr.effective_to > $2)
			AND cv.effective_from <= $2 AND (cv.effective_to IS NULL OR cv.effective_to > $2)
		),'[]'::jsonb)
		FROM reward.component_versions v WHERE v.reward_component_id=$1
		AND v.effective_from <= $2 AND (v.effective_to IS NULL OR v.effective_to > $2)
			ORDER BY v.effective_from DESC LIMIT 1`, allocation.RuleID, input.TransactionDate.Time).Scan(&versionID, &conditionSnapshot)
		var componentID any
		if versionID != nil {
			componentID = allocation.RuleID
		}
		reminderItems := []string{}
		if allocation.ActionMessage != "" {
			reminderItems = append(reminderItems, allocation.ActionMessage)
		}
		reminders, _ := json.Marshal(reminderItems)
		var qualifiedPlanID, qualificationStatusID *recommendations.ID
		_ = tx.QueryRow(ctx, `SELECT p.id,q.id FROM catalog.card_plans p
			JOIN catalog.member_card_qualification_statuses q ON q.card_plan_id=p.id
			WHERE q.member_card_id=$1 AND q.is_qualified AND q.effective_from <= $2
			AND (q.effective_to IS NULL OR q.effective_to > $2)
			AND p.id IN (SELECT value::uuid FROM reward.requirements rr
				JOIN reward.condition_versions cv ON cv.reward_condition_id=rr.reward_condition_id
				CROSS JOIN LATERAL jsonb_array_elements_text(cv.configuration_json->'card_plan_ids') value
				WHERE rr.reward_component_id=$3)
			LIMIT 1`, input.CardID, input.TransactionDate.Time, allocation.RuleID).Scan(&qualifiedPlanID, &qualificationStatusID)
		var remaining any
		if allocation.RemainingBefore != nil {
			remaining = allocation.RemainingBefore.String()
		}
		if _, err := tx.Exec(ctx, `INSERT INTO "transaction".reward_calculation_components
			(id,reward_calculation_id,reward_component_id,reward_component_version_id,reward_unit_id,
			 suggested_card_plan_id,qualified_card_plan_id,member_qualification_status_id,effect_type,reward_value,reward_rate,uncapped_reward,
			 allocated_reward,cap_used_before,cap_remaining_before,condition_snapshot_json,reminder_snapshot_json)
			VALUES ($1,$2,$3,$4,$5,NULLIF($6,'')::uuid,$7,$8,$9,$10::numeric,$11::numeric,$12::numeric,$13::numeric,$14::numeric,$15,$16,$17)`,
			newUUID(), calculationID, componentID, versionID, allocation.RewardUnit.ID,
			allocation.SuggestedCardPlanID, qualifiedPlanID, qualificationStatusID, string(allocation.EffectType),
			allocation.RewardValue.String(), allocation.RewardRate.String(), allocation.UncappedReward.String(),
			allocation.AllocatedReward.String(), allocation.UsedBefore.String(),
			remaining, conditionSnapshot, reminders); err != nil {
			return fmt.Errorf("insert reward calculation component: %w", err)
		}
	}
	for _, eventType := range []string{"TransactionRecorded", "RewardCalculated"} {
		if _, err := tx.Exec(ctx, `INSERT INTO integration.outbox_events(id,aggregate_type,aggregate_id,event_type,payload_json)
			VALUES ($1,'transaction',$2,$3,$4)`, newUUID(), transactionID, eventType, snapshot); err != nil {
			return err
		}
	}
	return nil
}

func loadActivityMonthlyUsage(ctx context.Context, tx pgx.Tx, userID recommendations.ID, date recommendations.LocalDate) map[string]recommendations.Decimal {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	rows, err := tx.Query(ctx, `SELECT c.reward_program_id,a.reward_unit_id,SUM(a.allocated_reward)::text
		FROM reward_allocations a
		JOIN transactions t ON t.id=a.transaction_id
		JOIN reward.components c ON c.id=a.benefit_id
		WHERE t.user_id=$1 AND t.transaction_date >=$2 AND t.transaction_date<$3
		GROUP BY c.reward_program_id,a.reward_unit_id`, userID, start, end)
	if err != nil {
		return map[string]recommendations.Decimal{}
	}
	defer rows.Close()
	out := map[string]recommendations.Decimal{}
	for rows.Next() {
		var activity, unit recommendations.ID
		var value string
		if rows.Scan(&activity, &unit, &value) == nil {
			out[string(activity)+":"+string(unit)] = recommendations.MustDecimal(value)
		}
	}
	return out
}

func lockKeys(ctx context.Context, tx pgx.Tx, keys []string) error {
	sort.Strings(keys)
	var previous string
	for _, key := range keys {
		if key == previous {
			continue
		}
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, key); err != nil {
			return fmt.Errorf("acquire transaction lock: %w", err)
		}
		previous = key
	}
	return nil
}

func lockKey(userID, cardID recommendations.ID, date recommendations.LocalDate) string {
	return fmt.Sprintf("%s:%s:%04d-%02d", userID, cardID, date.Year(), date.Month())
}

func loadPaymentMethodName(ctx context.Context, tx pgx.Tx, id string) (string, error) {
	var name string
	if err := tx.QueryRow(ctx, `SELECT name FROM payment_methods WHERE id=$1 AND is_active`, id).Scan(&name); err != nil {
		return "", fmt.Errorf("active payment method not found")
	}
	return name, nil
}

func loadMerchantName(ctx context.Context, tx pgx.Tx, id string) (string, error) {
	var name string
	if err := tx.QueryRow(ctx, `SELECT name FROM merchants WHERE id=$1 AND is_active`, id).Scan(&name); err != nil {
		return "", fmt.Errorf("active merchant not found")
	}
	return name, nil
}

func loadPaymentMethods(ctx context.Context, tx pgx.Tx, userID recommendations.ID) ([]recommendations.PaymentMethod, error) {
	rows, err := tx.Query(ctx, `SELECT p.id,p.name FROM payment_methods p
		LEFT JOIN member_profile.user_payment_methods up ON up.payment_method_id=p.id AND up.user_id=$2
		WHERE p.is_active AND p.id<>$1
		AND (p.type IN ('physical_card','online_card') OR up.is_available)
		ORDER BY p.id`, recommendations.AnyPaymentID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []recommendations.PaymentMethod{}
	for rows.Next() {
		var item recommendations.PaymentMethod
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func toTransaction(id, userID recommendations.ID, input WriteInput, paymentMethodName string, allocations []recommendations.RuleEvaluation) Transaction {
	return Transaction{
		ID: id, UserID: userID, CardID: input.CardID, AmountMinor: input.AmountMinor,
		CategoryID: input.CategoryID, MerchantID: input.MerchantID, MerchantName: input.MerchantName, PaymentMethodID: input.PaymentMethodID, PaymentMethodName: paymentMethodName, TransactionDate: input.TransactionDate,
		Note: input.Note, Allocations: allocations,
	}
}

func newUUID() recommendations.ID {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(fmt.Sprintf("generate UUID: %v", err))
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	raw := hex.EncodeToString(value[:])
	return recommendations.ID(raw[0:8] + "-" + raw[8:12] + "-" + raw[12:16] + "-" + raw[16:20] + "-" + raw[20:32])
}
