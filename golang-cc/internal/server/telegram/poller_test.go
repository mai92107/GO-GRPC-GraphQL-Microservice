package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	telegramcontroller "github.com/rafa/golang-cc/internal/controllers/telegram"
)

func TestSendKeyboardAndAnswerCallback(t *testing.T) {
	var mu sync.Mutex
	requests := map[string]map[string]any{}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		mu.Lock()
		requests[r.URL.Path] = payload
		mu.Unlock()
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	})}

	poller := NewPoller("test-token", nil)
	poller.client = client
	poller.sendResponse(context.Background(), 123, telegramcontroller.Response{
		Text: "請選擇消費分類：",
		Keyboard: [][]telegramcontroller.Button{{
			{Text: "餐飲", Data: "category:dining"},
			{Text: "旅遊", Data: "category:travel"},
		}},
	})
	if err := poller.answerCallback(context.Background(), "callback-id"); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	send := requests["/bottest-token/sendMessage"]
	if send["chat_id"] != float64(123) || send["reply_markup"] == nil {
		t.Fatalf("unexpected sendMessage payload: %#v", send)
	}
	answer := requests["/bottest-token/answerCallbackQuery"]
	if answer["callback_query_id"] != "callback-id" {
		t.Fatalf("unexpected answerCallbackQuery payload: %#v", answer)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
