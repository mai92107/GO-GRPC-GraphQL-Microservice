package bot

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
)

type TelegramPoller struct {
	botToken string
	chatID   string
	client   *http.Client
	handler  Handler
}

func NewTelegramPoller(botToken, chatID string, timeout time.Duration, handler Handler) *TelegramPoller {
	return &TelegramPoller{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: timeout},
		handler:  handler,
	}
}

func (poller *TelegramPoller) Run(ctx context.Context) {
	log.Print("telegram command polling started")
	offset := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := poller.getUpdates(ctx, offset)
		if err != nil {
			log.Printf("telegram getUpdates error=%q", err.Error())
			sleep(ctx, 3*time.Second)
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			if update.Message.Text == "" {
				continue
			}
			chatID := strconv.FormatInt(update.Message.Chat.ID, 10)
			if poller.chatID != "" && poller.chatID != chatID {
				log.Printf("ignore telegram command from unauthorized chat_id=%s", chatID)
				continue
			}
			response := poller.handler.HandleCommand(update.Message.Text)
			if err := poller.sendMessage(ctx, chatID, response); err != nil {
				log.Printf("telegram send command response error=%q", err.Error())
			}
		}
	}
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

func (poller *TelegramPoller) sendMessage(ctx context.Context, chatID, text string) error {
	payload := map[string]string{
		"chat_id": chatID,
		"text":    text,
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

func sleep(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

type telegramUpdate struct {
	UpdateID int             `json:"update_id"`
	Message  telegramMessage `json:"message"`
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
