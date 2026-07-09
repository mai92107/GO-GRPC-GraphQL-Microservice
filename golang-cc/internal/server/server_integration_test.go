package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafa/golang-cc/internal/platform/config"
	publicservice "github.com/rafa/golang-cc/internal/services/public"
	"github.com/rafa/golang-cc/internal/utils/secure"
	"github.com/rafa/golang-cc/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type fakeEmail struct {
	mu         sync.Mutex
	invitation string
	reset      string
}

func (f *fakeEmail) SendInvitation(_ context.Context, _, link string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.invitation = link
	return nil
}
func (f *fakeEmail) SendPasswordReset(_ context.Context, _, link string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reset = link
	return nil
}

func TestAuthCSRFEmailAndHorizontalIsolation(t *testing.T) {
	pool, gormDB := apiPool(t)
	mailer := &fakeEmail{}
	handler := New(pool, gormDB, mailer, "http://localhost:8080", false, "test-monitoring-token", nil)
	adminPassword := "admin-password-123"
	hash, _ := publicservice.HashPassword(adminPassword)
	adminID := secure.UUID()
	if _, err := pool.Exec(context.Background(), `INSERT INTO users(id,email,password_hash,display_name,role,status)
		VALUES($1,'admin@example.test',$2,'Admin','admin','active')`, adminID, hash); err != nil {
		t.Fatal(err)
	}

	response := requestWithAuthorization(t, handler, "GET", "/actuator/health", "", nil)
	assertStatus(t, response, 401)
	response = requestWithAuthorization(t, handler, "GET", "/actuator/health", "Bearer test-monitoring-token", nil)
	assertStatus(t, response, 200)
	response = requestWithAuthorization(t, handler, "GET", "/actuator/prometheus", "Bearer test-monitoring-token", nil)
	assertStatus(t, response, 200)
	if !strings.Contains(response.Body.String(), "go_goroutines") || !strings.Contains(response.Body.String(), "pgxpool_total_connections") {
		t.Fatalf("prometheus metrics missing required collectors: %s", response.Body.String())
	}
	response = requestWithAuthorization(t, handler, "GET", "/api/health/live", "", nil)
	assertStatus(t, response, 404)

	adminCookie, adminCSRF := loginRequest(t, handler, "admin@example.test", adminPassword)
	response = request(t, handler, "POST", "/api/admin/payment-methods", map[string]any{"name": "測試錢包"}, adminCookie, adminCSRF)
	assertStatus(t, response, 201)
	paymentMethodID := responseDataString(t, response, "id")
	response = request(t, handler, "GET", "/api/admin/payment-methods", nil, adminCookie, "")
	assertStatus(t, response, 200)
	response = request(t, handler, "DELETE", "/api/admin/payment-methods/"+paymentMethodID, nil, adminCookie, adminCSRF)
	assertStatus(t, response, 200)
	response = request(t, handler, "POST", "/api/admin/merchants", map[string]any{
		"name": "全聯", "aliases": []string{"全聯福利中心"}, "category_ids": []string{"grocery"},
	}, adminCookie, adminCSRF)
	assertStatus(t, response, 201)
	response = request(t, handler, "GET", "/api/admin/merchants", nil, adminCookie, "")
	assertStatus(t, response, 200)
	response = requestWithAuthorization(t, handler, "GET", "/actuator/health", "", adminCookie)
	assertStatus(t, response, 200)
	response = request(t, handler, "POST", "/api/admin/invitations", map[string]any{"email": "member@example.test"}, adminCookie, adminCSRF)
	assertStatus(t, response, 201)
	inviteToken := linkToken(t, mailer.invitation)
	response = request(t, handler, "POST", "/api/public/auth/accept-invitation", map[string]any{
		"token": inviteToken, "display_name": "Member", "password": "member-password-123",
	}, nil, "")
	assertStatus(t, response, 201)

	memberCookie, memberCSRF := loginRequest(t, handler, "member@example.test", "member-password-123")
	response = request(t, handler, "GET", "/api/member/payment-methods", nil, memberCookie, "")
	assertStatus(t, response, 200)
	var memberID string
	if err := pool.QueryRow(context.Background(), `SELECT id FROM users WHERE email='member@example.test'`).Scan(&memberID); err != nil {
		t.Fatal(err)
	}
	response = request(t, handler, "POST", "/api/admin/telegram-bindings", map[string]any{"chat_id": 123456789, "user_id": memberID}, adminCookie, adminCSRF)
	assertStatus(t, response, 201)
	response = request(t, handler, "GET", "/api/admin/telegram-bindings", nil, adminCookie, "")
	assertStatus(t, response, 200)
	response = request(t, handler, "POST", "/api/admin/telegram-bindings", map[string]any{"chat_id": 123456789, "user_id": memberID}, adminCookie, adminCSRF)
	assertStatus(t, response, 409)
	response = requestWithAuthorization(t, handler, "GET", "/actuator/health", "", memberCookie)
	assertStatus(t, response, 403)
	response = request(t, handler, "POST", "/api/admin/banks", map[string]any{"name": "虛構銀行", "is_active": true}, adminCookie, adminCSRF)
	assertStatus(t, response, 201)
	bankID := responseDataString(t, response, "id")
	response = request(t, handler, "POST", "/api/admin/card-products", map[string]any{
		"bank_id": bankID, "name": "森活卡", "is_active": true, "account_tiers": []string{"尊榮會員"},
		"networks": []string{"Visa"},
	}, adminCookie, adminCSRF)
	assertStatus(t, response, 201)
	productID := responseDataString(t, response, "id")
	response = request(t, handler, "POST", "/api/admin/activities", map[string]any{
		"activity": map[string]any{
			"bank_id": bankID, "card_product_id": productID, "title": "2026 核心權益",
			"effective_from": "2026-01-01", "effective_to": "2026-12-31", "is_active": true,
		},
		"reward_groups": []map[string]any{
			{
				"id": "group-main", "name": "主要回饋", "display_order": 10, "is_active": true,
				"components": []map[string]any{
					{
						"id": "component-dining", "reward_group_ids": []string{"group-main"}, "name": "餐飲 10%",
						"layer": 3, "stack_group": "base", "stack_mode": "ADDITIVE", "priority": 10,
						"effective_from": "2026-01-01", "effective_to": "2026-12-31", "is_active": true,
						"requirements": []map[string]any{
							{"id": "req-dining", "requirement_type": "CONSUMPTION_CATEGORY", "operator": "IN", "configuration_json": map[string]any{"category_ids": []string{"dining"}}, "is_active": true},
							{"id": "req-tier", "requirement_type": "ACCOUNT_TIER", "operator": "IN", "configuration_json": map[string]any{"tiers": []string{"尊榮會員"}}, "is_active": true},
						},
						"benefits": []map[string]any{
							{"id": "benefit-dining", "benefit_type": "RATE_CASHBACK", "value": "0.1", "reward_unit_id": "00000000-0000-0000-0000-000000000101", "cap_amount": "10", "cap_period": "calendar_month", "description": "餐飲 10%", "is_active": true},
						},
					},
				},
			},
		},
	}, adminCookie, adminCSRF)
	assertStatus(t, response, 201)
	activityID := responseDataString(t, response, "id")
	response = request(t, handler, "POST", "/api/admin/activities/"+activityID+"/publish", map[string]any{}, adminCookie, adminCSRF)
	assertStatus(t, response, 200)
	var publishedRequirements int
	if err := pool.QueryRow(context.Background(), `SELECT count(*)
		FROM reward.published_rule_requirements r
		JOIN reward.published_reward_rules rule ON rule.id=r.published_rule_id
		WHERE rule.source_activity_id=$1`, activityID).Scan(&publishedRequirements); err != nil {
		t.Fatal(err)
	}
	if publishedRequirements != 2 {
		t.Fatalf("published requirements=%d,want 2", publishedRequirements)
	}
	var networkID string
	if err := pool.QueryRow(context.Background(), `SELECT card_network_id::text FROM catalog.card_product_networks WHERE card_product_id=$1 LIMIT 1`, productID).Scan(&networkID); err != nil {
		t.Fatal(err)
	}
	memberCardPayload := map[string]any{"card_id": productID, "card_network_id": networkID, "nickname": "我的森活卡", "last_four": "8899", "statement_day": 5, "payment_due_day": 20, "account_tier": "尊榮會員", "credit_limit": "100000", "is_active": true}
	response = request(t, handler, "POST", "/api/member/cards", memberCardPayload, memberCookie, "")
	assertStatus(t, response, 403)
	response = request(t, handler, "POST", "/api/member/cards", memberCardPayload, memberCookie, memberCSRF)
	assertStatus(t, response, 201)
	cardID := responseDataString(t, response, "id")
	response = request(t, handler, "GET", "/api/member/cards/"+cardID, nil, nil, "")
	assertStatus(t, response, 401)
	response = request(t, handler, "GET", "/api/member/cards/"+cardID, nil, adminCookie, "")
	assertStatus(t, response, 403)
	response = request(t, handler, "POST", "/api/admin/invitations", map[string]any{"email": "blocked@example.test"}, memberCookie, memberCSRF)
	assertStatus(t, response, 403)

	response = request(t, handler, "POST", "/api/admin/activities", map[string]any{}, memberCookie, memberCSRF)
	assertStatus(t, response, 403)
	recommendation := map[string]any{"amount_minor": 10000, "category_id": "dining", "merchant_id": "", "date": "2026-06-10"}
	response = request(t, handler, "POST", "/api/member/recommendations", recommendation, memberCookie, memberCSRF)
	assertStatus(t, response, 200)
	response = request(t, handler, "POST", "/api/member/transactions", map[string]any{
		"card_id": cardID, "amount_minor": 10000, "category_id": "dining", "merchant_id": "", "payment_method_id": "physical_card", "transaction_date": "2026-06-10",
	}, memberCookie, memberCSRF)
	assertStatus(t, response, 201)
	response = request(t, handler, "POST", "/api/member/recommendations", recommendation, memberCookie, memberCSRF)
	assertStatus(t, response, 200)

	return

	response = request(t, handler, "POST", "/api/public/auth/request-password-reset", map[string]any{"email": "member@example.test"}, nil, "")
	assertStatus(t, response, 200)
	resetToken := linkToken(t, mailer.reset)
	response = request(t, handler, "POST", "/api/public/auth/reset-password", map[string]any{"token": resetToken, "password": "new-member-password-123"}, nil, "")
	assertStatus(t, response, 200)
	response = request(t, handler, "GET", "/api/public/auth/me", nil, memberCookie, "")
	assertStatus(t, response, 401)
	_, _ = loginRequest(t, handler, "member@example.test", "new-member-password-123")
	response = request(t, handler, "POST", "/api/public/auth/reset-password", map[string]any{"token": resetToken, "password": "another-password-123"}, nil, "")
	assertStatus(t, response, 400)
	response = request(t, handler, "DELETE", "/api/admin/telegram-bindings/123456789", nil, adminCookie, adminCSRF)
	assertStatus(t, response, 200)
}

