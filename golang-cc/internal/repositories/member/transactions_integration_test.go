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
)

func TestTransactionLifecycleRestoresCapAndIsolatesUsers(t *testing.T) {
	pool, gormDB := integrationPool(t)
	service := NewTransactionRepository(pool)
	ctx := context.Background()

	june := WriteInput{
		CardID: cardA, AmountMinor: 10000, CategoryID: "dining",
		MerchantID: "px_mart", MerchantName: "全聯福利中心", PaymentMethodID: "physical_card", TransactionDate: recommendations.MustLocalDate("2026-06-10"), Note: "June",
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
	if _, err := pool.Exec(ctx, `DELETE FROM card_activities WHERE id = $1`, activityA); err == nil {
		t.Fatal("expected rule used by allocations to be undeletable")
	}
}

func TestConcurrentCreatesDoNotExceedCap(t *testing.T) {
	pool, _ := integrationPool(t)
	service := NewTransactionRepository(pool)
	input := WriteInput{
		CardID: cardA, AmountMinor: 10000, CategoryID: "dining",
		MerchantID: "px_mart", MerchantName: "全聯福利中心", PaymentMethodID: "physical_card", TransactionDate: recommendations.MustLocalDate("2026-06-10"),
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

func TestNormalizedCatalogConstraints(t *testing.T) {
	pool, gormDB := integrationPool(t)
	ctx := context.Background()
	var cards, benefits int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM card_products WHERE id::text LIKE '51000000-%'`).Scan(&cards); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM card_activity_benefits WHERE card_activity_id::text LIKE '61000000-%'`).Scan(&benefits); err != nil {
		t.Fatal(err)
	}
	if cards != 10 || benefits < 23 {
		t.Fatalf("core catalog cards=%d benefits=%d", cards, benefits)
	}

	if _, err := pool.Exec(ctx, `INSERT INTO card_activities(id,card_product_id,name,start_date,end_date) VALUES
		('62000000-0000-0000-0000-000000000001','51000000-0000-0000-0000-000000000001','重疊活動','2026-06-01','2026-07-31')`); err == nil {
		t.Fatal("expected overlapping activity to be rejected")
	}

	if err := New(pool, gormDB).CreateCard(ctx, "22000000-0000-0000-0000-000000000001", string(userA), "51000000-0000-0000-0000-000000000001", CardWrite{}); err == nil {
		t.Fatal("expected DAWHO account tier to be required")
	}
}

func TestRewardVersionAndQualificationHistoryConstraints(t *testing.T) {
	pool, gormDB := integrationPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO reward.component_versions(
		id,reward_component_id,reward_unit_id,name,rate,effective_from,effective_to)
		SELECT '72000000-0000-0000-0000-000000000001',reward_component_id,reward_unit_id,'重疊版本',rate,
		       effective_from,effective_to
		FROM reward.component_versions
		WHERE reward_component_id='71000000-0000-0000-0000-000000000001'
		LIMIT 1`); err == nil {
		t.Fatal("expected overlapping reward component version to be rejected")
	}

	repo := New(pool, gormDB)
	cardID := "25000000-0000-0000-0000-000000000001"
	if err := repo.CreateCard(ctx, cardID, string(userA), "51000000-0000-0000-0000-000000000001", CardWrite{
		IsActive: true, AccountTier: "大戶",
	}); err != nil {
		t.Fatal(err)
	}
	overview, err := repo.RewardOverview(ctx, string(userA), cardID, time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.RewardGroups) == 0 || len(overview.QualifiedPlans) == 0 {
		t.Fatalf("overview groups=%d qualified plans=%d", len(overview.RewardGroups), len(overview.QualifiedPlans))
	}
	var planID string
	if err := pool.QueryRow(ctx, `SELECT p.id FROM catalog.card_plans p
		JOIN catalog.card_plan_versions pv ON pv.card_plan_id=p.id
		WHERE p.card_product_id='51000000-0000-0000-0000-000000000001' AND pv.name='大戶'`).Scan(&planID); err != nil {
		t.Fatal(err)
	}
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
	repo := New(pool, gormDB)
	cards := []struct{ id, product, tier string }{
		{"23000000-0000-0000-0000-000000000001", "51000000-0000-0000-0000-000000000001", "大戶Plus"},
		{"23000000-0000-0000-0000-000000000002", "51000000-0000-0000-0000-000000000002", ""},
		{"23000000-0000-0000-0000-000000000003", "51000000-0000-0000-0000-000000000003", ""},
		{"23000000-0000-0000-0000-000000000004", "51000000-0000-0000-0000-000000000004", ""},
	}
	for _, c := range cards {
		if err := repo.CreateCard(ctx, c.id, string(userA), c.product, CardWrite{IsActive: true, AccountTier: c.tier}); err != nil {
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
			if err := repo.CreateCard(ctx, cardID, string(userA), "51000000-0000-0000-0000-000000000001", CardWrite{IsActive: true, AccountTier: tc.tier}); err != nil {
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
	repo := New(pool, gormDB)
	for _, card := range []struct {
		id      string
		product string
		tier    string
	}{
		{"25000000-0000-0000-0000-000000000001", "51000000-0000-0000-0000-000000000001", "大大"},
		{"25000000-0000-0000-0000-000000000002", "51000000-0000-0000-0000-000000000007", ""},
	} {
		if err := repo.CreateCard(ctx, card.id, string(userA), card.product, CardWrite{Nickname: "LOL", IsActive: true, AccountTier: card.tier}); err != nil {
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
	for _, card := range []struct{ id, product string }{
		{"91000000-0000-0000-0000-000000000005", "51000000-0000-0000-0000-000000000005"},
		{"91000000-0000-0000-0000-000000000006", "51000000-0000-0000-0000-000000000006"},
		{"91000000-0000-0000-0000-000000000007", "51000000-0000-0000-0000-000000000007"},
		{"91000000-0000-0000-0000-000000000008", "51000000-0000-0000-0000-000000000008"},
		{"91000000-0000-0000-0000-000000000009", "51000000-0000-0000-0000-000000000009"},
		{"91000000-0000-0000-0000-000000000010", "51000000-0000-0000-0000-000000000010"},
	} {
		if _, err := pool.Exec(ctx, `INSERT INTO member_cards(id,user_id,card_product_id,nickname,is_active) VALUES($1,$2,$3,'',true)`, card.id, userA, card.product); err != nil {
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
		{`INSERT INTO card_products(id,bank_id,name) VALUES
			('50000000-0000-0000-0000-000000000001','40000000-0000-0000-0000-000000000001','森活卡'),
			('50000000-0000-0000-0000-000000000002','40000000-0000-0000-0000-000000000001','青雲卡')`, nil},
		{`INSERT INTO member_cards(id,user_id,card_product_id,nickname,is_active) VALUES
			($3,$1,'50000000-0000-0000-0000-000000000001','森活卡',true),
			($4,$2,'50000000-0000-0000-0000-000000000002','青雲卡',true)`,
			[]any{userA, userB, cardA, cardB}},
		{`
		INSERT INTO reward_preferences (user_id, reward_unit_id, weight)
			VALUES ($1, $2, 1)`,
			[]any{userA, cashID}},
		{`
		INSERT INTO card_activities
			(id, card_product_id, name, start_date, end_date, is_active)
			VALUES ($1, '50000000-0000-0000-0000-000000000001', '測試活動', '2026-01-01', '2026-12-31', true)`,
			[]any{activityA}},
		{`
		INSERT INTO card_activity_benefits(id,card_activity_id,reward_unit_id,name,rate,monthly_cap,stack_group)
			VALUES ($1,$2,$3,'餐飲 10%',0.1,10,'base')`, []any{ruleA, activityA, cashID}},
		{`
		INSERT INTO card_activity_benefit_categories (benefit_id, category_id)
			VALUES ($1, 'dining')`,
			[]any{ruleA}},
		{`INSERT INTO merchants(id,name,is_active,is_system) VALUES('70000000-0000-0000-0000-000000000001','全聯',true,false)`, nil},
		{`INSERT INTO merchant_aliases(merchant_id,alias) VALUES('70000000-0000-0000-0000-000000000001','全聯福利中心')`, nil},
		{`INSERT INTO card_activity_benefit_merchants (benefit_id, merchant_id)
			VALUES ($1, '70000000-0000-0000-0000-000000000001')`, []any{ruleA}},
	}
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatalf("seed integration data: %v", err)
		}
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
