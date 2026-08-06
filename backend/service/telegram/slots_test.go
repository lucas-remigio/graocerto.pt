package telegram

import (
	"testing"

	"github.com/lucas-remigio/wallet-tracker/types"
)

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func pendingWith(items []types.PendingItem, accountToken *string) *types.PendingParse {
	return &types.PendingParse{UserID: 1, ChatID: "chat-1", Items: items, AccountToken: accountToken}
}

// The outstanding question is derived from the data, so these cases pin the
// entire ordering rule in one place.
func TestNextSlot(t *testing.T) {
	resolved := types.PendingItem{Amount: 1, Description: "a", CategoryID: intPtr(1), TransactionTypeID: intPtr(2)}
	unresolved := types.PendingItem{Amount: 2, Description: "b"}

	tests := []struct {
		name      string
		pending   *types.PendingParse
		wantSlot  slot
		wantIndex int
	}{
		{
			name:      "asks for the first missing category",
			pending:   pendingWith([]types.PendingItem{resolved, unresolved, unresolved}, strPtr("tok")),
			wantSlot:  slotCategory,
			wantIndex: 1,
		},
		{
			name:      "categories come before the account",
			pending:   pendingWith([]types.PendingItem{unresolved}, nil),
			wantSlot:  slotCategory,
			wantIndex: 0,
		},
		{
			name:      "asks for the account once every category is known",
			pending:   pendingWith([]types.PendingItem{resolved}, nil),
			wantSlot:  slotAccount,
			wantIndex: -1,
		},
		{
			name:      "empty account token counts as missing",
			pending:   pendingWith([]types.PendingItem{resolved}, strPtr("")),
			wantSlot:  slotAccount,
			wantIndex: -1,
		},
		{
			name:      "everything resolved goes straight to confirmation",
			pending:   pendingWith([]types.PendingItem{resolved, resolved}, strPtr("tok")),
			wantSlot:  slotConfirmation,
			wantIndex: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSlot, gotIndex := nextSlot(tc.pending)
			if gotSlot != tc.wantSlot || gotIndex != tc.wantIndex {
				t.Fatalf("nextSlot = (%v, %d), want (%v, %d)", gotSlot, gotIndex, tc.wantSlot, tc.wantIndex)
			}
		})
	}
}

func TestApplyAnswerFillsCategoryAndDerivesType(t *testing.T) {
	catalog := testCatalog() // 1 Groceries debit, 2 Car debit, 3 Salary credit
	pending := pendingWith([]types.PendingItem{{Amount: 4.5, Description: "gasoil"}}, strPtr("tok-main"))

	confirmed, problem := applyAnswer(pending, "2", catalog)

	if problem != "" {
		t.Fatalf("unexpected problem: %s", problem)
	}
	if confirmed {
		t.Fatal("filling a category must not confirm anything")
	}
	if pending.Items[0].CategoryID == nil || *pending.Items[0].CategoryID != 2 {
		t.Fatalf("expected category 2, got %v", pending.Items[0].CategoryID)
	}
	if pending.Items[0].TransactionTypeID == nil || *pending.Items[0].TransactionTypeID != int(types.DebitTransactionType) {
		t.Fatalf("expected the derived debit type, got %v", pending.Items[0].TransactionTypeID)
	}
}

func TestApplyAnswerRejectsBadSelections(t *testing.T) {
	catalog := testCatalog()

	tests := []struct {
		name  string
		reply string
	}{
		{name: "not a number", reply: "groceries"},
		{name: "zero", reply: "0"},
		{name: "beyond the list", reply: "99"},
		{name: "negative", reply: "-1"},
		{name: "empty", reply: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pending := pendingWith([]types.PendingItem{{Amount: 1, Description: "x"}}, strPtr("tok-main"))

			confirmed, problem := applyAnswer(pending, tc.reply, catalog)

			if problem == "" {
				t.Fatal("expected the answer to be rejected")
			}
			if confirmed {
				t.Fatal("a rejected answer must not confirm")
			}
			if pending.Items[0].CategoryID != nil {
				t.Fatal("a rejected answer must leave the parse untouched")
			}
		})
	}
}

