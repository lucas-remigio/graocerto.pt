package telegram

import (
	"strings"
	"testing"
	"time"

	"github.com/lucas-remigio/wallet-tracker/types"
)

const testChatID = "chat-1"

type conversationFixture struct {
	service      *Service
	llm          *stubLLM
	pending      *memoryPendingStore
	transactions *recordingTransactionWriter
	users        *memoryUserStore
}

// newConversation wires the service against in-memory doubles. accounts is
// varied per test because the account question depends entirely on them.
func newConversation(t *testing.T, accounts []*types.Account) *conversationFixture {
	t.Helper()

	chatID := testChatID
	users := newMemoryUserStore(&types.User{ID: 1, FirstName: "Lucas", TelegramChatID: &chatID})
	pending := newMemoryPendingStore()
	llm := &stubLLM{}
	transactions := newRecordingTransactionWriter(412.31)

	service := NewService(ServiceDeps{
		Users:   users,
		Pending: pending,
		Categories: &stubCategoryReader{categories: []*types.CategoryDTO{
			category(1, "Groceries", types.DebitTransactionType),
			category(2, "Car", types.DebitTransactionType),
			category(3, "Salary", types.CreditTransactionType),
			category(4, "Transfer out", types.TransferTransactionType),
		}},
		Accounts:     &stubAccountReader{accounts: accounts},
		Transactions: transactions,
		LLM:          llm,
		Links:        NewLinkService(users, newMemoryTokenStore(), pending),
	})
	service.now = func() time.Time { return time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC) }

	return &conversationFixture{
		service:      service,
		llm:          llm,
		pending:      pending,
		transactions: transactions,
		users:        users,
	}
}

func twoAccounts() []*types.Account {
	return []*types.Account{
		account("tok-main", "Main", true),
		account("tok-savings", "Savings", false),
	}
}

func threeAccountsTwoFavourites() []*types.Account {
	return []*types.Account{
		account("tok-main", "Main", true),
		account("tok-savings", "Savings", true),
		account("tok-cash", "Cash", false),
	}
}

func (f *conversationFixture) send(t *testing.T, text string) string {
	t.Helper()

	reply, err := f.service.HandleMessage(testChatID, text)
	if err != nil {
		t.Fatalf("HandleMessage(%q) returned error: %v", text, err)
	}

	return reply
}

// Everything supplied by the message: one confirmation and done.
func TestConversationJumpsStraightToConfirmation(t *testing.T) {
	f := newConversation(t, twoAccounts())
	f.llm.response = `{"transactions":[{"amount":3.19,"description":"cookies","category_id":1,"confidence":0.95}],"account_token":"tok-savings"}`

	reply := f.send(t, "3.19 in cookies from savings")

	for _, want := range []string{"1 transaction", "Savings", "€3.19", "cookies", "Groceries", "debit"} {
		if !strings.Contains(reply, want) {
			t.Fatalf("summary is missing %q:\n%s", want, reply)
		}
	}

	if len(f.transactions.created) != 0 {
		t.Fatal("nothing may be written before the user confirms")
	}

	reply = f.send(t, "yes")

	if len(f.transactions.created) != 1 {
		t.Fatalf("expected 1 transaction written, got %d", len(f.transactions.created))
	}

	written := f.transactions.created[0]
	if written.AccountToken != "tok-savings" || written.CategoryId != 1 || written.Amount != 3.19 {
		t.Fatalf("wrong transaction written: %+v", written)
	}
	if written.Date != "2026-08-06" {
		t.Fatalf("expected today's date, got %q", written.Date)
	}
	if f.transactions.userIDs[0] != 1 {
		t.Fatalf("transaction written for the wrong user: %d", f.transactions.userIDs[0])
	}

	if !strings.Contains(reply, "Added 1 transaction") || !strings.Contains(reply, "€412.31") {
		t.Fatalf("unexpected confirmation reply:\n%s", reply)
	}

	if parse, _ := f.pending.GetPendingParse(testChatID); parse != nil {
		t.Fatal("the pending conversation should be cleared after committing")
	}
}

