package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIBase = "https://api.telegram.org"
	// pollTimeoutSeconds is Telegram's long-poll hold time: the request stays
	// open until an update arrives, so idle costs nothing.
	pollTimeoutSeconds = 30
	// pollRetryDelay backs off after a failed poll so a broken network or an
	// expired token cannot spin the loop.
	pollRetryDelay = 5 * time.Second
)

// Bot is the Telegram runtime: it polls for updates and renders whatever the
// Service tells it to say. It holds no conversation state of its own.
type Bot struct {
	token   string
	service *Service
	limiter ChatLimiter
	client  *http.Client
	apiBase string
	// offset is Telegram's own cursor. Acknowledging an update by asking for
	// the next one is what makes redelivery impossible without a dedup table.
	offset int64
}

func NewBot(token string, service *Service, limiter ChatLimiter) *Bot {
	return &Bot{
		token:   token,
		service: service,
		limiter: limiter,
		apiBase: defaultAPIBase,
		// Comfortably longer than the long-poll hold time.
		client: &http.Client{Timeout: (pollTimeoutSeconds + 15) * time.Second},
	}
}

// update is the slice of Telegram's payload this bot cares about.
type update struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type updatesResponse struct {
	OK          bool     `json:"ok"`
	Description string   `json:"description"`
	Result      []update `json:"result"`
}

// Run polls until the context is cancelled.
func (b *Bot) Run(ctx context.Context) {
	slog.Info("telegram bot started")

	for {
		if err := ctx.Err(); err != nil {
			slog.Info("telegram bot stopped")
			return
		}

		if err := b.pollOnce(ctx); err != nil {
			if ctx.Err() != nil {
				slog.Info("telegram bot stopped")
				return
			}

			slog.Error("telegram poll failed", "error", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(pollRetryDelay):
			}
		}
	}
}

// pollOnce fetches a batch of updates and handles them in order.
func (b *Bot) pollOnce(ctx context.Context) error {
	updates, err := b.getUpdates(ctx)
	if err != nil {
		return err
	}

	for _, u := range updates {
		// Advance the cursor even for updates we ignore, otherwise Telegram
		// keeps redelivering them.
		if u.UpdateID >= b.offset {
			b.offset = u.UpdateID + 1
		}

		b.handleUpdate(ctx, u)
	}

	return nil
}

func (b *Bot) handleUpdate(ctx context.Context, u update) {
	if u.Message == nil || strings.TrimSpace(u.Message.Text) == "" {
		return
	}

	chatID := strconv.FormatInt(u.Message.Chat.ID, 10)

	if b.limiter != nil && !b.limiter.Allow(chatID) {
		// Dropped rather than answered: replying to a flood would only
		// amplify it.
		slog.Warn("telegram chat rate limited", "chat_id", chatID)
		return
	}

	reply, err := b.reply(chatID, u.Message.Text)
	if err != nil {
		slog.Error("telegram message handling failed", "chat_id", chatID, "error", err)
	}

	if err := b.sendMessage(ctx, chatID, reply); err != nil {
		slog.Error("telegram send failed", "chat_id", chatID, "error", err)
	}
}

// reply routes the message. Commands are the bot's only local knowledge;
// everything else goes to the conversation untouched.
func (b *Bot) reply(chatID, text string) (string, error) {
	command, argument := splitCommand(text)

	switch command {
	case "/start":
		return msgStart, nil
	case "/link":
		return b.service.HandleLink(chatID, argument)
	default:
		return b.service.HandleMessage(chatID, text)
	}
}

// splitCommand separates "/link ABC" into its command and argument, tolerating
// the "@botname" suffix Telegram adds in group chats.
func splitCommand(text string) (string, string) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return "", ""
	}

	command, argument, _ := strings.Cut(trimmed, " ")
	if at := strings.Index(command, "@"); at != -1 {
		command = command[:at]
	}

	return strings.ToLower(command), strings.TrimSpace(argument)
}

func (b *Bot) getUpdates(ctx context.Context) ([]update, error) {
	payload := map[string]any{
		"timeout":         pollTimeoutSeconds,
		"allowed_updates": []string{"message"},
	}
	if b.offset > 0 {
		payload["offset"] = b.offset
	}

	body, err := b.call(ctx, "getUpdates", payload)
	if err != nil {
		return nil, err
	}

	var response updatesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode telegram updates: %w", err)
	}

	if !response.OK {
		return nil, fmt.Errorf("telegram getUpdates rejected: %s", response.Description)
	}

	return response.Result, nil
}

func (b *Bot) sendMessage(ctx context.Context, chatID, text string) error {
	if text == "" {
		return nil
	}

	_, err := b.call(ctx, "sendMessage", map[string]any{
		"chat_id": chatID,
		"text":    text,
	})

	return err
}

func (b *Bot) call(ctx context.Context, method string, payload map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode %s payload: %w", method, err)
	}

	url := fmt.Sprintf("%s/bot%s/%s", b.apiBase, b.token, method)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build %s request: %w", method, err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := b.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("telegram %s request failed: %w", method, err)
	}
	defer response.Body.Close()

	responseBody := new(bytes.Buffer)
	if _, err := responseBody.ReadFrom(response.Body); err != nil {
		return nil, fmt.Errorf("failed to read telegram %s response: %w", method, err)
	}

	if response.StatusCode != http.StatusOK {
		// The token must never reach the logs, so only the status is reported.
		return nil, fmt.Errorf("telegram %s returned status %d", method, response.StatusCode)
	}

	return responseBody.Bytes(), nil
}
