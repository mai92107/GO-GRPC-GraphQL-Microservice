package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	telegramcontroller "github.com/rafa/golang-cc/internal/controllers/telegram"
)

type Poller struct {
	token      string
	baseURL    string
	client     *http.Client
	controller *telegramcontroller.Controller
}

func NewPoller(token string, controller *telegramcontroller.Controller) *Poller {
	return &Poller{
		token: token, baseURL: "https://api.telegram.org", controller: controller,
		client: &http.Client{Timeout: 35 * time.Second},
	}
}

func (p *Poller) Run(ctx context.Context) {
	log.Print("telegram recommendation polling started")
	offset := 0
	for ctx.Err() == nil {
		updates, err := p.getUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("telegram getUpdates failed: %v", err)
				sleep(ctx, 3*time.Second)
			}
			continue
		}
		for _, update := range updates {
			offset = update.UpdateID + 1
			if update.CallbackQuery.ID != "" {
				chatID := update.CallbackQuery.Message.Chat.ID
				if err := p.answerCallback(ctx, update.CallbackQuery.ID); err != nil {
					log.Printf("telegram answerCallbackQuery failed: %v", err)
				}
				response := p.controller.HandleCallback(ctx, chatID, update.CallbackQuery.Data)
				p.sendResponse(ctx, chatID, response)
				continue
			}
			if update.Message.Text != "" {
				response := p.controller.HandleMessage(ctx, update.Message.Chat.ID, update.Message.Text)
				p.sendResponse(ctx, update.Message.Chat.ID, response)
			}
		}
	}
}

func (p *Poller) sendResponse(ctx context.Context, chatID int64, response telegramcontroller.Response) {
	if response.Text == "" {
		return
	}
	payload := sendMessageRequest{ChatID: chatID, Text: response.Text}
	if len(response.Keyboard) > 0 {
		payload.ReplyMarkup = &replyMarkup{InlineKeyboard: make([][]inlineButton, 0, len(response.Keyboard))}
		for _, row := range response.Keyboard {
			buttons := make([]inlineButton, 0, len(row))
			for _, button := range row {
				buttons = append(buttons, inlineButton{Text: button.Text, CallbackData: button.Data})
			}
			payload.ReplyMarkup.InlineKeyboard = append(payload.ReplyMarkup.InlineKeyboard, buttons)
		}
	}
	if err := p.post(ctx, "sendMessage", payload); err != nil {
		log.Printf("telegram sendMessage failed: %v", err)
	}
}

func (p *Poller) answerCallback(ctx context.Context, callbackID string) error {
	return p.post(ctx, "answerCallbackQuery", map[string]string{"callback_query_id": callbackID})
}

func (p *Poller) getUpdates(ctx context.Context, offset int) ([]update, error) {
	values := url.Values{"timeout": []string{"25"}}
	if offset > 0 {
		values.Set("offset", strconv.Itoa(offset))
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL("getUpdates")+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("http %s", response.Status)
	}
	var result struct {
		OK     bool     `json:"ok"`
		Result []update `json:"result"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("response ok=false")
	}
	return result.Result, nil
}

func (p *Poller) post(ctx context.Context, method string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL(method), bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := p.client.Do(request)
	if err != nil {
		return fmt.Errorf("request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("http %d", response.StatusCode)
	}
	return nil
}

func (p *Poller) apiURL(method string) string {
	return p.baseURL + "/bot" + p.token + "/" + method
}

func sleep(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

type sendMessageRequest struct {
	ChatID      int64        `json:"chat_id"`
	Text        string       `json:"text"`
	ReplyMarkup *replyMarkup `json:"reply_markup,omitempty"`
}

type replyMarkup struct {
	InlineKeyboard [][]inlineButton `json:"inline_keyboard,omitempty"`
}

type inlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type update struct {
	UpdateID      int           `json:"update_id"`
	Message       message       `json:"message"`
	CallbackQuery callbackQuery `json:"callback_query"`
}

type callbackQuery struct {
	ID      string  `json:"id"`
	Data    string  `json:"data"`
	Message message `json:"message"`
}

type message struct {
	Text string `json:"text"`
	Chat chat   `json:"chat"`
}

type chat struct {
	ID int64 `json:"id"`
}