// The full slot-filling path: missing category, then the account, then confirm.
func TestConversationAsksCategoryThenAccount(t *testing.T) {
	f := newConversation(t, threeAccountsTwoFavourites())
	f.llm.response = `{"transactions":[
		{"amount":3.19,"description":"cookies","category_id":1,"confidence":0.95},
		{"amount":4.50,"description":"gasoil","category_id":null,"confidence":0}
	],"account_token":null}`

	reply := f.send(t, "3.19 in cookies, 4.50 gasoil")
	if !strings.Contains(reply, "gasoil") || !strings.Contains(reply, "2. Car") {
		t.Fatalf("expected a category question listing the options:\n%s", reply)
	}

	reply = f.send(t, "2")
	if !strings.Contains(reply, "Which account?") || !strings.Contains(reply, "Main (default)") {
		t.Fatalf("expected the account question with a marked default:\n%s", reply)
	}

	reply = f.send(t, "yes")
	if !strings.Contains(reply, "I understood 2 transactions") || !strings.Contains(reply, "Main") {
		t.Fatalf("expected the summary on the default account:\n%s", reply)
	}

	f.send(t, "yes")

	if len(f.transactions.created) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(f.transactions.created))
	}
	if f.transactions.created[1].CategoryId != 2 {
		t.Fatalf("the answered category was not applied: %+v", f.transactions.created[1])
	}
	for _, written := range f.transactions.created {
		if written.AccountToken != "tok-main" {
			t.Fatalf("expected the default account, got %q", written.AccountToken)
		}
	}
}

// A single account means there is nothing to ask about.
func TestConversationSkipsAccountQuestionWhenThereIsNoChoice(t *testing.T) {
	f := newConversation(t, []*types.Account{account("tok-only", "Wallet", false)})
	f.llm.response = `{"transactions":[{"amount":10,"description":"lunch","category_id":1,"confidence":0.95}],"account_token":null}`

	reply := f.send(t, "10 lunch")

	if strings.Contains(reply, "Which account?") {
		t.Fatalf("should not ask about the only account:\n%s", reply)
	}
	if !strings.Contains(reply, "Wallet") {
		t.Fatalf("the summary must still name the account:\n%s", reply)
	}
}

func TestConversationRepeatsQuestionOnBadAnswer(t *testing.T) {
	f := newConversation(t, twoAccounts())
	f.llm.response = `{"transactions":[{"amount":4.50,"description":"gasoil","category_id":null,"confidence":0}],"account_token":"tok-main"}`

	f.send(t, "4.50 gasoil")
	reply := f.send(t, "99")

	if !strings.Contains(reply, msgPickCategoryNumber) {
		t.Fatalf("expected a rejection message:\n%s", reply)
	}
	if !strings.Contains(reply, "1. Groceries") {
		t.Fatalf("expected the list to be repeated:\n%s", reply)
	}

	parse, _ := f.pending.GetPendingParse(testChatID)
	if parse == nil || parse.Items[0].CategoryID != nil {
		t.Fatal("a rejected answer must leave the stored parse untouched")
	}
}

func TestConversationCancel(t *testing.T) {
	f := newConversation(t, twoAccounts())
	f.llm.response = `{"transactions":[{"amount":10,"description":"lunch","category_id":1,"confidence":0.95}],"account_token":"tok-main"}`

	f.send(t, "10 lunch")

	if reply := f.send(t, "cancel"); reply != msgCancelled {
		t.Fatalf("unexpected cancel reply: %s", reply)
	}

	if parse, _ := f.pending.GetPendingParse(testChatID); parse != nil {
		t.Fatal("cancel must discard the pending parse")
	}
	if len(f.transactions.created) != 0 {
		t.Fatal("cancel must not write anything")
	}

	if reply := f.send(t, "cancel"); reply != msgNothingToDo {
		t.Fatalf("expected nothing to cancel, got: %s", reply)
	}
}

// A second "yes" arrives after the pending row is gone, so it is treated as new
// input rather than booking the same items twice.
func TestConversationDoesNotDoubleBookOnRepeatedYes(t *testing.T) {
	f := newConversation(t, twoAccounts())
	f.llm.response = `{"transactions":[{"amount":10,"description":"lunch","category_id":1,"confidence":0.95}],"account_token":"tok-main"}`

	f.send(t, "10 lunch")
	f.send(t, "yes")

	if len(f.transactions.created) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(f.transactions.created))
	}

	f.llm.response = `{"transactions":[],"account_token":null}`
	reply := f.send(t, "yes")

	if len(f.transactions.created) != 1 {
		t.Fatalf("a repeated yes booked again: %d transactions", len(f.transactions.created))
	}
	if reply != msgNotUnderstood {
		t.Fatalf("unexpected reply to a stray yes: %s", reply)
	}
}

