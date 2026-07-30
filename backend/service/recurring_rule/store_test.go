package recurring_rule

import (
	"testing"
	"time"

	"github.com/lucas-remigio/wallet-tracker/types"
)

func strPtr(s string) *string { return &s }

func TestPlanPendingGeneration(t *testing.T) {
	groupA := "group-a"

	rules := []*types.RecurringRule{
		{ID: 1}, // standalone
		{ID: 2, UserID: 7, RecurringTransferGroupID: &groupA}, // transfer side A
		{ID: 3, UserID: 7, RecurringTransferGroupID: &groupA}, // transfer side B (same group)
		{ID: 4}, // standalone
		{ID: 5, RecurringTransferGroupID: strPtr("")}, // empty group => standalone
	}

	units := planPendingGeneration(rules)

	if len(units) != 4 {
		t.Fatalf("expected 4 units, got %d", len(units))
	}

	// Transfer group collapses to a single unit carrying a representative rule.
	transferUnits := 0
	for _, u := range units {
		if u.Transfer != nil {
			transferUnits++
			if u.Transfer.RecurringTransferGroupID == nil || *u.Transfer.RecurringTransferGroupID != groupA {
				t.Fatalf("transfer unit has wrong group id: %+v", u.Transfer)
			}
			if u.Transfer.UserID != 7 {
				t.Fatalf("transfer unit lost user id: %+v", u.Transfer)
			}
			if u.Single != nil {
				t.Fatalf("unit should not be both single and transfer")
			}
		}
	}
	if transferUnits != 1 {
		t.Fatalf("expected transfer group to collapse into 1 unit, got %d", transferUnits)
	}

	// The 3 non-transfer rows (ids 1, 4, 5) should each be a standalone unit.
	singles := 0
	for _, u := range units {
		if u.Single != nil {
			singles++
		}
	}
	if singles != 3 {
		t.Fatalf("expected 3 standalone units, got %d", singles)
	}
}

func TestCalculateNextRunDate(t *testing.T) {
	tests := []struct {
		name       string
		current    string
		frequency  types.RecurringFrequency
		interval   int
		want       string
		shouldFail bool
	}{
		{name: "daily", current: "2026-04-09", frequency: types.RecurringDaily, interval: 1, want: "2026-04-10"},
		{name: "weekly", current: "2026-04-09", frequency: types.RecurringWeekly, interval: 2, want: "2026-04-23"},
		{name: "monthly", current: "2026-04-09", frequency: types.RecurringMonthly, interval: 1, want: "2026-05-09"},
		{name: "every_x_days", current: "2026-04-09", frequency: types.RecurringEveryXDays, interval: 5, want: "2026-04-14"},
		{name: "invalid_frequency", current: "2026-04-09", frequency: "invalid", interval: 1, shouldFail: true},
		{name: "invalid_date", current: "2026-99-09", frequency: types.RecurringDaily, interval: 1, shouldFail: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateNextRunDate(tt.current, tt.frequency, tt.interval)
			if tt.shouldFail {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEndOfMonthWindow(t *testing.T) {
	tests := []struct {
		name      string
		now       string
		wantStart string
		wantEnd   string
		wantDays  int
	}{
		{name: "mid month", now: "2026-07-29", wantStart: "2026-07-29", wantEnd: "2026-07-31", wantDays: 2},
		{name: "last day", now: "2026-07-31", wantStart: "2026-07-31", wantEnd: "2026-07-31", wantDays: 0},
		{name: "first day", now: "2026-07-01", wantStart: "2026-07-01", wantEnd: "2026-07-31", wantDays: 30},
		{name: "february non-leap", now: "2026-02-15", wantStart: "2026-02-15", wantEnd: "2026-02-28", wantDays: 13},
		{name: "february leap", now: "2028-02-15", wantStart: "2028-02-15", wantEnd: "2028-02-29", wantDays: 14},
		{name: "december rollover", now: "2026-12-10", wantStart: "2026-12-10", wantEnd: "2026-12-31", wantDays: 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now, err := time.Parse("2006-01-02", tt.now)
			if err != nil {
				t.Fatalf("bad test date: %v", err)
			}
			start, end := endOfMonthWindow(now)
			if start.Format("2006-01-02") != tt.wantStart {
				t.Fatalf("start: got %s, want %s", start.Format("2006-01-02"), tt.wantStart)
			}
			if end.Format("2006-01-02") != tt.wantEnd {
				t.Fatalf("end: got %s, want %s", end.Format("2006-01-02"), tt.wantEnd)
			}
			if days := int(end.Sub(start).Hours() / 24); days != tt.wantDays {
				t.Fatalf("days remaining: got %d, want %d", days, tt.wantDays)
			}
		})
	}
}

func TestExpandRuleOccurrences(t *testing.T) {
	// July 2026 window: today through end of month.
	windowStart, _ := time.Parse("2006-01-02", "2026-07-29")
	windowEnd, _ := time.Parse("2006-01-02", "2026-07-31")

	monthly := func(id int, typeID int, amount float64, nextRun string) *types.RecurringRule {
		return &types.RecurringRule{
			ID:                id,
			AccountToken:      "acc",
			CategoryID:        1,
			TransactionTypeID: typeID,
			Amount:            amount,
			Frequency:         types.RecurringMonthly,
			IntervalValue:     1,
			NextRunDate:       nextRun,
		}
	}

	t.Run("no rules -> zero summary", func(t *testing.T) {
		items, summary, err := expandRuleOccurrences(nil, windowStart, windowEnd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected no items, got %d", len(items))
		}
		if summary.Credit != 0 || summary.Debit != 0 || summary.Difference != 0 {
			t.Fatalf("expected zero summary, got %+v", summary)
		}
	})

	t.Run("debit due before month end is included", func(t *testing.T) {
		rules := []*types.RecurringRule{monthly(1, transactionTypeDebit, 50, "2026-07-30")}
		items, summary, err := expandRuleOccurrences(rules, windowStart, windowEnd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if summary.Debit != 50 || summary.Difference != -50 {
			t.Fatalf("expected debit 50 / difference -50, got %+v", summary)
		}
	})

	t.Run("occurrence after month end is excluded", func(t *testing.T) {
		rules := []*types.RecurringRule{monthly(1, transactionTypeCredit, 100, "2026-08-05")}
		items, summary, err := expandRuleOccurrences(rules, windowStart, windowEnd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected no items, got %d", len(items))
		}
		if summary.Difference != 0 {
			t.Fatalf("expected zero difference, got %+v", summary)
		}
	})

	t.Run("credit and debit net into difference", func(t *testing.T) {
		rules := []*types.RecurringRule{
			monthly(1, transactionTypeCredit, 1000, "2026-07-29"),
			monthly(2, transactionTypeDebit, 250, "2026-07-31"),
		}
		_, summary, err := expandRuleOccurrences(rules, windowStart, windowEnd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary.Credit != 1000 || summary.Debit != 250 || summary.Difference != 750 {
			t.Fatalf("expected 1000/250/750, got %+v", summary)
		}
	})

	t.Run("invalid next run date errors", func(t *testing.T) {
		rules := []*types.RecurringRule{monthly(1, transactionTypeDebit, 10, "not-a-date")}
		if _, _, err := expandRuleOccurrences(rules, windowStart, windowEnd); err == nil {
			t.Fatalf("expected error for invalid date, got nil")
		}
	})
}