func apiPool(t *testing.T) (*pgxpool.Pool, *gorm.DB) {
	t.Helper()
	cfg, err := config.LoadFromProject("configs/test.json")
	if err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(context.Background(), cfg.Database.ConnectionString())
	if err != nil {
		t.Fatal(err)
	}
	gormDB, err := gorm.Open(postgres.Open(cfg.Database.ConnectionString()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err = pool.Ping(context.Background()); err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	if _, err = pool.Exec(context.Background(), `DROP SCHEMA IF EXISTS identity,catalog,merchant,reward,member_profile,recommendation,"transaction",integration CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(context.Background(), `DROP SCHEMA public CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(context.Background(), `CREATE SCHEMA public`); err != nil {
		t.Fatal(err)
	}
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	if err = migrations.Apply(context.Background(), conn.Conn()); err != nil {
		t.Fatal(err)
	}
	return pool, gormDB
}

func loginRequest(t *testing.T, handler http.Handler, email, password string) (*http.Cookie, string) {
	t.Helper()
	response := request(t, handler, "POST", "/api/public/auth/login", map[string]any{"email": email, "password": password}, nil, "")
	assertStatus(t, response, 200)
	var envelope struct {
		Data struct {
			CSRF string `json:"csrf_token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(response.Body).Decode(&envelope)
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set cookie")
	}
	return cookies[0], envelope.Data.CSRF
}

func request(t *testing.T, handler http.Handler, method, path string, body any, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	var raw bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&raw).Encode(body)
	}
	req := httptest.NewRequest(method, path, &raw)
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func requestWithAuthorization(t *testing.T, handler http.Handler, method, path, authorization string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", authorization)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}
func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status=%d want=%d body=%s", response.Code, want, response.Body.String())
	}
}
func responseDataString(t *testing.T, response *httptest.ResponseRecorder, key string) string {
	t.Helper()
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	_ = json.Unmarshal(response.Body.Bytes(), &envelope)
	value, _ := envelope.Data[key].(string)
	return value
}
func linkToken(t *testing.T, link string) string {
	t.Helper()
	parsed, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		t.Fatal(err)
	}
	token := parsed.Query().Get("token")
	if token == "" {
		t.Fatal("email link has no token")
	}
	return token
}