func TestConversationRequiresALinkedChat(t *testing.T) {
	f := newConversation(t, twoAccounts())

	reply, err := f.service.HandleMessage("unknown-chat", "10 lunch")
	if err != nil {
		t.Fatalf("an unlinked chat is not an error: %v", err)
	}
	if reply != msgNotLinked {
		t.Fatalf("expected the linking instructions, got: %s", reply)
	}
	if f.llm.lastPrompt != "" {
		t.Fatal("an unlinked chat must never reach the llm")
	}
}

func TestConversationHandlesUnparseableMessage(t *testing.T) {
	f := newConversation(t, twoAccounts())
	f.llm.response = `{"transactions":[],"account_token":null}`

	if reply := f.send(t, "good morning"); reply != msgNotUnderstood {
		t.Fatalf("unexpected reply: %s", reply)
	}

	if parse, _ := f.pending.GetPendingParse(testChatID); parse != nil {
		t.Fatal("an empty parse must not start a conversation")
	}
}

func TestConversationWithoutAccountsOrCategories(t *testing.T) {
	f := newConversation(t, nil)
	if reply := f.send(t, "10 lunch"); reply != msgNoAccounts {
		t.Fatalf("expected the no-accounts message, got: %s", reply)
	}

	f = newConversation(t, twoAccounts())
	f.service.categories = &stubCategoryReader{}
	if reply := f.send(t, "10 lunch"); reply != msgNoCategories {
		t.Fatalf("expected the no-categories message, got: %s", reply)
	}
}

// A partial write must be reported, never rounded up to success.
func TestConversationReportsPartialFailure(t *testing.T) {
	f := newConversation(t, twoAccounts())
	f.transactions.successLimit = 1
	f.llm.response = `{"transactions":[
		{"amount":1,"description":"first","category_id":1,"confidence":0.95},
		{"amount":2,"description":"second","category_id":1,"confidence":0.95}
	],"account_token":"tok-main"}`

	f.send(t, "1 first, 2 second")

	reply, err := f.service.HandleMessage(testChatID, "yes")
	if err == nil {
		t.Fatal("expected the write failure to surface as an error for the log")
	}

	if !strings.Contains(reply, "Added 1 transaction") || !strings.Contains(reply, "1 of 2 could not be saved") {
		t.Fatalf("the reply must admit the partial failure:\n%s", reply)
	}

	if parse, _ := f.pending.GetPendingParse(testChatID); parse != nil {
		t.Fatal("the pending parse must be cleared even on partial failure")
	}
}

func TestConversationReportsTotalFailure(t *testing.T) {
	f := newConversation(t, twoAccounts())
	f.transactions.successLimit = 0
	f.llm.response = `{"transactions":[{"amount":1,"description":"only","category_id":1,"confidence":0.95}],"account_token":"tok-main"}`

	f.send(t, "1 only")

	reply, err := f.service.HandleMessage(testChatID, "yes")
	if err == nil {
		t.Fatal("expected an error for the log")
	}
	if reply != msgNothingSaved {
		t.Fatalf("expected an honest failure message, got:\n%s", reply)
	}
}

func TestHandleLink(t *testing.T) {
	f := newConversation(t, twoAccounts())

	if reply, _ := f.service.HandleLink("chat-9", ""); reply != msgLinkUsage {
		t.Fatalf("expected usage help, got: %s", reply)
	}

	if reply, _ := f.service.HandleLink("chat-9", "BADCODE1"); reply != msgLinkInvalid {
		t.Fatalf("expected an invalid code message, got: %s", reply)
	}

	code, _, err := f.service.links.IssueCode(1)
	if err != nil {
		t.Fatalf("IssueCode returned error: %v", err)
	}

	reply, err := f.service.HandleLink("chat-9", code)
	if err != nil {
		t.Fatalf("HandleLink returned error: %v", err)
	}
	if !strings.Contains(reply, "Lucas") {
		t.Fatalf("expected a greeting naming the user, got: %s", reply)
	}
}
