package telegram

import (
	"testing"

	"github.com/lucas-remigio/wallet-tracker/types"
)

func category(id int, name string, typeID types.TransactionTypeID) *types.CategoryDTO {
	slug := map[types.TransactionTypeID]string{
		types.CreditTransactionType:   "credit",
		types.DebitTransactionType:    "debit",
		types.TransferTransactionType: "transfer",
	}[typeID]

	return &types.CategoryDTO{
		ID:              id,
		CategoryName:    name,
		TransactionType: &types.TransactionType{ID: int(typeID), TypeSlug: slug},
	}
}

func account(token, name string, favorite bool) *types.Account {
	return &types.Account{Token: token, AccountName: name, IsFavorite: favorite}
}

// A transfer category needs two accounts and CreateTransfer; booking one as a
// single transaction would leave the other side of the movement missing.
func TestCatalogExcludesTransferAndDeletedCategories(t *testing.T) {
	deletedAt := "2026-01-01"
	deleted := category(4, "Old", types.DebitTransactionType)
	deleted.DeletedAt = &deletedAt

	catalog := NewCatalog([]*types.CategoryDTO{
		category(1, "Groceries", types.DebitTransactionType),
		category(2, "Salary", types.CreditTransactionType),
		category(3, "Transfer out", types.TransferTransactionType),
		deleted,
	}, nil)

	if len(catalog.Categories) != 2 {
		t.Fatalf("expected 2 selectable categories, got %d", len(catalog.Categories))
	}

	if catalog.Category(3) != nil {
		t.Fatal("transfer category must not be selectable")
	}

	if catalog.Category(4) != nil {
		t.Fatal("soft deleted category must not be selectable")
	}

	if catalog.Category(1) == nil || catalog.Category(2) == nil {
		t.Fatal("expected credit and debit categories to be selectable")
	}
}

func TestCategoryLabelQualifiesSubcategories(t *testing.T) {
	parentID := 1
	child := category(2, "Fuel", types.DebitTransactionType)
	child.ParentCategoryId = &parentID

	orphan := category(3, "Ghost", types.DebitTransactionType)
	missingParent := 99
	orphan.ParentCategoryId = &missingParent

	catalog := NewCatalog([]*types.CategoryDTO{
		category(1, "Car", types.DebitTransactionType),
		child,
		orphan,
	}, nil)

	if got := catalog.CategoryLabel(catalog.Category(2)); got != "Fuel (Car)" {
		t.Fatalf("expected qualified label, got %q", got)
	}

	if got := catalog.CategoryLabel(catalog.Category(1)); got != "Car" {
		t.Fatalf("expected bare label for a root category, got %q", got)
	}

	if got := catalog.CategoryLabel(catalog.Category(3)); got != "Ghost" {
		t.Fatalf("expected bare label when the parent is unknown, got %q", got)
	}
}

func TestAccountResolution(t *testing.T) {
	tests := []struct {
		name        string
		accounts    []*types.Account
		wantAuto    string // token, empty means the user must choose
		wantDefault string // token marked (default), empty means none
	}{
		{
			name:     "single account needs no question",
			accounts: []*types.Account{account("a", "Main", false)},
			wantAuto: "a",
		},
		{
			name: "single favourite needs no question",
			accounts: []*types.Account{
				account("a", "Main", false),
				account("b", "Savings", true),
			},
			wantAuto:    "b",
			wantDefault: "b",
		},
		{
			name: "several favourites ask, first one is the default",
			accounts: []*types.Account{
				account("a", "Main", true),
				account("b", "Savings", true),
				account("c", "Cash", false),
			},
			wantDefault: "a",
		},
		{
			name: "no favourites ask with no default",
			accounts: []*types.Account{
				account("a", "Main", false),
				account("b", "Savings", false),
			},
		},
		{
			name: "no accounts at all",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			catalog := NewCatalog(nil, tc.accounts)

			auto := catalog.AutoAccount()
			if tc.wantAuto == "" && auto != nil {
				t.Fatalf("expected the user to be asked, got auto account %q", auto.Token)
			}
			if tc.wantAuto != "" && (auto == nil || auto.Token != tc.wantAuto) {
				t.Fatalf("expected auto account %q, got %v", tc.wantAuto, auto)
			}

			def := catalog.DefaultAccount()
			if tc.wantDefault == "" && def != nil {
				t.Fatalf("expected no default, got %q", def.Token)
			}
			if tc.wantDefault != "" && (def == nil || def.Token != tc.wantDefault) {
				t.Fatalf("expected default %q, got %v", tc.wantDefault, def)
			}
		})
	}
}
