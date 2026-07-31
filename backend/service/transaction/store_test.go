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
	// Month-over-month: only the last two months (indices n-2, n-1) matter.
	roots := []*types.CategoryTrend{
		// 100 -> 150 => +50% up, delta 50
		{ID: 1, Name: "Dining", Totals: []float64{0, 0, 0, 0, 100, 150}},
		// 500 -> 300 => -40% down, delta 200 (biggest money move)
		{ID: 2, Name: "Transport", Totals: []float64{0, 0, 0, 0, 500, 300}},
		// 0 -> 40 => new
		{ID: 3, Name: "Gym", Totals: []float64{0, 0, 0, 0, 0, 40}},
		// flat (within ±2%) => excluded
		{ID: 4, Name: "Rent", Totals: []float64{0, 0, 0, 0, 500, 500}},
		// 200 -> 0 => dropped to zero, skipped (likely "not yet" this month)
		{ID: 5, Name: "Travel", Totals: []float64{0, 0, 0, 0, 200, 0}},
	}

	movers := computeCategoryMovers(roots, 4, 5)

	byID := make(map[int]*types.CategoryMover)
	for _, m := range movers {
		byID[m.ID] = m
	}

	if _, ok := byID[4]; ok {
		t.Fatalf("flat category should be excluded from movers")
	}
	if _, ok := byID[5]; ok {
		t.Fatalf("category that dropped to zero should be excluded from movers")
	}
	if len(movers) != 3 {
		t.Fatalf("expected 3 movers, got %d", len(movers))
	}

	if byID[1].Direction != "up" || byID[1].Pct == nil || *byID[1].Pct != 50 {
		t.Fatalf("Dining: got dir=%s pct=%v", byID[1].Direction, byID[1].Pct)
	}
	if byID[2].Direction != "down" || byID[2].Pct == nil || *byID[2].Pct != -40 {
		t.Fatalf("Transport: got dir=%s pct=%v", byID[2].Direction, byID[2].Pct)
	}
	if byID[3].Direction != "new" || byID[3].Pct != nil {
		t.Fatalf("Gym: expected new with nil pct, got dir=%s pct=%v", byID[3].Direction, byID[3].Pct)
	}

	// Ranked by absolute euro change: Transport (200) > Dining (50) > Gym (40).
	if movers[0].ID != 2 {
		t.Fatalf("expected biggest money mover (Transport) first, got id=%d", movers[0].ID)
	}
}

func TestComputeCategoryMovers_NoComparison(t *testing.T) {
	roots := []*types.CategoryTrend{{ID: 1, Totals: []float64{100}}}
	// baseIdx = -1 signals "no valid pair to compare" (e.g. a single-month window).
	if movers := computeCategoryMovers(roots, -1, 0); len(movers) != 0 {
		t.Fatalf("expected no movers without a comparison pair, got %d", len(movers))
	}
}

func TestDefaultCompareIndices(t *testing.T) {
	// 12-month axis ending on the current month.
	axis := buildMonthAxis(mustDate(t, "2026-08-15"), 12) // ends Aug 2026 (idx 11)

	// Early in the month (< 1/3 elapsed): skip the incomplete current month and
	// compare the two most recent COMPLETE months (Jun→Jul = idx 9→10).
	if b, c := defaultCompareIndices(axis, mustDate(t, "2026-08-02")); b != 9 || c != 10 {
		t.Fatalf("early-month: expected 9,10 got %d,%d", b, c)
	}

	// Late in the month (> 1/3 elapsed): keep current-vs-previous (Jul→Aug = 10→11).
	if b, c := defaultCompareIndices(axis, mustDate(t, "2026-08-28")); b != 10 || c != 11 {
		t.Fatalf("late-month: expected 10,11 got %d,%d", b, c)
	}

	// Fewer than two months: no pair to compare.
	if b, _ := defaultCompareIndices(buildMonthAxis(mustDate(t, "2026-08-15"), 1), mustDate(t, "2026-08-02")); b != -1 {
		t.Fatalf("single-month axis: expected base -1, got %d", b)
	}
}

func TestBucketMonthlyAmounts(t *testing.T) {
	// A 3-month axis: Jan, Feb, Mar 2026.
	indexByKey := map[int]int{
		2026*100 + 1: 0,
		2026*100 + 2: 1,
		2026*100 + 3: 2,
	}
	rows := []*monthlyAmount{
		{year: 2026, month: 1, amount: 100},
		{year: 2026, month: 3, amount: 50.005}, // rounds to 50.01
		{year: 2025, month: 12, amount: 999},   // outside the window -> dropped
	}

	got := bucketMonthlyAmounts(rows, indexByKey, 3)

	want := []float64{100, 0, 50.01}
	if len(got) != len(want) {
		t.Fatalf("expected %d buckets, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("bucket[%d] = %v; want %v", i, got[i], want[i])
		}
	}
}

func TestBucketMonthlyAmounts_Empty(t *testing.T) {
	got := bucketMonthlyAmounts(nil, map[int]int{}, 3)
	if len(got) != 3 {
		t.Fatalf("expected a zeroed slice of length 3, got %d", len(got))
	}
	for i, v := range got {
		if v != 0 {
			t.Fatalf("bucket[%d] = %v; want 0", i, v)
		}
	}
}
