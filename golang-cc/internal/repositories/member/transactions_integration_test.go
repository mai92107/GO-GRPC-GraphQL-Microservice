package member

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafa/golang-cc/internal/platform/config"
	"github.com/rafa/golang-cc/internal/services/recommendations"
	"github.com/rafa/golang-cc/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	userA     recommendations.ID = "10000000-0000-0000-0000-000000000001"
	userB     recommendations.ID = "10000000-0000-0000-0000-000000000002"
	cardA     recommendations.ID = "20000000-0000-0000-0000-000000000001"
	cardB     recommendations.ID = "20000000-0000-0000-0000-000000000002"
	ruleA     recommendations.ID = "30000000-0000-0000-0000-000000000001"
	activityA recommendations.ID = "30000000-0000-0000-0000-000000000002"
	cashID    recommendations.ID = "00000000-0000-0000-0000-000000000101"
	merchantA recommendations.ID = "70000000-0000-0000-0000-000000000001"
)

func TestTransactionLifecycleRestoresCapAndIsolatesUsers(t *testing.T) {
	pool, gormDB := integrationPool(t)
	service := NewTransactionRepository(pool)
	ctx := context.Background()

	june := WriteInput{
		CardID: cardA, AmountMinor: 10000, CategoryID: "dining",
		MerchantID: string(merchantA), MerchantName: "全聯福利中心", PaymentMethodID: "physical_card", TransactionDate: recommendations.MustLocalDate("2026-06-10"), Note: "June",
	}
	created, err := service.Create(ctx, userA, june)
	if err != nil {
		t.Fatal(err)
	}
	assertAllocation(t, created, "10")
	calculations, err := New(pool, gormDB).RewardCalculations(ctx, string(userA), string(created.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(calculations) != 1 || calculations[0].CalculationType != "original" {
		t.Fatalf("calculations=%+v", calculations)
	}

	if err := service.Delete(ctx, userB, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other user delete error = %v, want ErrNotFound", err)
	}

	second, err := service.Create(ctx, userA, june)
	if err != nil {
		t.Fatal(err)
	}
	assertAllocation(t, second, "0")

	july := june
	july.TransactionDate = recommendations.MustLocalDate("2026-07-01")
	updated, err := service.Update(ctx, userA, created.ID, july)
	if err != nil {
		t.Fatal(err)
	}
	assertAllocation(t, updated, "10")
	calculations, err = New(pool, gormDB).RewardCalculations(ctx, string(userA), string(created.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(calculations) != 2 || calculations[0].CalculationType != "manual_recalculation" || calculations[1].Status != "superseded" {
		t.Fatalf("recalculation history=%+v", calculations)
	}

	third, err := service.Create(ctx, userA, june)
	if err != nil {
		t.Fatal(err)
	}
	assertAllocation(t, third, "10")

	if err := service.Delete(ctx, userA, created.ID); err != nil {
		t.Fatal(err)
	}
	replacement, err := service.Create(ctx, userA, july)
	if err != nil {
		t.Fatal(err)
	}
	assertAllocation(t, replacement, "10")

	if _, err := pool.Exec(ctx, `DELETE FROM member_cards WHERE id = $1 AND user_id = $2`, cardA, userA); err == nil {
		t.Fatal("expected card used by transactions to be undeletable")
	}
	if _, err := pool.Exec(ctx, `DELETE FROM reward.published_reward_rules WHERE id = $1`, ruleA); err == nil {
		t.Fatal("expected rule used by allocations to be undeletable")
	}
}

func TestConcurrentCreatesDoNotExceedCap(t *testing.T) {
	pool, _ := integrationPool(t)
	service := NewTransactionRepository(pool)
	input := WriteInput{
		CardID: cardA, AmountMinor: 10000, CategoryID: "dining",
		MerchantID: string(merchantA), MerchantName: "全聯福利中心", PaymentMethodID: "physical_card", TransactionDate: recommendations.MustLocalDate("2026-06-10"),
	}

	var wait sync.WaitGroup
	wait.Add(2)
	results := make(chan Transaction, 2)
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			defer wait.Done()
			transaction, err := service.Create(context.Background(), userA, input)
			results <- transaction
			errs <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	total := recommendations.Decimal{}
	for result := range results {
		total = total.Add(result.Allocations[0].AllocatedReward)
	}
	if total.Cmp(recommendations.MustDecimal("10")) != 0 {
		t.Fatalf("concurrent allocated total = %s, want 10", total.String())
	}
}

func TestMemberCardCreditLimitHistoryOnUpdate(t *testing.T) {
	pool, gormDB := integrationPool(t)
	ctx := context.Background()
	repo := New(pool, gormDB)
	cardID := "26000000-0000-0000-0000-000000000001"
	if err := repo.CreateCard(ctx, cardID, string(userA), "50000000-0000-0000-0000-000000000002", CardWrite{
		IsActive:    true,
		Network:     "Visa",
		CreditLimit: "100000",
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateCard(ctx, cardID, string(userA), CardWrite{
		IsActive:    true,
		Network:     "Visa",
		CreditLimit: "150000",
	}); err != nil {
		t.Fatal(err)
	}

	rows, err := pool.Query(ctx, `SELECT credit_limit::text, effective_to IS NULL
		FROM member_card_credit_limits
		WHERE member_card_id=$1
		ORDER BY effective_from`, cardID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	type historyRow struct {
		limit   string
		current bool
	}
	var history []historyRow
	for rows.Next() {
		var row historyRow
		if err := rows.Scan(&row.limit, &row.current); err != nil {
			t.Fatal(err)
		}
		history = append(history, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 {
		t.Fatalf("history count=%d,want 2: %+v", len(history), history)
	}
	if history[0].limit != "100000.00" || history[0].current {
		t.Fatalf("old history row=%+v,want closed 100000.00", history[0])
	}
	if history[1].limit != "150000.00" || !history[1].current {
		t.Fatalf("current history row=%+v,want open 150000.00", history[1])
	}
}

func TestRecommendationUsesCreditLimitEffectiveOnTransactionDate(t *testing.T) {
	pool, _ := integrationPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `UPDATE member_cards SET credit_limit=15 WHERE id=$1`, cardA); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO member_card_credit_limits(member_card_id, credit_limit, effective_from, effective_to) VALUES
		($1, 5, '2026-06-01 00:00:00+08', '2026-06-15 12:00:00+08'),
		($1, 15, '2026-06-15 12:00:00+08', NULL)`, cardA); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reward.published_reward_rules
		SET cap_amount=NULL, cap_formula='member_card.credit_limit'
		WHERE id=$1`, ruleA); err != nil {
		t.Fatal(err)
	}
	txRepo := NewTransactionRepository(pool)
	tests := []struct {
		name string
		date string
		want string
	}{
		{name: "before limit change", date: "2026-06-10", want: "5"},
		{name: "same day as limit change uses current limit", date: "2026-06-15", want: "10"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := txRepo.Recommend(ctx, userA, 10000, "dining", string(merchantA), "全聯福利中心", recommendations.MustLocalDate(test.date))
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Recommendations) != 1 || len(result.Recommendations[0].Allocations) != 1 {
				t.Fatalf("recommendations=%+v", result.Recommendations)
			}
			got := result.Recommendations[0].Allocations[0].AllocatedReward
			if got.Cmp(recommendations.MustDecimal(test.want)) != 0 {
				t.Fatalf("allocated=%s,want %s", got.String(), test.want)
			}
		})
	}
}

func TestNormalizedCatalogConstraints(t *testing.T) {
	pool, gormDB := integrationPool(t)
	ctx := context.Background()
	var cards, benefits int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM card_products WHERE id::text LIKE '51000000-%'`).Scan(&cards); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reward.published_reward_rules`).Scan(&benefits); err != nil {
		t.Fatal(err)
	}
	if cards != 10 || benefits == 0 {
		t.Fatalf("core catalog cards=%d published benefits=%d", cards, benefits)
	}

	if err := New(pool, gormDB).CreateCard(ctx, "22000000-0000-0000-0000-000000000001", string(userA), "51000000-0000-0000-0000-000000000001", CardWrite{}); err == nil {
		t.Fatal("expected DAWHO account tier to be required")
	}
}

func TestQualificationHistoryConstraints(t *testing.T) {
	pool, gormDB := integrationPool(t)
	ctx := context.Background()
	clearSeedMemberCards(t, pool)
	seedDAWHOQualifiedPlan(t, pool, "大戶")
	repo := New(pool, gormDB)
	cardID := "25000000-0000-0000-0000-000000000001"
	if err := repo.CreateCard(ctx, cardID, string(userA), "51000000-0000-0000-0000-000000000001", CardWrite{
		IsActive: true, Network: "Visa", AccountTier: "大戶", CreditLimit: "100000",
	}); err != nil {
		t.Fatal(err)
	}
	var planID string
	planID = "52000000-0000-0000-0000-000000000001"
	effective := time.Date(2026, 6, 20, 12, 0, 0, 0, time.FixedZone("Asia/Taipei", 8*60*60))
	if err := repo.SetQualificationStatus(ctx, string(userA), cardID, planID, false, effective); err != nil {
		t.Fatal(err)
	}
	var historyCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM catalog.member_card_qualification_statuses
		WHERE member_card_id=$1 AND card_plan_id=$2`, cardID, planID).Scan(&historyCount); err != nil {
		t.Fatal(err)
	}
	if historyCount != 2 {
		t.Fatalf("qualification history count=%d,want 2", historyCount)
	}
}

func TestCoreCardRepresentativeRewards(t *testing.T) {
	pool, gormDB := integrationPool(t)
	ctx := context.Background()
	clearSeedMemberCards(t, pool)
	repo := New(pool, gormDB)
	cards := []struct{ id, product, tier string }{
		{"23000000-0000-0000-0000-000000000001", "51000000-0000-0000-0000-000000000001", "大戶Plus"},
		{"23000000-0000-0000-0000-000000000002", "51000000-0000-0000-0000-000000000002", ""},
		{"23000000-0000-0000-0000-000000000003", "51000000-0000-0000-0000-000000000003", ""},
		{"23000000-0000-0000-0000-000000000004", "51000000-0000-0000-0000-000000000004", ""},
	}
	for _, c := range cards {
		if err := repo.CreateCard(ctx, c.id, string(userA), c.product, CardWrite{IsActive: true, Network: "Visa", AccountTier: c.tier, CreditLimit: "100000"}); err != nil {
			t.Fatal(err)
		}
	}
	txRepo := NewTransactionRepository(pool)
	date := recommendations.MustLocalDate("2026-06-10")
	tests := []struct{ name, category, merchant, want, reminder string }{
		{"DAWHO 現金回饋信用卡", "dining", "一般餐廳", "50", "需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單"},
		{"SPORT 卡", "sports", "運動中心", "50", "需先完成當期活動登錄"},
		{"Richart 卡", "dining", "一般餐廳", "38", "需於 Richart Life APP 切換至符合消費情境的方案"},
		{"uniopen 聯名卡", "overseas", "海外實體商店", "110", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := txRepo.Recommend(ctx, userA, 100000, tc.category, "", tc.merchant, date)
			if err != nil {
				t.Fatal(err)
			}
			var found *recommendations.CardRecommendation
			for i := range result.Recommendations {
				if result.Recommendations[i].CardName == tc.name {
					found = &result.Recommendations[i]
					break
				}
			}
			if found == nil {
				t.Fatalf("card not recommended: %+v", result)
			}
			if found.TotalUnweighted.Cmp(recommendations.MustDecimal(tc.want)) != 0 {
				t.Fatalf("reward=%s,want %s", found.TotalUnweighted.String(), tc.want)
			}
			if tc.reminder != "" && (len(found.Reminders) != 1 || found.Reminders[0] != tc.reminder) {
				t.Fatalf("reminders=%v", found.Reminders)
			}
		})
	}
}

func TestDAWHO2026AccountTierRewards(t *testing.T) {
	pool, gormDB := integrationPool(t)
	ctx := context.Background()
	clearSeedMemberCards(t, pool)
	repo := New(pool, gormDB)
	txRepo := NewTransactionRepository(pool)
	date := recommendations.MustLocalDate("2026-06-10")
	tiers := []struct {
		name         string
		tier         string
		domesticWant string
		overseasWant string
		reminders    int
	}{
		{"大大", "大大", "10", "20", 0},
		{"大戶", "大戶", "35", "45", 1},
		{"大戶Plus", "大戶Plus", "50", "60", 1},
	}
	for index, tc := range tiers {
		t.Run(tc.name, func(t *testing.T) {
			cardID := fmt.Sprintf("24000000-0000-0000-0000-%012d", index+1)
			if _, err := pool.Exec(ctx, `DELETE FROM member_cards WHERE user_id=$1 AND card_product_id=$2`, userA, "51000000-0000-0000-0000-000000000001"); err != nil {
				t.Fatal(err)
			}
			if err := repo.CreateCard(ctx, cardID, string(userA), "51000000-0000-0000-0000-000000000001", CardWrite{IsActive: true, Network: "Visa", AccountTier: tc.tier, CreditLimit: "100000"}); err != nil {
				t.Fatal(err)
			}
			for _, scenario := range []struct {
				category string
				want     string
			}{
				{"dining", tc.domesticWant},
				{"overseas", tc.overseasWant},
			} {
				result, err := txRepo.Recommend(ctx, userA, 100000, scenario.category, "", "測試商店", date)
				if err != nil {
					t.Fatal(err)
				}
				var found *recommendations.CardRecommendation
				for i := range result.Recommendations {
					if result.Recommendations[i].CardID == recommendations.ID(cardID) {
						found = &result.Recommendations[i]
						break
					}
				}
				if found == nil {
					t.Fatalf("%s card not recommended: %+v", scenario.category, result)
				}
				if found.TotalUnweighted.Cmp(recommendations.MustDecimal(scenario.want)) != 0 {
					t.Fatalf("%s reward=%s,want %s", scenario.category, found.TotalUnweighted.String(), scenario.want)
				}
				if len(found.Reminders) != tc.reminders {
					t.Fatalf("%s reminders=%v,want %d", scenario.category, found.Reminders, tc.reminders)
				}
			}
		})
	}
}

func TestRecommendationDisambiguatesSameNicknameAndRanksEachCardOnce(t *testing.T) {
	pool, gormDB := integrationPool(t)
	ctx := context.Background()
	clearSeedMemberCards(t, pool)
	repo := New(pool, gormDB)
	for _, card := range []struct {
		id      string
		product string
		tier    string
	}{
		{"25000000-0000-0000-0000-000000000001", "51000000-0000-0000-0000-000000000001", "大大"},
		{"25000000-0000-0000-0000-000000000002", "51000000-0000-0000-0000-000000000007", ""},
	} {
		if err := repo.CreateCard(ctx, card.id, string(userA), card.product, CardWrite{Nickname: "LOL", IsActive: true, Network: "Visa", AccountTier: card.tier, CreditLimit: "100000"}); err != nil {
			t.Fatal(err)
		}
	}

	result, err := NewTransactionRepository(pool).Recommend(
		ctx, userA, 20000, "transport", "", "捷運", recommendations.MustLocalDate("2026-06-15"),
	)
	if err != nil {
		t.Fatal(err)
	}
	namesByID := map[recommendations.ID]string{}
	for _, item := range result.Recommendations {
		if _, exists := namesByID[item.CardID]; exists {
			t.Fatalf("card %s appeared more than once", item.CardID)
		}
		namesByID[item.CardID] = item.CardName
	}
	if namesByID["25000000-0000-0000-0000-000000000001"] != "LOL（DAWHO 現金回饋信用卡）" {
		t.Fatalf("unexpected DAWHO name: %q", namesByID["25000000-0000-0000-0000-000000000001"])
	}
	if namesByID["25000000-0000-0000-0000-000000000002"] != "LOL（英雄聯盟信用卡（已停止申辦））" {
		t.Fatalf("unexpected LOL card name: %q", namesByID["25000000-0000-0000-0000-000000000002"])
	}
}

func TestLatestCardActivitiesRespectPaymentMethodsAndStacking(t *testing.T) {
	pool, _ := integrationPool(t)
	ctx := context.Background()
	clearSeedMemberCards(t, pool)
	for _, card := range []struct{ id, product string }{
		{"91000000-0000-0000-0000-000000000005", "51000000-0000-0000-0000-000000000005"},
		{"91000000-0000-0000-0000-000000000006", "51000000-0000-0000-0000-000000000006"},
		{"91000000-0000-0000-0000-000000000007", "51000000-0000-0000-0000-000000000007"},
		{"91000000-0000-0000-0000-000000000008", "51000000-0000-0000-0000-000000000008"},
		{"91000000-0000-0000-0000-000000000009", "51000000-0000-0000-0000-000000000009"},
		{"91000000-0000-0000-0000-000000000010", "51000000-0000-0000-0000-000000000010"},
	} {
		if _, err := pool.Exec(ctx, `INSERT INTO member_cards(id,user_id,card_product_id,network,nickname,is_active) VALUES($1,$2,$3,'Visa','',true)`, card.id, userA, card.product); err != nil {
			t.Fatal(err)
		}
	}
	repository := NewTransactionRepository(pool)
	date := recommendations.MustLocalDate("2026-06-15")

	assertCardReward := func(cardName, category, merchant, method, want string) {
		t.Helper()
		var merchantID string
		if merchant != "" {
			if err := pool.QueryRow(ctx, `SELECT id FROM merchants WHERE name=$1`, merchant).Scan(&merchantID); errors.Is(err, pgx.ErrNoRows) {
				merchantID = ""
			} else if err != nil {
				t.Fatalf("merchant %q not found: %v", merchant, err)
			}
		}
		result, err := repository.Recommend(ctx, userA, 100000, category, merchantID, merchant, date)
		if err != nil {
			t.Fatal(err)
		}
		seen := []string{}
		for _, card := range result.Recommendations {
			for _, option := range card.PaymentOptions {
				seen = append(seen, fmt.Sprintf("%s/%s=%s allocations=%d", card.CardName, option.PaymentMethodID, option.TotalUnweighted.String(), len(option.Allocations)))
			}
			if card.CardName == cardName {
				for _, option := range card.PaymentOptions {
					if paymentOptionIncludes(option, method) && option.TotalUnweighted.Cmp(recommendations.MustDecimal(want)) == 0 {
						return
					}
				}
			}
		}
		t.Fatalf("%s method=%s reward=%s not recommended; merchant_id=%q seen=%v empty_reason=%q", cardName, method, want, merchantID, seen, result.EmptyReason)
	}

	assertCardReward("U Bear 信用卡", "online", "Steam", recommendations.AnyPaymentID, "100")
	assertCardReward("U Bear 信用卡", "dining", "一般餐廳", "line_pay", "30")
	assertCardReward("Unicard", "dining", "一般餐廳", "physical_card", "40")
	assertCardReward("英雄聯盟信用卡（已停止申辦）", "overseas", "海外實體商店", "apple_pay", "25")
	assertCardReward("英雄聯盟信用卡（已停止申辦）", "overseas", "海外實體商店", "line_pay", "10")
	assertCardReward("Pi 拍錢包信用卡", "dining", "一般餐廳", "physical_card", "10")
	assertCardReward("小小兵回饋卡", "dining", "一般餐廳", "physical_card", "12.34")
	assertCardReward("小小兵回饋卡", "grocery", "美廉社", "physical_card", "50")
	assertCardReward("小小兵回饋卡", "travel", "環球影城", "physical_card", "100")
	assertCardReward("LINE Bank聯名卡", "insurance", "保險公司", "physical_card", "10")
	assertCardReward("LINE Bank聯名卡", "overseas", "海外網路商店", "online_card", "25")
	assertCardReward("LINE Bank聯名卡", "online", "PChome", "online_card", "40")
}

func paymentOptionIncludes(option recommendations.PaymentOption, method string) bool {
	if option.PaymentMethodID == recommendations.AnyPaymentID {
		return true
	}
	if option.PaymentMethodID == method {
		return true
	}
	for _, candidate := range option.PaymentMethods {
		if candidate.ID == method {
			return true
		}
	}
	return false
}

func integrationPool(t *testing.T) (*pgxpool.Pool, *gorm.DB) {
	t.Helper()
	cfg, err := config.LoadFromProject("configs/test.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Database.ConnectionString())
	if err != nil {
		t.Fatal(err)
	}
	gormDB, err := gorm.Open(postgres.Open(cfg.Database.ConnectionString()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err = pool.Ping(ctx); err != nil {
		t.Skipf("database unavailable: %v", err)
	}

	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS identity,catalog,merchant,reward,member_profile,recommendation,"transaction",integration CASCADE`); err != nil {
		t.Fatalf("reset module schemas: %v", err)
	}
	if _, err := pool.Exec(ctx, `DROP SCHEMA public CASCADE`); err != nil {
		t.Fatalf("reset database: %v", err)
	}
	if _, err := pool.Exec(ctx, `CREATE SCHEMA public`); err != nil {
		t.Fatalf("reset database: %v", err)
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := migrations.Apply(ctx, connection.Conn()); err != nil {
		connection.Release()
		t.Fatalf("apply migrations: %v", err)
	}
	connection.Release()
	seedIntegrationData(t, pool)
	return pool, gormDB
}

func seedIntegrationData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	statements := []struct {
		sql  string
		args []any
	}{
		{`
		INSERT INTO users (id, email, password_hash, display_name, role, status) VALUES
			($1, 'a@example.test', 'hash', 'A', 'member', 'active'),
			($2, 'b@example.test', 'hash', 'B', 'member', 'active')`,
			[]any{userA, userB}},
		{`INSERT INTO member_profile.user_payment_methods(id,user_id,payment_method_id,is_available)
			SELECT md5('test-payment:'||$1::text||':'||id)::uuid,$1::uuid,id,true
			FROM payment_methods WHERE type IN ('mobile_payment','electronic_ticket')`, []any{userA}},
		{`INSERT INTO banks(id,name) VALUES('40000000-0000-0000-0000-000000000001','虛構銀行')`, nil},
		{`INSERT INTO card_products(id,bank_id,name,networks) VALUES
			('50000000-0000-0000-0000-000000000001','40000000-0000-0000-0000-000000000001','森活卡','Visa'),
			('50000000-0000-0000-0000-000000000002','40000000-0000-0000-0000-000000000001','青雲卡','Visa')`, nil},
		{`INSERT INTO member_cards(id,user_id,card_product_id,network,nickname,is_active) VALUES
			($3,$1,'50000000-0000-0000-0000-000000000001','Visa','森活卡',true),
			($4,$2,'50000000-0000-0000-0000-000000000002','Visa','青雲卡',true)`,
			[]any{userA, userB, cardA, cardB}},
		{`
		INSERT INTO reward_preferences (user_id, reward_unit_id, weight)
			VALUES ($1, $2, 1)`,
			[]any{userA, cashID}},
		{`INSERT INTO merchants(id,name,is_active,is_system) VALUES('70000000-0000-0000-0000-000000000001','全聯',true,false)`, nil},
		{`INSERT INTO merchant_aliases(merchant_id,alias) VALUES('70000000-0000-0000-0000-000000000001','全聯福利中心')`, nil},
		{`
		INSERT INTO reward.activities
			(id, bank_id, card_product_id, title, effective_from, effective_to, is_active, published_at, published_checksum)
			VALUES ($1, '40000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001',
			        '測試活動', '2026-01-01', '2026-12-31', true, now(), 'seed')`,
			[]any{activityA}},
		{`
		INSERT INTO reward.activity_groups(id, activity_id, name, display_order, is_active)
			VALUES ('30000000-0000-0000-0000-000000000003', $1, '基本回饋', 10, true)`,
			[]any{activityA}},
		{`
		INSERT INTO reward.activity_components(id, name, layer, stack_group, stack_mode, priority, effective_from, effective_to, is_active)
			VALUES ('30000000-0000-0000-0000-000000000004', '餐飲 10%', 1, 'base', 'ADDITIVE', 10, '2026-01-01', '2026-12-31', true)`,
			nil},
		{`
		INSERT INTO reward.activity_component_groups(reward_component_id, reward_group_id)
			VALUES ('30000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000003')`,
			nil},
		{`
		INSERT INTO reward.activity_requirements(id, reward_component_id, requirement_type, operator, configuration_json, description, is_active)
			VALUES ('30000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000004',
			        'CONSUMPTION_CATEGORY', 'IN', '{"category_ids":["dining"]}', '限餐飲', true)`,
			nil},
		{`
		INSERT INTO reward.activity_benefits(id, reward_component_id, benefit_type, value, reward_unit_id, cap_amount, cap_period, description, is_active)
			VALUES ($1, '30000000-0000-0000-0000-000000000004', 'RATE_CASHBACK', 0.1, $2, 10, 'calendar_month', '餐飲 10%', true)`,
			[]any{ruleA, cashID}},
		{`
		INSERT INTO reward.published_activities(id, source_activity_id, bank_id, card_product_id, title, effective_from, effective_to, published_at, source_checksum)
			VALUES ($1, $1, '40000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001',
			        '測試活動', '2026-01-01', '2026-12-31', now(), 'seed')`,
			[]any{activityA}},
		{`
		INSERT INTO reward.published_reward_rules(
			id,published_activity_id,source_activity_id,source_component_id,source_benefit_id,
			name,layer,display_order,stack_group,stack_policy,priority,effect_type,reward_value,reward_unit_id,
			cap_amount,cap_period,effective_from,effective_to,is_active)
			VALUES ($1,$2,$2,'30000000-0000-0000-0000-000000000004',$1,
			        '餐飲 10%',1,10,'base','stack',10,'ADD_RATE',0.1,$3,10,'calendar_month','2026-01-01','2026-12-31',true)`,
			[]any{ruleA, activityA, cashID}},
		{`
		INSERT INTO reward.published_rule_requirements(id,published_rule_id,source_requirement_id,requirement_type,operator,configuration_json,description)
			VALUES ('30000000-0000-0000-0000-000000000006',$1,'30000000-0000-0000-0000-000000000005',
			        'CONSUMPTION_CATEGORY','IN','{"category_ids":["dining"]}','限餐飲')`,
			[]any{ruleA}},
		{`
		INSERT INTO reward.published_rule_benefits(id,published_rule_id,source_benefit_id,benefit_type,value,reward_unit_id,cap_amount,cap_period,description)
			VALUES ($1,$1,$1,'RATE_CASHBACK',0.1,$2,10,'calendar_month','餐飲 10%')`,
			[]any{ruleA, cashID}},
	}
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatalf("seed integration data: %v", err)
		}
	}
	seedCorePublishedRuntime(t, pool)
}

func seedCorePublishedRuntime(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	rules := []struct {
		productID     string
		ruleName      string
		category      string
		paymentMethod string
		rate          string
		accountTier   string
		actionMessage string
	}{
		{"51000000-0000-0000-0000-000000000001", "DAWHO 大大國內", "dining", "", "0.01", "大大", ""},
		{"51000000-0000-0000-0000-000000000001", "DAWHO 大大海外", "overseas", "", "0.02", "大大", ""},
		{"51000000-0000-0000-0000-000000000001", "DAWHO 大戶國內", "dining", "", "0.035", "大戶", "需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單"},
		{"51000000-0000-0000-0000-000000000001", "DAWHO 大戶海外", "overseas", "", "0.045", "大戶", "需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單"},
		{"51000000-0000-0000-0000-000000000001", "DAWHO 大戶Plus國內", "dining", "", "0.05", "大戶Plus", "需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單"},
		{"51000000-0000-0000-0000-000000000001", "DAWHO 大戶Plus海外", "overseas", "", "0.06", "大戶Plus", "需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單"},
		{"51000000-0000-0000-0000-000000000001", "DAWHO 交通", "transport", "", "0.01", "大大", ""},
		{"51000000-0000-0000-0000-000000000002", "SPORT 運動", "sports", "", "0.05", "", "需先完成當期活動登錄"},
		{"51000000-0000-0000-0000-000000000003", "Richart 餐飲", "dining", "", "0.038", "", "需於 Richart Life APP 切換至符合消費情境的方案"},
		{"51000000-0000-0000-0000-000000000004", "uniopen 海外", "overseas", "", "0.11", "", ""},
		{"51000000-0000-0000-0000-000000000005", "U Bear 線上", "online", "", "0.10", "", ""},
		{"51000000-0000-0000-0000-000000000005", "U Bear 餐飲 LINE Pay", "dining", "line_pay", "0.03", "", ""},
		{"51000000-0000-0000-0000-000000000006", "Unicard 餐飲實體", "dining", "physical_card", "0.04", "", ""},
		{"51000000-0000-0000-0000-000000000007", "LOL 海外 Apple Pay", "overseas", "apple_pay", "0.025", "", ""},
		{"51000000-0000-0000-0000-000000000007", "LOL 海外 LINE Pay", "overseas", "line_pay", "0.01", "", ""},
		{"51000000-0000-0000-0000-000000000007", "LOL 交通", "transport", "", "0.01", "", ""},
		{"51000000-0000-0000-0000-000000000008", "Pi 餐飲實體", "dining", "physical_card", "0.01", "", ""},
		{"51000000-0000-0000-0000-000000000009", "小小兵餐飲", "dining", "physical_card", "0.01234", "", ""},
		{"51000000-0000-0000-0000-000000000009", "小小兵量販", "grocery", "physical_card", "0.05", "", ""},
		{"51000000-0000-0000-0000-000000000009", "小小兵旅遊", "travel", "physical_card", "0.10", "", ""},
		{"51000000-0000-0000-0000-000000000010", "LINE Bank 保險", "insurance", "physical_card", "0.01", "", ""},
		{"51000000-0000-0000-0000-000000000010", "LINE Bank 海外網路", "overseas", "online_card", "0.025", "", ""},
		{"51000000-0000-0000-0000-000000000010", "LINE Bank 線上", "online", "online_card", "0.04", "", ""},
	}
	for _, rule := range rules {
		seedPublishedRule(t, pool, rule.productID, rule.ruleName, rule.category, rule.paymentMethod, rule.rate, rule.accountTier, rule.actionMessage)
	}
}

func seedPublishedRule(t *testing.T, pool *pgxpool.Pool, productID, ruleName, category, paymentMethod, rate, accountTier, actionMessage string) {
	t.Helper()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `
WITH ids AS (
	SELECT
		md5('core-activity:' || $1)::uuid AS activity_id,
		md5('core-group:' || $1)::uuid AS group_id,
		md5('core-component:' || $1 || ':' || $2)::uuid AS component_id,
		md5('core-benefit:' || $1 || ':' || $2)::uuid AS benefit_id,
		md5('core-category-req:' || $1 || ':' || $2)::uuid AS category_req_id,
		md5('core-payment-req:' || $1 || ':' || $2)::uuid AS payment_req_id,
		md5('core-tier-req:' || $1 || ':' || $2)::uuid AS tier_req_id,
		md5('core-action-req:' || $1 || ':' || $2)::uuid AS action_req_id
), activity AS (
	INSERT INTO reward.activities(id,bank_id,card_product_id,title,effective_from,effective_to,is_active,published_at,published_checksum)
	SELECT ids.activity_id,cp.bank_id,cp.id,'核心代表性 published runtime','2026-01-01','2026-12-31',true,now(),'core-seed'
	FROM ids JOIN catalog.card_products cp ON cp.id=$1::uuid
	ON CONFLICT(id) DO NOTHING
	RETURNING id
), grp AS (
	INSERT INTO reward.activity_groups(id,activity_id,name,display_order,is_active)
	SELECT ids.group_id,ids.activity_id,'核心代表性',10,true FROM ids
	ON CONFLICT(id) DO NOTHING
	RETURNING id
), component AS (
	INSERT INTO reward.activity_components(id,name,layer,stack_group,stack_mode,priority,effective_from,effective_to,is_active)
	SELECT ids.component_id,$2::text,1,'core-' || $2::text,'ADDITIVE',10,'2026-01-01','2026-12-31',true FROM ids
	ON CONFLICT(id) DO NOTHING
	RETURNING id
), link AS (
	INSERT INTO reward.activity_component_groups(reward_component_id,reward_group_id)
	SELECT ids.component_id,ids.group_id FROM ids
	ON CONFLICT DO NOTHING
), benefit AS (
	INSERT INTO reward.activity_benefits(id,reward_component_id,benefit_type,value,reward_unit_id,description,is_active)
	SELECT ids.benefit_id,ids.component_id,'RATE_CASHBACK',$5::numeric,$8::uuid,$2::text,true FROM ids
	ON CONFLICT(id) DO NOTHING
), cat_req AS (
	INSERT INTO reward.activity_requirements(id,reward_component_id,requirement_type,operator,configuration_json,description,is_active)
	SELECT ids.category_req_id,ids.component_id,'CONSUMPTION_CATEGORY','IN',jsonb_build_object('category_ids',jsonb_build_array($3::text)),$3::text,true FROM ids
	ON CONFLICT(id) DO NOTHING
), pay_req AS (
	INSERT INTO reward.activity_requirements(id,reward_component_id,requirement_type,operator,configuration_json,description,is_active)
	SELECT ids.payment_req_id,ids.component_id,'PAYMENT_METHOD','IN',jsonb_build_object('payment_method_codes',jsonb_build_array($4::text)),$4::text,true FROM ids
	WHERE $4::text <> ''
	ON CONFLICT(id) DO NOTHING
), tier_req AS (
	INSERT INTO reward.activity_requirements(id,reward_component_id,requirement_type,operator,configuration_json,description,is_active)
	SELECT ids.tier_req_id,ids.component_id,'ACCOUNT_TIER','IN',jsonb_build_object('tiers',jsonb_build_array($6::text)),$6::text,true FROM ids
	WHERE $6::text <> ''
	ON CONFLICT(id) DO NOTHING
), action_req AS (
	INSERT INTO reward.activity_requirements(id,reward_component_id,requirement_type,operator,configuration_json,description,is_active)
	SELECT ids.action_req_id,ids.component_id,'ACTION_REQUIRED','IN',jsonb_build_object('action_codes',jsonb_build_array($7::text)),$7::text,true FROM ids
	WHERE $7::text <> ''
	ON CONFLICT(id) DO NOTHING
), pub_activity AS (
	INSERT INTO reward.published_activities(id,source_activity_id,bank_id,card_product_id,title,effective_from,effective_to,published_at,source_checksum)
	SELECT ids.activity_id,ids.activity_id,cp.bank_id,cp.id,'核心代表性 published runtime','2026-01-01','2026-12-31',now(),'core-seed'
	FROM ids JOIN catalog.card_products cp ON cp.id=$1::uuid
	ON CONFLICT(id) DO NOTHING
), pub_rule AS (
	INSERT INTO reward.published_reward_rules(id,published_activity_id,source_activity_id,source_component_id,source_benefit_id,name,layer,display_order,stack_group,stack_policy,priority,effect_type,reward_value,reward_unit_id,effective_from,effective_to,is_active)
	SELECT ids.benefit_id,ids.activity_id,ids.activity_id,ids.component_id,ids.benefit_id,$2::text,1,10,'core-' || $2::text,'stack',10,'ADD_RATE',$5::numeric,$8::uuid,'2026-01-01','2026-12-31',true FROM ids
	ON CONFLICT(id) DO NOTHING
), pub_cat_req AS (
	INSERT INTO reward.published_rule_requirements(id,published_rule_id,source_requirement_id,requirement_type,operator,configuration_json,description)
	SELECT ids.category_req_id,ids.benefit_id,ids.category_req_id,'CONSUMPTION_CATEGORY','IN',jsonb_build_object('category_ids',jsonb_build_array($3::text)),$3::text FROM ids
	ON CONFLICT(id) DO NOTHING
), pub_pay_req AS (
	INSERT INTO reward.published_rule_requirements(id,published_rule_id,source_requirement_id,requirement_type,operator,configuration_json,description)
	SELECT ids.payment_req_id,ids.benefit_id,ids.payment_req_id,'PAYMENT_METHOD','IN',jsonb_build_object('payment_method_codes',jsonb_build_array($4::text)),$4::text FROM ids
	WHERE $4::text <> ''
	ON CONFLICT(id) DO NOTHING
), pub_tier_req AS (
	INSERT INTO reward.published_rule_requirements(id,published_rule_id,source_requirement_id,requirement_type,operator,configuration_json,description)
	SELECT ids.tier_req_id,ids.benefit_id,ids.tier_req_id,'ACCOUNT_TIER','IN',jsonb_build_object('tiers',jsonb_build_array($6::text)),$6::text FROM ids
	WHERE $6::text <> ''
	ON CONFLICT(id) DO NOTHING
), pub_action_req AS (
	INSERT INTO reward.published_rule_requirements(id,published_rule_id,source_requirement_id,requirement_type,operator,configuration_json,description)
	SELECT ids.action_req_id,ids.benefit_id,ids.action_req_id,'ACTION_REQUIRED','IN',jsonb_build_object('action_codes',jsonb_build_array($7::text)),$7::text FROM ids
	WHERE $7::text <> ''
	ON CONFLICT(id) DO NOTHING
)
INSERT INTO reward.published_rule_benefits(id,published_rule_id,source_benefit_id,benefit_type,value,reward_unit_id,description)
SELECT ids.benefit_id,ids.benefit_id,ids.benefit_id,'RATE_CASHBACK',$5::numeric,$8::uuid,$2::text FROM ids
ON CONFLICT(id) DO NOTHING`, productID, ruleName, category, paymentMethod, rate, accountTier, actionMessage, cashID); err != nil {
		t.Fatal(err)
	}
}

