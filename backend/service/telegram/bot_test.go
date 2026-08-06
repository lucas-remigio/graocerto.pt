package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// telegramServer stands in for the Telegram API: it hands out a scripted batch
// of updates and records everything the bot sends back.
type telegramServer struct {
	server   *httptest.Server
	sent     []map[string]any
	requests []map[string]any
	batches  [][]update
}

func newTelegramServer(t *testing.T, batches [][]update) *telegramServer {
	t.Helper()

	fake := &telegramServer{batches: batches}
	fake.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)

		switch {
		case strings.HasSuffix(r.URL.Path, "/getUpdates"):
			fake.requests = append(fake.requests, payload)

			batch := []update{}
			if len(fake.batches) > 0 {
				batch = fake.batches[0]
				fake.batches = fake.batches[1:]
			}

			_ = json.NewEncoder(w).Encode(updatesResponse{OK: true, Result: batch})

		case strings.HasSuffix(r.URL.Path, "/sendMessage"):
			fake.sent = append(fake.sent, payload)
			fmt.Fprint(w, `{"ok":true}`)

		default:
			http.Error(w, "unexpected method", http.StatusNotFound)
		}
	}))

	t.Cleanup(fake.server.Close)
	return fake
}

func textUpdate(id, chatID int64, text string) update {
	u := update{UpdateID: id}
	u.Message = &struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	}{Text: text}
	u.Message.Chat.ID = chatID
	return u
}

func newBotForTest(t *testing.T, fake *telegramServer, limiter ChatLimiter) (*Bot, *conversationFixture) {
	t.Helper()

	f := newConversation(t, twoAccounts())
	f.llm.response = `{"transactions":[{"amount":10,"description":"lunch","category_id":1,"confidence":0.95}],"account_token":"tok-main"}`

	// Telegram chat ids are numeric, so the linked chat must match what the
	// bot derives from an update.
	linkedChat := "1"
	f.users.users[1].TelegramChatID = &linkedChat

	bot := NewBot("test-token", f.service, limiter)
	bot.apiBase = fake.server.URL

	return bot, f
}

func TestBotRoutesCommandsAndMessages(t *testing.T) {
	// chat 1 is the linked chat in the conversation fixture.
	fake := newTelegramServer(t, [][]update{{
		textUpdate(10, 1, "/start"),
		textUpdate(11, 1, "10 lunch"),
	}})
	bot, _ := newBotForTest(t, fake, nil)

	if err := bot.pollOnce(context.Background()); err != nil {
		t.Fatalf("pollOnce returned error: %v", err)
	}

	if len(fake.sent) != 2 {
		t.Fatalf("expected 2 replies, got %d", len(fake.sent))
	}

	if !strings.Contains(fake.sent[0]["text"].(string), "/link") {
		t.Fatalf("/start should explain linking, got: %v", fake.sent[0]["text"])
	}

	if !strings.Contains(fake.sent[1]["text"].(string), "I understood") {
		t.Fatalf("expected a summary for the transaction, got: %v", fake.sent[1]["text"])
	}

	// Replies must go back to the chat they came from.
	if fake.sent[0]["chat_id"] != "1" {
		t.Fatalf("replied to the wrong chat: %v", fake.sent[0]["chat_id"])
	}
}

func TestBotAdvancesOffsetSoUpdatesAreNotRedelivered(t *testing.T) {
	fake := newTelegramServer(t, [][]update{
		{textUpdate(10, 1, "hello"), textUpdate(11, 1, "hello again")},
		{},
	})
	bot, _ := newBotForTest(t, fake, nil)
	bot.service.parser = NewParser(&stubLLM{response: `{"transactions":[],"account_token":null}`})

	ctx := context.Background()
	if err := bot.pollOnce(ctx); err != nil {
		t.Fatalf("first pollOnce returned error: %v", err)
	}
	if err := bot.pollOnce(ctx); err != nil {
		t.Fatalf("second pollOnce returned error: %v", err)
	}

	if bot.offset != 12 {
		t.Fatalf("expected the cursor past the last update, got %d", bot.offset)
	}

	if len(fake.requests) != 2 {
		t.Fatalf("expected 2 poll requests, got %d", len(fake.requests))
	}

	// The first poll has no cursor yet; the second must acknowledge the batch.
	if _, ok := fake.requests[0]["offset"]; ok {
		t.Fatal("the first poll should not send an offset")
	}
	if got := fake.requests[1]["offset"]; got != float64(12) {
		t.Fatalf("expected offset 12 on the second poll, got %v", got)
	}
}

func TestBotLinkCommand(t *testing.T) {
	fake := newTelegramServer(t, [][]update{{textUpdate(1, 9, "/link@GraoCertoBot NOTACODE")}})
	bot, _ := newBotForTest(t, fake, nil)

	if err := bot.pollOnce(context.Background()); err != nil {
		t.Fatalf("pollOnce returned error: %v", err)
	}

	if len(fake.sent) != 1 || fake.sent[0]["text"] != msgLinkInvalid {
		t.Fatalf("expected the invalid code reply, got %v", fake.sent)
	}
}

// The limiter guards LLM spend, so a throttled chat must not reach the parser.
func TestBotDropsRateLimitedChats(t *testing.T) {
	fake := newTelegramServer(t, [][]update{{textUpdate(1, 1, "10 lunch")}})
	bot, f := newBotForTest(t, fake, ChatLimiterFunc(func(string) bool { return false }))

	if err := bot.pollOnce(context.Background()); err != nil {
		t.Fatalf("pollOnce returned error: %v", err)
	}

	if len(fake.sent) != 0 {
		t.Fatalf("a rate limited chat should get no reply, got %v", fake.sent)
	}
	if f.llm.lastPrompt != "" {
		t.Fatal("a rate limited chat must never reach the llm")
	}
	if bot.offset != 2 {
		t.Fatal("a dropped update must still be acknowledged")
	}
}

func TestBotIgnoresNonTextUpdates(t *testing.T) {
	fake := newTelegramServer(t, [][]update{{
		{UpdateID: 5},           // e.g. a photo, no message body we handle
		textUpdate(6, 1, "   "), // whitespace only
	}})
	bot, _ := newBotForTest(t, fake, nil)

	if err := bot.pollOnce(context.Background()); err != nil {
		t.Fatalf("pollOnce returned error: %v", err)
	}

	if len(fake.sent) != 0 {
		t.Fatalf("expected no replies, got %v", fake.sent)
	}
	if bot.offset != 7 {
		t.Fatalf("ignored updates must still advance the cursor, got %d", bot.offset)
	}
}

func TestBotSurfacesApiFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	f := newConversation(t, twoAccounts())
	bot := NewBot("bad-token", f.service, nil)
	bot.apiBase = server.URL

	err := bot.pollOnce(context.Background())
	if err == nil {
		t.Fatal("expected a poll error")
	}
	if strings.Contains(err.Error(), "bad-token") {
		t.Fatalf("the bot token must never appear in errors: %v", err)
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		text     string
		command  string
		argument string
	}{
		{text: "/start", command: "/start"},
		{text: "/link ABC123", command: "/link", argument: "ABC123"},
		{text: "/link@GraoCertoBot ABC123", command: "/link", argument: "ABC123"},
		{text: "  /LINK abc  ", command: "/link", argument: "abc"},
		{text: "3.19 cookies", command: "", argument: ""},
	}

	for _, tc := range tests {
		command, argument := splitCommand(tc.text)
		if command != tc.command || argument != tc.argument {
			t.Fatalf("splitCommand(%q) = (%q, %q), want (%q, %q)", tc.text, command, argument, tc.command, tc.argument)
		}
	}
}
