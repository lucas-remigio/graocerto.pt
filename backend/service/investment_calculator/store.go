package investment_calculator

import (
	"fmt"
	"math"

	"github.com/lucas-remigio/wallet-tracker/types"
)

type Store struct {
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) CalculateInvestmentYearlyReturn(initialInvestment, monthlyContribution, annualReturnRate float64, investmentDurationYears int) (*types.InvestmentCalculatorResult, error) {
	// Validate inputs
	if monthlyContribution <= 0 || investmentDurationYears <= 0 {
		return nil, fmt.Errorf("monthly contribution and investment duration must be positive")
	}

	if initialInvestment < 0 {
		return nil, fmt.Errorf("initial investment cannot be negative")
	}

	// Validate return rate (should be between 0 and 1)
	if annualReturnRate < 0 || annualReturnRate > 1 {
		return nil, fmt.Errorf("annual return rate must be between 0 and 1")
	}

	// Pre-calculate common values to avoid redundant calculations
	monthlyReturn := annualReturnRate / 12
	totalMonths := investmentDurationYears * 12

	// Calculate totals using optimized formulas
	totalInvestment := initialInvestment + (monthlyContribution * float64(totalMonths))
	totalValue := calculateCompoundInterest(initialInvestment, monthlyContribution, monthlyReturn, totalMonths)
	totalReturn := totalValue - totalInvestment

	// Generate yearly breakdown with optimized calculations
	yearlyBreakdown := generateYearlyBreakdown(initialInvestment, monthlyContribution, monthlyReturn, investmentDurationYears)

	return &types.InvestmentCalculatorResult{
		TotalInvestment: totalInvestment,
		TotalReturn:     totalReturn,
		TotalValue:      totalValue,
		YearlyBreakdown: yearlyBreakdown,
	}, nil
}

// calculateCompoundInterest calculates final value with optimized math operations
func calculateCompoundInterest(initialInvestment, monthlyContribution, monthlyReturn float64, totalMonths int) float64 {
	// Handle edge case where there's no return
	if monthlyReturn == 0 {
		return initialInvestment + (monthlyContribution * float64(totalMonths))
	}

	// Pre-calculate compound factor to avoid repeated math.Pow calls
	compoundFactor := math.Pow(1+monthlyReturn, float64(totalMonths))

	// Calculate future value of initial investment
	initialValue := initialInvestment * compoundFactor

	// Calculate future value of monthly contributions using annuity formula
	// FV = PMT * [(1 + r)^n - 1] / r
	annuityValue := monthlyContribution * ((compoundFactor - 1) / monthlyReturn)

	return initialValue + annuityValue
}

// generateYearlyBreakdown creates yearly breakdown with optimized calculations
func generateYearlyBreakdown(initialInvestment, monthlyContribution, monthlyReturn float64, years int) []types.YearlyBreakdown {
	breakdown := make([]types.YearlyBreakdown, years)

	// Pre-calculate base values to avoid redundant operations
	oneReturnRate := 1 + monthlyReturn

	for year := 1; year <= years; year++ {
		monthsElapsed := year * 12

		// Calculate total investment (linear growth)
		totalInvestment := initialInvestment + (monthlyContribution * float64(monthsElapsed))

		// Calculate total value with optimized compound formula
		var totalValue float64
		if monthlyReturn == 0 {
			totalValue = totalInvestment
		} else {
			// Use pre-calculated compound factor
			compoundFactor := math.Pow(oneReturnRate, float64(monthsElapsed))
			initialValue := initialInvestment * compoundFactor
			annuityValue := monthlyContribution * ((compoundFactor - 1) / monthlyReturn)
			totalValue = initialValue + annuityValue
		}

		breakdown[year-1] = types.YearlyBreakdown{
			Year:            year,
			TotalInvestment: totalInvestment,
			TotalReturn:     totalValue - totalInvestment,
			TotalValue:      totalValue,
		}
	}

	return breakdown
}
