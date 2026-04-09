package recurring_rule

import (
	"testing"

	"github.com/lucas-remigio/wallet-tracker/types"
)

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
