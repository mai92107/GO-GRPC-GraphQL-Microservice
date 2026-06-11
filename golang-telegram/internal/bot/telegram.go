package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang-springboot-monitor-bot/internal/applog"
)

type TelegramPoller struct {
	botToken string
	chatID   string
	client   *http.Client
	handler  Handler
	logger   applog.ApplicationLogger
}

func NewTelegramPoller(botToken, chatID string, timeout time.Duration, handler Handler, logger applog.ApplicationLogger) *TelegramPoller {
	return &TelegramPoller{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: timeout},
		handler:  handler,
		logger:   logger,
	}
}

func (poller *TelegramPoller) Run(ctx context.Context) {
	if err := poller.registerCommands(ctx); err != nil {
		poller.logger.Error(ctx, "telegram_register_commands_failed", err)
	}
	poller.logger.Info(ctx, "telegram_polling_started")
	offset := 0
	failures := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := poller.getUpdates(ctx, offset)
		if err != nil {
			poller.logger.Error(ctx, "telegram_polling_failed", err)
			failures++
			delay := time.Duration(1<<min(failures, 5)) * time.Second
			sleep(ctx, delay)
			continue
		}
		failures = 0

		for _, update := range updates {
			offset = update.UpdateID + 1
			message := update.Message
			callback := ""
			if update.CallbackQuery.Data != "" {
				message = update.CallbackQuery.Message
				callback = update.CallbackQuery.Data
				if err := poller.answerCallbackQuery(ctx, update.CallbackQuery.ID); err != nil {
					poller.logger.Error(ctx, "telegram_answer_callback_failed", err)
				}
			}
			chatID := strconv.FormatInt(message.Chat.ID, 10)
			if poller.chatID != "" && poller.chatID != chatID {
				poller.logger.Warn(ctx, "telegram_unauthorized_update", applog.Field{Key: "chat_id", Value: chatID})
				continue
			}
			response := poller.handler.Handle(ctx, chatID, message.Text, callback)
			if err := poller.sendReply(ctx, chatID, response); err != nil {
				poller.logger.Error(ctx, "telegram_send_response_failed", err)
			}
		}
	}
}

func (poller *TelegramPoller) registerCommands(ctx context.Context) error {
	payload := map[string]any{
		"commands": []map[string]string{
			{"command": "menu", "description": "開啟管理選單"},
			{"command": "status", "description": "查看服務狀態"},
			{"command": "metric", "description": "分層查詢監控指標"},
			{"command": "alerts", "description": "查看目前告警"},
			{"command": "check", "description": "立即檢查全部服務"},
			{"command": "list_services", "description": "查看服務清單"},
			{"command": "trend", "description": "查詢指標趨勢圖"},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", poller.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := poller.client.Do(req)
	if err != nil {
		return fmt.Errorf("Telegram 指令選單註冊失敗：%w", sanitizeTelegramError(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Telegram 指令選單註冊失敗：HTTP %d", resp.StatusCode)
	}
	return nil
}

func (poller *TelegramPoller) getUpdates(ctx context.Context, offset int) ([]telegramUpdate, error) {
	values := url.Values{}
	values.Set("timeout", "25")
	if offset > 0 {
		values.Set("offset", strconv.Itoa(offset))
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?%s", poller.botToken, values.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := poller.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram getUpdates request failed: %w", sanitizeTelegramError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("getUpdates failed: http %d", resp.StatusCode)
	}

	var body struct {
		OK     bool             `json:"ok"`
		Result []telegramUpdate `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if !body.OK {
		return nil, fmt.Errorf("getUpdates response ok=false")
	}
	return body.Result, nil
}

func (poller *TelegramPoller) sendReply(ctx context.Context, chatID string, reply Reply) error {
	if reply.Silent {
		return nil
	}
	if len(reply.Photo) > 0 {
		return poller.sendPhoto(ctx, chatID, reply)
	}
	chunks := splitTelegramText(reply.Text, 4000)
	for i, chunk := range chunks {
		keyboard := reply.Keyboard
		if i != len(chunks)-1 {
			keyboard = nil
		}
		if err := poller.sendMessage(ctx, chatID, chunk, keyboard); err != nil {
			return err
		}
	}
	return nil
}

func (poller *TelegramPoller) answerCallbackQuery(ctx context.Context, callbackQueryID string) error {
	if callbackQueryID == "" {
		return nil
	}
	body, err := json.Marshal(map[string]string{"callback_query_id": callbackQueryID})
	if err != nil {
		return err
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", poller.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := poller.client.Do(req)
	if err != nil {
		return fmt.Errorf("Telegram 按鈕回應失敗：%w", sanitizeTelegramError(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Telegram 按鈕回應失敗：HTTP %d", resp.StatusCode)
	}
	return nil
}

func (poller *TelegramPoller) sendMessage(ctx context.Context, chatID, text string, keyboard [][]Button) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	if len(keyboard) > 0 {
		payload["reply_markup"] = map[string]any{"inline_keyboard": keyboard}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", poller.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := poller.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram sendMessage request failed: %w", sanitizeTelegramError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("sendMessage failed: http %d", resp.StatusCode)
	}
	return nil
}

func splitTelegramText(text string, maxRunes int) []string {
	if maxRunes <= 0 {
		return []string{text}
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return []string{text}
	}
	var chunks []string
	for len(runes) > 0 {
		end := maxRunes
		if len(runes) < end {
			end = len(runes)
		}
		if end < len(runes) {
			for i := end; i > 0; i-- {
				if runes[i-1] == '\n' {
					end = i
					break
				}
			}
		}
		chunks = append(chunks, string(runes[:end]))
		runes = runes[end:]
	}
	return chunks
}

func (poller *TelegramPoller) sendPhoto(ctx context.Context, chatID string, reply Reply) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("chat_id", chatID)
	_ = writer.WriteField("caption", reply.Text)
	part, err := writer.CreateFormFile("photo", "trend.png")
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, bytes.NewReader(reply.Photo)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", poller.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := poller.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram sendPhoto request failed: %w", sanitizeTelegramError(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("sendPhoto failed: http %d", resp.StatusCode)
	}
	return nil
}

func sleep(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

type telegramUpdate struct {
	UpdateID      int                   `json:"update_id"`
	Message       telegramMessage       `json:"message"`
	CallbackQuery telegramCallbackQuery `json:"callback_query"`
}

type telegramCallbackQuery struct {
	ID      string          `json:"id"`
	Data    string          `json:"data"`
	Message telegramMessage `json:"message"`
}

type telegramMessage struct {
	Text string       `json:"text"`
	Chat telegramChat `json:"chat"`
}

type telegramChat struct {
	ID int64 `json:"id"`
}

type sanitizedTelegramError struct {
	message string
}

func (err sanitizedTelegramError) Error() string {
	return err.message
}

func sanitizeTelegramError(err error) error {
	if err == nil {
		return nil
	}
	return sanitizedTelegramError{message: "request failed"}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
