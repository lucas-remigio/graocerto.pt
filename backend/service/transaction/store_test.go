package transaction

import (
	"testing"

	"github.com/lucas-remigio/wallet-tracker/types"
)

func TestCalculateTransactionTotals_IgnoresPending(t *testing.T) {
	s := &Store{}
	transactions := []*types.TransactionDTO{
		{
			Amount:    100,
			IsPending: false,
			Category: &types.CategoryDTO{
				TransactionType: &types.TransactionType{ID: int(types.CreditTransactionType)},
			},
		},
		{
			Amount:    30,
			IsPending: true,
			Category: &types.CategoryDTO{
				TransactionType: &types.TransactionType{ID: int(types.DebitTransactionType)},
			},
		},
	}

	totals, err := s.CalculateTransactionTotals(transactions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if totals.Credit != 100 {
		t.Fatalf("expected credit 100, got %v", totals.Credit)
	}
	if totals.Debit != 0 {
		t.Fatalf("expected debit 0, got %v", totals.Debit)
	}
	if totals.Difference != 100 {
		t.Fatalf("expected difference 100, got %v", totals.Difference)
	}
}
