package telegram

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
	memberservice "github.com/rafa/golang-cc/internal/services/member"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

const (
	stateTTL               = 10 * time.Minute
	categoryCallbackPrefix = "category:"
	merchantCallbackPrefix = "merchant:"
	merchantPagePrefix     = "merchant_page:"
	otherMerchantCallback  = "merchant:other"
	maxMessageSize         = 4096
	merchantPageSize       = 8
)

var amountPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]{1,2})?$`)
var taipeiLocation = time.FixedZone("Asia/Taipei", 8*60*60)

type BindingRepository interface {
	UserIDByChat(context.Context, int64) (string, error)
}

type MemberService interface {
	Categories(context.Context) ([]domain.MemberCategory, error)
	Merchants(context.Context, string) ([]domain.MemberMerchant, error)
	Recommend(context.Context, string, memberservice.RecommendationInput) (recommendations.Result, error)
}

type Button struct {
	Text string
	Data string
}

type Response struct {
	Text     string
	Keyboard [][]Button
}

type conversation struct {
	Step         string
	CategoryID   string
	CategoryName string
	AmountMinor  int64
	MerchantID   string
	MerchantName string
	UpdatedAt    time.Time
}

type Controller struct {
	bindings BindingRepository
	members  MemberService
	now      func() time.Time
	mu       sync.Mutex
	states   map[int64]conversation
}

func New(bindings BindingRepository, members MemberService) *Controller {
	return NewWithClock(bindings, members, time.Now)
}

func NewWithClock(bindings BindingRepository, members MemberService, now func() time.Time) *Controller {
	return &Controller{bindings: bindings, members: members, now: now, states: map[int64]conversation{}}
}

func (c *Controller) HandleMessage(ctx context.Context, chatID int64, text string) Response {
	text = strings.TrimSpace(text)
	switch text {
	case "/help", "/start":
		return Response{Text: "輸入 /recommand 開始信用卡推薦，或輸入 /cancel 取消目前流程。"}
	case "/cancel":
		c.clear(chatID)
		return Response{Text: "已取消目前推薦流程。"}
	case "/skip":
		state, ok := c.state(chatID)
		if ok && state.Step == "other_merchant" {
			return c.recommend(ctx, chatID, state, "")
		}
		return Response{Text: "目前沒有可略過的步驟。"}
	case "/recommand":
		return c.start(ctx, chatID)
	}

	state, ok := c.state(chatID)
	if !ok {
		return Response{Text: "請輸入 /recommand 開始信用卡推薦。"}
	}
	switch state.Step {
	case "category":
		return Response{Text: "請點選上方的消費分類，或輸入 /cancel 取消。"}
	case "merchant_choice":
		return Response{Text: "請點選上方的店家，或輸入 /cancel 取消。"}
	case "other_merchant":
		return c.recommend(ctx, chatID, state, text)
	case "amount":
		amountMinor, ok := parseAmountMinor(text)
		if !ok {
			return Response{Text: "金額格式錯誤，請輸入大於零且最多兩位小數的金額，例如：1000"}
		}
		state.Step = "merchant_choice"
		state.AmountMinor = amountMinor
		state.UpdatedAt = c.now()
		c.set(chatID, state)
		return c.merchantKeyboard(ctx, chatID, state)
	default:
		c.clear(chatID)
		return Response{Text: "推薦流程已失效，請重新輸入 /recommand。"}
	}
}

func (c *Controller) HandleCallback(ctx context.Context, chatID int64, data string) Response {
	if _, err := c.bindings.UserIDByChat(ctx, chatID); err != nil {
		c.clear(chatID)
		return Response{Text: unboundMessage(chatID)}
	}
	state, ok := c.state(chatID)
	if !ok {
		return Response{Text: "選單已失效，請重新輸入 /recommand。"}
	}
	if pageText, ok := strings.CutPrefix(data, merchantPagePrefix); ok {
		page, err := strconv.Atoi(pageText)
		if err != nil || state.Step != "merchant_choice" {
			return Response{Text: "店家選單已失效，請重新輸入 /recommand。"}
		}
		return c.merchantKeyboardPage(ctx, chatID, state, page)
	}
	if code, ok := strings.CutPrefix(data, merchantCallbackPrefix); ok {
		if state.Step != "merchant_choice" {
			return Response{Text: "店家選單已失效，請重新輸入 /recommand。"}
		}
		if data == otherMerchantCallback {
			state.Step, state.MerchantID, state.MerchantName, state.UpdatedAt = "other_merchant", "", "", c.now()
			c.set(chatID, state)
			return Response{Text: "可輸入其他店家名稱，或輸入 /skip 略過。"}
		}
		merchants, err := c.members.Merchants(ctx, state.CategoryID)
		if err != nil {
			return Response{Text: "目前無法讀取店家，請稍後再試。"}
		}
		for _, merchant := range merchants {
			if merchant.ID == code {
				state.MerchantID, state.MerchantName = merchant.ID, merchant.Name
				return c.recommend(ctx, chatID, state, merchant.Name)
			}
		}
		return Response{Text: "此店家已停用，請重新輸入 /recommand。"}
	}
	if state.Step != "category" {
		return Response{Text: "分類選單已失效，請重新輸入 /recommand。"}
	}
	code, ok := strings.CutPrefix(data, categoryCallbackPrefix)
	if !ok || code == "" {
		return Response{Text: "分類選擇無效，請重新輸入 /recommand。"}
	}
	categories, err := c.members.Categories(ctx)
	if err != nil {
		return Response{Text: "目前無法讀取分類，請稍後再試。"}
	}
	for _, category := range categories {
		if category.ID == code {
			state.Step, state.CategoryID, state.CategoryName, state.UpdatedAt = "amount", category.ID, category.Name, c.now()
			c.set(chatID, state)
			return Response{Text: fmt.Sprintf("已選擇「%s」。請輸入消費金額，例如：1000", category.Name)}
		}
	}
	c.clear(chatID)
	return Response{Text: "此分類已停用，請重新輸入 /recommand。"}
}

func (c *Controller) start(ctx context.Context, chatID int64) Response {
	c.clear(chatID)
	if _, err := c.bindings.UserIDByChat(ctx, chatID); err != nil {
		return Response{Text: unboundMessage(chatID)}
	}
	categories, err := c.members.Categories(ctx)
	if err != nil {
		return Response{Text: "目前無法讀取消費分類，請稍後再試。"}
	}
	keyboard := make([][]Button, 0, (len(categories)+1)/2)
	for _, category := range categories {
		button := Button{Text: category.Name, Data: categoryCallbackPrefix + category.ID}
		if len(keyboard) == 0 || len(keyboard[len(keyboard)-1]) == 2 {
			keyboard = append(keyboard, []Button{button})
		} else {
			keyboard[len(keyboard)-1] = append(keyboard[len(keyboard)-1], button)
		}
	}
	if len(keyboard) == 0 {
		return Response{Text: "目前沒有可使用的消費分類。"}
	}
	c.set(chatID, conversation{Step: "category", UpdatedAt: c.now()})
	return Response{Text: "請選擇消費分類：", Keyboard: keyboard}
}

func (c *Controller) merchantKeyboard(ctx context.Context, chatID int64, state conversation) Response {
	return c.merchantKeyboardPage(ctx, chatID, state, 0)
}

func (c *Controller) merchantKeyboardPage(ctx context.Context, chatID int64, state conversation, page int) Response {
	merchants, err := c.members.Merchants(ctx, state.CategoryID)
	if err != nil {
		return Response{Text: "目前無法讀取店家，請稍後再試。"}
	}
	if page < 0 {
		page = 0
	}
	start := page * merchantPageSize
	if start >= len(merchants) && len(merchants) > 0 {
		start, page = 0, 0
	}
	end := min(start+merchantPageSize, len(merchants))
	keyboard := make([][]Button, 0, merchantPageSize/2+2)
	for _, merchant := range merchants[start:end] {
		button := Button{Text: merchant.Name, Data: merchantCallbackPrefix + merchant.ID}
		if len(keyboard) == 0 || len(keyboard[len(keyboard)-1]) == 2 {
			keyboard = append(keyboard, []Button{button})
		} else {
			keyboard[len(keyboard)-1] = append(keyboard[len(keyboard)-1], button)
		}
	}
	navigation := []Button{}
	if page > 0 {
		navigation = append(navigation, Button{Text: "上一頁", Data: merchantPagePrefix + strconv.Itoa(page-1)})
	}
	if end < len(merchants) {
		navigation = append(navigation, Button{Text: "下一頁", Data: merchantPagePrefix + strconv.Itoa(page+1)})
	}
	if len(navigation) > 0 {
		keyboard = append(keyboard, navigation)
	}
	keyboard = append(keyboard, []Button{{Text: "其他", Data: otherMerchantCallback}})
	state.Step, state.UpdatedAt = "merchant_choice", c.now()
	c.set(chatID, state)
	return Response{Text: "請選擇店家：", Keyboard: keyboard}
}

func (c *Controller) recommend(ctx context.Context, chatID int64, state conversation, merchant string) Response {
	defer c.clear(chatID)
	userID, err := c.bindings.UserIDByChat(ctx, chatID)
	if err != nil {
		return Response{Text: unboundMessage(chatID)}
	}
	now := c.now().In(taipeiLocation)
	date, _ := recommendations.ParseLocalDate(now.Format("2006-01-02"))
	result, err := c.members.Recommend(ctx, userID, memberservice.RecommendationInput{
		AmountMinor: state.AmountMinor, CategoryID: state.CategoryID, MerchantID: state.MerchantID, MerchantName: merchant, Date: date,
	})
	if err != nil {
		return Response{Text: "推薦計算失敗，請稍後重新輸入 /recommand。"}
	}
	if len(result.Recommendations) == 0 {
		reason := result.EmptyReason
		if reason == "" {
			reason = "沒有符合條件的信用卡"
		}
		return Response{Text: fmt.Sprintf("找不到推薦結果。\n原因：%s", reason)}
	}
	return Response{Text: formatRecommendation(state, merchant, date, result.Recommendations[0])}
}

func formatRecommendation(state conversation, merchant string, date recommendations.LocalDate, item recommendations.CardRecommendation) string {
	lines := []string{
		"推薦第一名：" + item.CardName,
		fmt.Sprintf("消費：NT$%s｜%s｜%s｜%s", formatAmount(state.AmountMinor), state.CategoryName, merchant, date.String()),
	}
	for _, option := range item.PaymentOptions {
		rate := recommendations.SumRates(option).Mul(recommendations.MustDecimal("100"))
		lines = append(lines, "", "支付方式："+option.PaymentMethodName, "預估回饋("+recommendations.FormatDecimal(rate, 6)+"%)：")
		for _, allocation := range option.Allocations {
			line := fmt.Sprintf("- %s／%s：%s", allocation.ActivityName, allocation.RuleName,
				recommendations.FormatReward(allocation.AllocatedReward, allocation.RewardUnit))
			if allocation.RemainingBefore != nil {
				line += fmt.Sprintf("（本回饋額度 剩餘 %s）",
					recommendations.FormatReward(*allocation.RemainingBefore, allocation.RewardUnit))
			}
			lines = append(lines, line)
		}
		if len(option.Reminders) > 0 {
			lines = append(lines, "提醒：")
			for _, reminder := range option.Reminders {
				lines = append(lines, "- "+reminder)
			}
		}
	}
	return truncate(strings.Join(lines, "\n"))
}

func parseAmountMinor(value string) (int64, bool) {
	if !amountPattern.MatchString(value) {
		return 0, false
	}
	parts := strings.SplitN(value, ".", 2)
	major, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}
	minor := int64(0)
	if len(parts) == 2 {
		fraction := parts[1]
		if len(fraction) == 1 {
			fraction += "0"
		}
		minor, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, false
		}
	}
	if major > (1<<63-1-minor)/100 {
		return 0, false
	}
	result := major*100 + minor
	return result, result > 0
}

func formatAmount(amountMinor int64) string {
	if amountMinor%100 == 0 {
		return strconv.FormatInt(amountMinor/100, 10)
	}
	return fmt.Sprintf("%d.%02d", amountMinor/100, amountMinor%100)
}

func truncate(value string) string {
	runes := []rune(value)
	const suffix = "\n\n訊息過長，部分內容已省略。"
	if len(runes) <= maxMessageSize {
		return value
	}
	limit := maxMessageSize - len([]rune(suffix))
	return string(runes[:limit]) + suffix
}

func unboundMessage(chatID int64) string {
	return fmt.Sprintf("此 Telegram Chat 尚未綁定會員。\nchat_id：%d\n請將 chat_id 提供給管理員。", chatID)
}

func (c *Controller) state(chatID int64) (conversation, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, ok := c.states[chatID]
	if ok && c.now().Sub(state.UpdatedAt) > stateTTL {
		delete(c.states, chatID)
		return conversation{}, false
	}
	return state, ok
}

func (c *Controller) set(chatID int64, state conversation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.states[chatID] = state
}

func (c *Controller) clear(chatID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.states, chatID)
}
