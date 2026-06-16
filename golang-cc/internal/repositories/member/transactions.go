package member

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	rows, err := s.pool.Query(ctx, `SELECT t.id,t.card_id,t.amount_minor,t.category_code,COALESCE(t.merchant_code,''),t.merchant_name,t.payment_method_code,p.name,t.transaction_date::text,COALESCE(t.note,'')
		FROM transactions t JOIN payment_methods p ON p.code=t.payment_method_code
		WHERE t.user_id=$1 AND ($2='' OR t.id::text=$2) ORDER BY t.transaction_date DESC,t.created_at DESC`, userID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.TransactionSummary{}
	for rows.Next() {
		var item domain.TransactionSummary
		if err := rows.Scan(&item.ID, &item.CardID, &item.AmountMinor, &item.CategoryCode, &item.MerchantCode, &item.MerchantName, &item.PaymentMethodCode, &item.PaymentMethodName, &item.TransactionDate, &item.Note); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (s *TransactionRepository) Recommend(ctx context.Context, userID recommendations.ID, amountMinor int64, category, merchantCode, merchantName string, date recommendations.LocalDate) (recommendations.Result, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return recommendations.Result{}, err
	}
	defer tx.Rollback(ctx)
	if merchantCode != "" {
		if _, err := loadMerchantName(ctx, tx, merchantCode); err != nil {
			return recommendations.Result{}, err
		}
	}
	rows, err := tx.Query(ctx, `SELECT mc.id, mc.user_id,
		CASE WHEN NULLIF(mc.nickname,'') IS NULL THEN cp.name ELSE mc.nickname || '（' || cp.name || '）' END,
		COALESCE(mc.account_tier,''), mc.is_active
		FROM member_cards mc JOIN card_products cp ON cp.id=mc.card_product_id
		WHERE mc.user_id = $1 AND mc.is_active AND cp.is_active`, userID)
	if err != nil {
		return recommendations.Result{}, err
	}
	var cards []recommendations.Card
	for rows.Next() {
		var card recommendations.Card
		if err := rows.Scan(&card.ID, &card.UserID, &card.Name, &card.AccountTier, &card.IsActive); err != nil {
			rows.Close()
			return recommendations.Result{}, err
		}
		cards = append(cards, card)
	}
	rows.Close()
	var rules []recommendations.RewardRule
	for _, card := range cards {
		cardRules, err := loadRules(ctx, tx, userID, card.ID)
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
	methods, err := loadPaymentMethods(ctx, tx)
	if err != nil {
		return recommendations.Result{}, err
	}
	return recommendations.Recommend(recommendations.RecommendationInput{
		UserID: userID, AmountMinor: amountMinor, CategoryCode: category, MerchantCode: merchantCode, MerchantName: merchantName, PaymentMethods: methods, Date: date,
		Cards: cards, Rules: rules, Preferences: preferences, MonthlyUsage: usage, ActivityMonthlyUsage: loadActivityMonthlyUsage(ctx, tx, userID, date),
	})
}

func (s *TransactionRepository) Create(ctx context.Context, userID recommendations.ID, input WriteInput) (Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Transaction{}, fmt.Errorf("begin create transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if input.MerchantCode != "" {
		input.MerchantName, err = loadMerchantName(ctx, tx, input.MerchantCode)
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
	paymentMethodName, err := loadPaymentMethodName(ctx, tx, input.PaymentMethodCode)
	if err != nil {
		return Transaction{}, err
	}
	transactionID := newUUID()
	if _, err := tx.Exec(ctx, `INSERT INTO transactions
		(id, user_id, card_id, amount_minor, category_code, merchant_code, merchant_name, payment_method_code, transaction_date, note)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,''), $7, $8, $9, NULLIF($10, ''))`,
		transactionID, userID, input.CardID, input.AmountMinor, input.CategoryCode,
		input.MerchantCode, input.MerchantName, input.PaymentMethodCode, input.TransactionDate.Time, input.Note); err != nil {
		return Transaction{}, fmt.Errorf("insert transaction: %w", err)
	}
	if err := insertAllocations(ctx, tx, transactionID, result.Allocations); err != nil {
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
	if input.MerchantCode != "" {
		input.MerchantName, err = loadMerchantName(ctx, tx, input.MerchantCode)
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
		card_id = $3, amount_minor = $4, category_code = $5, merchant_code=NULLIF($6,''), merchant_name = $7, payment_method_code = $8, transaction_date = $9,
		note = NULLIF($10, ''), updated_at = now()
		WHERE id = $1 AND user_id = $2`,
		transactionID, userID, input.CardID, input.AmountMinor, input.CategoryCode,
		input.MerchantCode, input.MerchantName, input.PaymentMethodCode, input.TransactionDate.Time, input.Note); err != nil {
		return Transaction{}, fmt.Errorf("update transaction: %w", err)
	}
	result, err := buildRecommendation(ctx, tx, userID, input)
	if err != nil {
		return Transaction{}, err
	}
	paymentMethodName, err := loadPaymentMethodName(ctx, tx, input.PaymentMethodCode)
	if err != nil {
		return Transaction{}, err
	}
	if err := insertAllocations(ctx, tx, transactionID, result.Allocations); err != nil {
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
	paymentMethodName, err := loadPaymentMethodName(ctx, tx, input.PaymentMethodCode)
	if err != nil {
		return recommendations.CardRecommendation{}, err
	}
	cards, err := loadCard(ctx, tx, userID, input.CardID)
	if err != nil {
		return recommendations.CardRecommendation{}, err
	}
	rules, err := loadRules(ctx, tx, userID, input.CardID)
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
		UserID: userID, AmountMinor: input.AmountMinor, CategoryCode: input.CategoryCode, MerchantCode: input.MerchantCode, MerchantName: input.MerchantName,
		PaymentMethods: []recommendations.PaymentMethod{{Code: input.PaymentMethodCode, Name: paymentMethodName}},
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

func loadCard(ctx context.Context, tx pgx.Tx, userID, cardID recommendations.ID) ([]recommendations.Card, error) {
	var card recommendations.Card
	if err := tx.QueryRow(ctx, `SELECT mc.id, mc.user_id,
		CASE WHEN NULLIF(mc.nickname,'') IS NULL THEN cp.name ELSE mc.nickname || '（' || cp.name || '）' END,
		COALESCE(mc.account_tier,''), mc.is_active
		FROM member_cards mc JOIN card_products cp ON cp.id=mc.card_product_id
		WHERE mc.id = $1 AND mc.user_id = $2 AND mc.is_active AND cp.is_active`, cardID, userID).
		Scan(&card.ID, &card.UserID, &card.Name, &card.AccountTier, &card.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("active card not found")
		}
		return nil, fmt.Errorf("load card: %w", err)
	}
	return []recommendations.Card{card}, nil
}

func loadRules(ctx context.Context, tx pgx.Tx, userID, cardID recommendations.ID) ([]recommendations.RewardRule, error) {
	rows, err := tx.Query(ctx, `SELECT b.id,a.id,a.name,$1::uuid,mc.id,b.name,b.rate::text,
		b.monthly_cap::text,a.start_date,a.end_date,(a.is_active AND b.is_active),b.stack_group,b.priority,b.required_account_tiers,b.action_required,b.action_message,
		ARRAY(SELECT p.payment_method_code FROM card_activity_benefit_payment_methods p WHERE p.benefit_id=b.id ORDER BY p.payment_method_code),
		a.shared_monthly_caps->>u.id::text,
		u.id, u.code, u.name, u.symbol, u.symbol_position, u.twd_rate::text, u.precision,
		array_agg(c.category_code ORDER BY c.category_code),
		COALESCE(array_agg(DISTINCT m.merchant_code ORDER BY m.merchant_code) FILTER (WHERE m.merchant_code IS NOT NULL), '{}')
		FROM member_cards mc
		JOIN card_activities a ON a.card_product_id = mc.card_product_id
		JOIN card_activity_benefits b ON b.card_activity_id=a.id
		JOIN reward_units u ON u.id = b.reward_unit_id
		JOIN card_activity_benefit_categories c ON c.benefit_id = b.id
		LEFT JOIN card_activity_benefit_merchants m ON m.benefit_id = b.id
		WHERE mc.user_id = $1 AND mc.id = $2
		GROUP BY b.id,a.id,u.id,mc.id`, userID, cardID)
	if err != nil {
		return nil, fmt.Errorf("query rules: %w", err)
	}
	defer rows.Close()

	var rules []recommendations.RewardRule
	for rows.Next() {
		var rule recommendations.RewardRule
		var rate string
		var cap *string
		var sharedCap *string
		var twdRate string
		var start, end *time.Time
		if err := rows.Scan(&rule.ID, &rule.ActivityID, &rule.ActivityName, &rule.UserID, &rule.CardID, &rule.Name, &rate,
			&cap, &start, &end, &rule.IsActive, &rule.StackGroup, &rule.Priority, &rule.RequiredAccountTiers, &rule.ActionRequired, &rule.ActionMessage, &rule.PaymentMethods, &sharedCap, &rule.RewardUnit.ID, &rule.RewardUnit.Code,
			&rule.RewardUnit.Name, &rule.RewardUnit.Symbol, &rule.RewardUnit.SymbolPosition, &twdRate, &rule.RewardUnit.Precision,
			&rule.CategoryCode, &rule.MerchantCodes); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}
		rule.Rate = recommendations.MustDecimal(rate)
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
	return rules, rows.Err()
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

func loadActivityMonthlyUsage(ctx context.Context, tx pgx.Tx, userID recommendations.ID, date recommendations.LocalDate) map[string]recommendations.Decimal {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	rows, err := tx.Query(ctx, `SELECT b.card_activity_id,a.reward_unit_id,SUM(a.allocated_reward)::text FROM reward_allocations a JOIN transactions t ON t.id=a.transaction_id JOIN card_activity_benefits b ON b.id=a.benefit_id WHERE t.user_id=$1 AND t.transaction_date >=$2 AND t.transaction_date<$3 GROUP BY b.card_activity_id,a.reward_unit_id`, userID, start, end)
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

func loadPaymentMethodName(ctx context.Context, tx pgx.Tx, code string) (string, error) {
	var name string
	if err := tx.QueryRow(ctx, `SELECT name FROM payment_methods WHERE code=$1 AND is_active`, code).Scan(&name); err != nil {
		return "", fmt.Errorf("active payment method not found")
	}
	return name, nil
}

func loadMerchantName(ctx context.Context, tx pgx.Tx, code string) (string, error) {
	var name string
	if err := tx.QueryRow(ctx, `SELECT name FROM merchants WHERE code=$1 AND is_active`, code).Scan(&name); err != nil {
		return "", fmt.Errorf("active merchant not found")
	}
	return name, nil
}

func loadPaymentMethods(ctx context.Context, tx pgx.Tx) ([]recommendations.PaymentMethod, error) {
	rows, err := tx.Query(ctx, `SELECT code,name FROM payment_methods WHERE is_active AND code<>$1 ORDER BY code`, recommendations.AnyPaymentCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []recommendations.PaymentMethod{}
	for rows.Next() {
		var item recommendations.PaymentMethod
		if err := rows.Scan(&item.Code, &item.Name); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func toTransaction(id, userID recommendations.ID, input WriteInput, paymentMethodName string, allocations []recommendations.RuleEvaluation) Transaction {
	return Transaction{
		ID: id, UserID: userID, CardID: input.CardID, AmountMinor: input.AmountMinor,
		CategoryCode: input.CategoryCode, MerchantCode: input.MerchantCode, MerchantName: input.MerchantName, PaymentMethodCode: input.PaymentMethodCode, PaymentMethodName: paymentMethodName, TransactionDate: input.TransactionDate,
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
