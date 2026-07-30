package transaction

import (
	"testing"
	"time"

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

func TestBuildMonthAxis(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2026-07-15")

	axis := buildMonthAxis(now, 12)
	if len(axis) != 12 {
		t.Fatalf("expected 12 months, got %d", len(axis))
	}
	// Chronological, ending in the current month.
	if axis[11].Month != 7 || axis[11].Year != 2026 {
		t.Fatalf("last axis point = %d/%d; want 7/2026", axis[11].Month, axis[11].Year)
	}
	// 12 months back from July 2026 => August 2025 is the first point.
	if axis[0].Month != 8 || axis[0].Year != 2025 {
		t.Fatalf("first axis point = %d/%d; want 8/2025", axis[0].Month, axis[0].Year)
	}

	// Year rollover within a short window.
	short := buildMonthAxis(mustDate(t, "2026-01-20"), 3)
	want := []types.TrendMonth{{Month: 11, Year: 2025}, {Month: 12, Year: 2025}, {Month: 1, Year: 2026}}
	for i, w := range want {
		if short[i] != w {
			t.Fatalf("short[%d] = %+v; want %+v", i, short[i], w)
		}
	}
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("bad date %q: %v", s, err)
	}
	return d
}