func clearSeedMemberCards(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `DELETE FROM member_cards WHERE user_id=$1`, userA); err != nil {
		t.Fatal(err)
	}
}

func seedDAWHOQualifiedPlan(t *testing.T, pool *pgxpool.Pool, name string) {
	t.Helper()
	ctx := context.Background()
	planID := "52000000-0000-0000-0000-000000000001"
	if _, err := pool.Exec(ctx, `UPDATE catalog.card_products SET qualified_type=$1 WHERE id='51000000-0000-0000-0000-000000000001'`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO catalog.card_plans(id,card_product_id,plan_type,is_active,display_order)
		VALUES($1,'51000000-0000-0000-0000-000000000001','qualified',true,1)
		ON CONFLICT(id) DO UPDATE SET is_active=true`, planID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO catalog.card_plan_versions(id,card_plan_id,name,description,effective_from,published_at)
		VALUES('52000000-0000-0000-0000-000000000101',$1,$2,'測試資格','2000-01-01',now())
		ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name`, planID, name); err != nil {
		t.Fatal(err)
	}
}

func assertAllocation(t *testing.T, transaction Transaction, expected string) {
	t.Helper()
	if len(transaction.Allocations) != 1 {
		t.Fatalf("allocation count = %d, want 1", len(transaction.Allocations))
	}
	if transaction.Allocations[0].AllocatedReward.Cmp(recommendations.MustDecimal(expected)) != 0 {
		t.Fatalf("allocated = %s, want %s", transaction.Allocations[0].AllocatedReward.String(), expected)
	}
}
