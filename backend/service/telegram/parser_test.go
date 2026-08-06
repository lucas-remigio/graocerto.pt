package telegram

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lucas-remigio/wallet-tracker/types"
)

// stubLLM records the prompt it was given and replays a canned completion,
// exactly as the real client would after stripping the JSON out of the reply.
type stubLLM struct {
	response   string
	err        error
	lastPrompt string
}

func (s *stubLLM) GenerateGPT4Response(prompt string) (string, error) {
	s.lastPrompt = prompt
	if s.err != nil {
		return "", s.err
	}
	return s.response, nil
}

func testCatalog() *Catalog {
	return NewCatalog(
		[]*types.CategoryDTO{
			category(1, "Groceries", types.DebitTransactionType),
			category(2, "Car", types.DebitTransactionType),
			category(3, "Salary", types.CreditTransactionType),
			category(4, "Transfer out", types.TransferTransactionType),
		},
		[]*types.Account{
			account("tok-main", "Main", true),
			account("tok-savings", "Savings", false),
		},
	)
}

func TestParseResolvesItemsAndAccount(t *testing.T) {
	llm := &stubLLM{response: `{"transactions":[
		{"amount":3.19,"description":"cookies","category_id":1,"confidence":0.95},
		{"amount":4.50,"description":"gasoil","category_id":null,"confidence":0}
	],"account_token":"tok-savings"}`}

	items, accountToken, err := NewParser(llm).Parse("3.19 cookies, 4.50 gasoil from savings", testCatalog())
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].CategoryID == nil || *items[0].CategoryID != 1 {
		t.Fatalf("expected category 1, got %v", items[0].CategoryID)
	}

	// The type is derived from the category, never taken from the model.
	if items[0].TransactionTypeID == nil || *items[0].TransactionTypeID != int(types.DebitTransactionType) {
		t.Fatalf("expected the debit type to be derived, got %v", items[0].TransactionTypeID)
	}

	if items[1].CategoryID != nil {
		t.Fatal("an unmatched category must stay unresolved so the user is asked")
	}
	if items[1].TransactionTypeID != nil {
		t.Fatal("no type can be derived without a category")
	}

	if accountToken == nil || *accountToken != "tok-savings" {
		t.Fatalf("expected the named account, got %v", accountToken)
	}
}

func TestParseRejectsCategoriesTheUserCannotUse(t *testing.T) {
	tests := []struct {
		name       string
		categoryID string
	}{
		{name: "category belonging to another user", categoryID: "99"},
		{name: "transfer category", categoryID: "4"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			llm := &stubLLM{response: fmt.Sprintf(
				`{"transactions":[{"amount":10,"description":"x","category_id":%s,"confidence":0.95}],"account_token":null}`,
				tc.categoryID,
			)}

			items, _, err := NewParser(llm).Parse("whatever", testCatalog())
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}

			if len(items) != 1 {
				t.Fatalf("expected the item to survive, got %d", len(items))
			}

			if items[0].CategoryID != nil {
				t.Fatalf("expected the category to be rejected, got %v", *items[0].CategoryID)
			}
		})
	}
}

// A category the model is only guessing at is worse than no category: the user
// gets asked instead, which costs one tap and cannot misfile the money.
func TestParseRequiresConfidenceToAcceptACategory(t *testing.T) {
	tests := []struct {
		name       string
		confidence string
		wantKept   bool
	}{
		{name: "sure", confidence: `"confidence":0.95`, wantKept: true},
		{name: "exactly at the threshold", confidence: `"confidence":0.75`, wantKept: true},
		{name: "just below the threshold", confidence: `"confidence":0.74`},
		{name: "guessing", confidence: `"confidence":0.5`},
		{name: "no opinion", confidence: `"confidence":0`},
		// If the model stops answering the question, ask the user rather than
		// silently falling back to trusting it.
		{name: "field omitted entirely", confidence: `"unused":0`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			llm := &stubLLM{response: fmt.Sprintf(
				`{"transactions":[{"amount":3.19,"description":"cookies","category_id":1,%s}],"account_token":null}`,
				tc.confidence,
			)}

			items, _, err := NewParser(llm).Parse("3.19 cookies", testCatalog())
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}

			if len(items) != 1 {
				t.Fatalf("the transaction itself must survive, got %d items", len(items))
			}

			kept := items[0].CategoryID != nil
			if kept != tc.wantKept {
				t.Fatalf("category kept = %v, want %v", kept, tc.wantKept)
			}

			// A rejected category must not leave a derived type behind.
			if !kept && items[0].TransactionTypeID != nil {
				t.Fatal("no type may be derived from a rejected category")
			}
		})
	}
}

