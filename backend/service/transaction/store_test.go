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

func TestComputeCategoryMovers(t *testing.T) {
	// 6-month window: half = 3, so prior = months[0..3), recent = months[3..6).
	roots := []*types.CategoryTrend{
		// prior avg 100, recent avg 150 => +50% up
		{ID: 1, Name: "Dining", Totals: []float64{100, 100, 100, 150, 150, 150}},
		// prior avg 200, recent avg 100 => -50% down
		{ID: 2, Name: "Transport", Totals: []float64{200, 200, 200, 100, 100, 100}},
		// no prior spend, recent spend => new
		{ID: 3, Name: "Gym", Totals: []float64{0, 0, 0, 40, 40, 40}},
		// flat (within ±2%) => excluded
		{ID: 4, Name: "Rent", Totals: []float64{500, 500, 500, 500, 500, 500}},
	}

	movers := computeCategoryMovers(roots, 6)

	byID := make(map[int]*types.CategoryMover)
	for _, m := range movers {
		byID[m.ID] = m
	}

	if _, ok := byID[4]; ok {
		t.Fatalf("flat category should be excluded from movers")
	}
	if len(movers) != 3 {
		t.Fatalf("expected 3 movers, got %d", len(movers))
	}

	if byID[1].Direction != "up" || byID[1].Pct == nil || *byID[1].Pct != 50 {
		t.Fatalf("Dining: got dir=%s pct=%v", byID[1].Direction, byID[1].Pct)
	}
	if byID[2].Direction != "down" || byID[2].Pct == nil || *byID[2].Pct != -50 {
		t.Fatalf("Transport: got dir=%s pct=%v", byID[2].Direction, byID[2].Pct)
	}
	if byID[3].Direction != "new" || byID[3].Pct != nil {
		t.Fatalf("Gym: expected new with nil pct, got dir=%s pct=%v", byID[3].Direction, byID[3].Pct)
	}

	// New floats to the top of the ranking.
	if movers[0].ID != 3 {
		t.Fatalf("expected new category (Gym) first, got id=%d", movers[0].ID)
	}
}

func TestComputeCategoryMovers_TooFewMonths(t *testing.T) {
	roots := []*types.CategoryTrend{{ID: 1, Totals: []float64{100}}}
	if movers := computeCategoryMovers(roots, 1); len(movers) != 0 {
		t.Fatalf("expected no movers for n<2, got %d", len(movers))
	}
}
