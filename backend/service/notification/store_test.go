package notification

import (
	"reflect"
	"testing"
)

func TestBudgetThresholdsCrossed(t *testing.T) {
	tests := []struct {
		name   string
		spent  float64
		budget int
		warn   int
		want   []int
	}{
		{name: "no budget", spent: 100, budget: 0, warn: 80, want: nil},
		{name: "no spend", spent: 0, budget: 400, warn: 80, want: nil},
		{name: "below warn", spent: 316, budget: 400, warn: 80, want: []int{}},             // 79%
		{name: "exactly warn", spent: 320, budget: 400, warn: 80, want: []int{80}},         // 80%
		{name: "between warn and 100", spent: 360, budget: 400, warn: 80, want: []int{80}}, // 90%
		{name: "exactly 100pct", spent: 400, budget: 400, warn: 80, want: []int{80, 100}},
		{name: "over 100pct", spent: 420, budget: 400, warn: 80, want: []int{80, 100}}, // 105%
		{name: "negative spend", spent: -50, budget: 400, warn: 80, want: nil},
		{name: "custom warn 50 crossed", spent: 200, budget: 400, warn: 50, want: []int{50}}, // 50%
		{name: "custom warn 50 below", spent: 100, budget: 400, warn: 50, want: []int{}},     // 25%
		{name: "custom warn 90 not yet", spent: 340, budget: 400, warn: 90, want: []int{}},   // 85%
		{name: "warn 100 dedups over-budget", spent: 400, budget: 400, warn: 100, want: []int{100}},
		{name: "warn 100 below", spent: 360, budget: 400, warn: 100, want: []int{}}, // 90%
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := budgetThresholdsCrossed(tt.spent, tt.budget, tt.warn)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("budgetThresholdsCrossed(%v, %d, %d) = %v, want %v", tt.spent, tt.budget, tt.warn, got, tt.want)
			}
		})
	}
}