func TestPromptAsksForConfidence(t *testing.T) {
	llm := &stubLLM{response: `{"transactions":[],"account_token":null}`}

	if _, _, err := NewParser(llm).Parse("3.19 cookies", testCatalog()); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if !strings.Contains(llm.lastPrompt, "confidence") {
		t.Fatalf("the prompt no longer asks for a confidence score:\n%s", llm.lastPrompt)
	}
}

func TestParseRejectsForeignAccountToken(t *testing.T) {
	llm := &stubLLM{response: `{"transactions":[{"amount":10,"description":"x","category_id":1,"confidence":0.95}],"account_token":"someone-elses"}`}

	_, accountToken, err := NewParser(llm).Parse("whatever", testCatalog())
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if accountToken != nil {
		t.Fatalf("expected an unknown account token to be dropped, got %q", *accountToken)
	}
}

func TestParseDropsUnbookableAmounts(t *testing.T) {
	llm := &stubLLM{response: `{"transactions":[
		{"amount":0,"description":"zero","category_id":1,"confidence":0.95},
		{"amount":-5,"description":"negative","category_id":1,"confidence":0.95},
		{"amount":1000000000,"description":"too big","category_id":1,"confidence":0.95},
		{"amount":12.5,"description":"fine","category_id":1,"confidence":0.95}
	],"account_token":null}`}

	items, _, err := NewParser(llm).Parse("whatever", testCatalog())
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected only the valid amount to survive, got %d items", len(items))
	}

	if items[0].Description != "fine" {
		t.Fatalf("kept the wrong item: %q", items[0].Description)
	}
}

func TestParseTruncatesLongDescription(t *testing.T) {
	long := strings.Repeat("a", 400)
	llm := &stubLLM{response: fmt.Sprintf(
		`{"transactions":[{"amount":1,"description":%q,"category_id":1,"confidence":0.95}],"account_token":null}`, long,
	)}

	items, _, err := NewParser(llm).Parse("whatever", testCatalog())
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(items[0].Description) != maxDescriptionLength {
		t.Fatalf("expected description truncated to %d, got %d", maxDescriptionLength, len(items[0].Description))
	}
}

func TestParseEmptyResult(t *testing.T) {
	llm := &stubLLM{response: `{"transactions":[],"account_token":null}`}

	items, accountToken, err := NewParser(llm).Parse("good morning", testCatalog())
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("expected no items, got %d", len(items))
	}
	if accountToken != nil {
		t.Fatal("expected no account")
	}
}

func TestParseSurfacesErrors(t *testing.T) {
	if _, _, err := NewParser(&stubLLM{err: fmt.Errorf("upstream down")}).Parse("x", testCatalog()); err == nil {
		t.Fatal("expected the llm error to surface")
	}

	if _, _, err := NewParser(&stubLLM{response: "not json"}).Parse("x", testCatalog()); err == nil {
		t.Fatal("expected a decode error")
	}
}

// The model may only ever see ids the user owns and can actually book against.
func TestPromptOffersOnlyUsableValues(t *testing.T) {
	llm := &stubLLM{response: `{"transactions":[],"account_token":null}`}

	if _, _, err := NewParser(llm).Parse("3.19 cookies", testCatalog()); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	prompt := llm.lastPrompt

	for _, want := range []string{"id=1", "id=2", "id=3", "token=tok-main", "token=tok-savings", "3.19 cookies"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt is missing %q:\n%s", want, prompt)
		}
	}

	if strings.Contains(prompt, "id=4") {
		t.Fatal("prompt offered a transfer category to the model")
	}
}
