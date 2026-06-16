package telegram

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
	memberservice "github.com/rafa/golang-cc/internal/services/member"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

type fakeBindings struct {
	users map[int64]string
}

func (f fakeBindings) UserIDByChat(_ context.Context, chatID int64) (string, error) {
	userID := f.users[chatID]
	if userID == "" {
		return "", domain.ErrNotFound
	}
	return userID, nil
}

type fakeMembers struct {
	categories []domain.MemberCategory
	merchants  []domain.MemberMerchant
	result     recommendations.Result
	err        error
	input      memberservice.RecommendationInput
	userID     string
}

func (f *fakeMembers) Categories(context.Context) ([]domain.MemberCategory, error) {
	return f.categories, f.err
}
func (f *fakeMembers) Merchants(context.Context, string) ([]domain.MemberMerchant, error) {
	return f.merchants, f.err
}

func (f *fakeMembers) Recommend(_ context.Context, userID string, input memberservice.RecommendationInput) (recommendations.Result, error) {
	f.userID, f.input = userID, input
	return f.result, f.err
}

func TestInteractiveRecommendationFlow(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, time.FixedZone("test", 8*60*60))
	members := &fakeMembers{
		categories: []domain.MemberCategory{{Code: "dining", Name: "餐飲"}, {Code: "grocery", Name: "量販超市"}, {Code: "travel", Name: "旅遊"}},
		merchants:  []domain.MemberMerchant{{Code: "px_mart", Name: "全聯"}},
		result: recommendations.Result{Recommendations: []recommendations.CardRecommendation{{
			CardName:   "最佳卡",
			TotalScore: recommendations.MustDecimal("100"),
			Allocations: []recommendations.RuleEvaluation{{
				ActivityName: "本期活動", RuleName: "量販回饋", RewardUnit: recommendations.RewardUnit{Symbol: "NT$", Precision: 2},
				AllocatedReward: recommendations.MustDecimal("100"),
			}},
			PaymentOptions: []recommendations.PaymentOption{{
				PaymentMethodCode: "line_pay", PaymentMethodName: "LINE Pay",
				Allocations: []recommendations.RuleEvaluation{{ActivityName: "本期活動", RuleName: "量販回饋", RewardUnit: recommendations.RewardUnit{Symbol: "NT$", Precision: 2}, AllocatedReward: recommendations.MustDecimal("100")}},
				Reminders:   []string{"需先登錄"},
			}},
		}}},
	}
	controller := NewWithClock(fakeBindings{users: map[int64]string{123: "member-1"}}, members, func() time.Time { return now })

	response := controller.HandleMessage(context.Background(), 123, "/recommand")
	if response.Text != "請選擇消費分類：" {
		t.Fatalf("unexpected category response: %+v", response)
	}
	response = controller.HandleCallback(context.Background(), 123, "category:grocery")
	if !strings.Contains(response.Text, "量販超市") {
		t.Fatalf("unexpected callback response: %q", response.Text)
	}
	response = controller.HandleMessage(context.Background(), 123, "1,000")
	if !strings.Contains(response.Text, "金額格式錯誤") {
		t.Fatalf("invalid amount response: %q", response.Text)
	}
	response = controller.HandleMessage(context.Background(), 123, "1000.50")
	if !strings.Contains(response.Text, "請選擇店家") {
		t.Fatalf("amount response: %q", response.Text)
	}
	response = controller.HandleCallback(context.Background(), 123, "merchant:px_mart")
	if !strings.Contains(response.Text, "推薦第一名：最佳卡") || !strings.Contains(response.Text, "需先登錄") {
		t.Fatalf("recommendation response: %q", response.Text)
	}
	if members.userID != "member-1" || members.input.AmountMinor != 100050 || members.input.CategoryCode != "grocery" || members.input.MerchantCode != "px_mart" {
		t.Fatalf("unexpected recommendation input: user=%s input=%+v", members.userID, members.input)
	}
}

func TestUnboundCancelAndExpiredConversation(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	members := &fakeMembers{categories: []domain.MemberCategory{{Code: "dining", Name: "餐飲"}}}
	controller := NewWithClock(fakeBindings{users: map[int64]string{1: "member"}}, members, func() time.Time { return now })

	if response := controller.HandleMessage(context.Background(), 2, "/recommand"); !strings.Contains(response.Text, "chat_id：2") {
		t.Fatalf("unbound response: %q", response.Text)
	}
	controller.HandleMessage(context.Background(), 1, "/recommand")
	controller.HandleCallback(context.Background(), 1, "category:dining")
	now = now.Add(11 * time.Minute)
	if response := controller.HandleMessage(context.Background(), 1, "100"); !strings.Contains(response.Text, "/recommand") {
		t.Fatalf("expired response: %q", response.Text)
	}
	if response := controller.HandleMessage(context.Background(), 1, "/cancel"); !strings.Contains(response.Text, "已取消") {
		t.Fatalf("cancel response: %q", response.Text)
	}
	if response := controller.HandleCallback(context.Background(), 1, "category:dining"); !strings.Contains(response.Text, "已失效") {
		t.Fatalf("stale callback response: %q", response.Text)
	}
}

func TestCategoryAndRecommendationFailures(t *testing.T) {
	members := &fakeMembers{categories: []domain.MemberCategory{{Code: "dining", Name: "餐飲"}}}
	controller := New(fakeBindings{users: map[int64]string{1: "member"}}, members)
	controller.HandleMessage(context.Background(), 1, "/recommand")
	if response := controller.HandleCallback(context.Background(), 1, "category:disabled"); !strings.Contains(response.Text, "已停用") {
		t.Fatalf("disabled category response: %q", response.Text)
	}

	members.err = errors.New("failed")
	if response := controller.HandleMessage(context.Background(), 1, "/recommand"); !strings.Contains(response.Text, "無法讀取") {
		t.Fatalf("category failure response: %q", response.Text)
	}
}
