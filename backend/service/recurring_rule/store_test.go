package recurring_rule

import (
	"testing"

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