func TestApplyAnswerToleratesPunctuation(t *testing.T) {
	catalog := testCatalog()

	for _, reply := range []string{"2", " 2 ", "2.", "#2", "2)"} {
		pending := pendingWith([]types.PendingItem{{Amount: 1, Description: "x"}}, strPtr("tok-main"))

		if _, problem := applyAnswer(pending, reply, catalog); problem != "" {
			t.Fatalf("reply %q was rejected: %s", reply, problem)
		}

		if pending.Items[0].CategoryID == nil || *pending.Items[0].CategoryID != 2 {
			t.Fatalf("reply %q selected %v", reply, pending.Items[0].CategoryID)
		}
	}
}

func TestApplyAnswerFillsAccount(t *testing.T) {
	// tok-main is the only favourite, so it is the marked default.
	catalog := testCatalog()
	resolved := types.PendingItem{Amount: 1, Description: "x", CategoryID: intPtr(1), TransactionTypeID: intPtr(2)}

	tests := []struct {
		name  string
		reply string
		want  string
	}{
		{name: "by number", reply: "2", want: "tok-savings"},
		{name: "yes takes the default", reply: "yes", want: "tok-main"},
		{name: "empty takes the default", reply: "", want: "tok-main"},
		{name: "portuguese yes", reply: "sim", want: "tok-main"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pending := pendingWith([]types.PendingItem{resolved}, nil)

			confirmed, problem := applyAnswer(pending, tc.reply, catalog)

			if problem != "" {
				t.Fatalf("unexpected problem: %s", problem)
			}
			if confirmed {
				t.Fatal("filling the account must not confirm anything")
			}
			if pending.AccountToken == nil || *pending.AccountToken != tc.want {
				t.Fatalf("expected %q, got %v", tc.want, pending.AccountToken)
			}
		})
	}
}

// With no favourite there is no default, so a bare "yes" cannot pick one.
func TestApplyAnswerAccountWithoutDefaultRequiresANumber(t *testing.T) {
	catalog := NewCatalog(nil, []*types.Account{
		account("a", "Main", false),
		account("b", "Savings", false),
	})
	pending := pendingWith([]types.PendingItem{{Amount: 1, CategoryID: intPtr(1), TransactionTypeID: intPtr(2)}}, nil)

	if _, problem := applyAnswer(pending, "yes", catalog); problem == "" {
		t.Fatal("expected yes to be rejected when there is no default")
	}
	if pending.AccountToken != nil {
		t.Fatal("no account should have been selected")
	}

	if _, problem := applyAnswer(pending, "2", catalog); problem != "" {
		t.Fatalf("expected a number to work: %s", problem)
	}
	if pending.AccountToken == nil || *pending.AccountToken != "b" {
		t.Fatalf("expected account b, got %v", pending.AccountToken)
	}
}

func TestApplyAnswerConfirmation(t *testing.T) {
	catalog := testCatalog()
	resolved := types.PendingItem{Amount: 1, Description: "x", CategoryID: intPtr(1), TransactionTypeID: intPtr(2)}

	confirmed, problem := applyAnswer(pendingWith([]types.PendingItem{resolved}, strPtr("tok-main")), "yes", catalog)
	if problem != "" || !confirmed {
		t.Fatalf("expected yes to confirm, got confirmed=%v problem=%q", confirmed, problem)
	}

	confirmed, problem = applyAnswer(pendingWith([]types.PendingItem{resolved}, strPtr("tok-main")), "maybe", catalog)
	if confirmed || problem == "" {
		t.Fatalf("expected an unclear reply to re-ask, got confirmed=%v problem=%q", confirmed, problem)
	}
}

func TestCancelWords(t *testing.T) {
	for _, word := range []string{"cancel", "/cancel", "Cancelar", " NO ", "não"} {
		if !isCancel(word) {
			t.Fatalf("%q should cancel", word)
		}
	}

	for _, word := range []string{"yes", "2", "cookies", ""} {
		if isCancel(word) {
			t.Fatalf("%q should not cancel", word)
		}
	}
}
