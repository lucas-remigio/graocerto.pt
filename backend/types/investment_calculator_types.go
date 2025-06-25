package types

type InvestmentCalculatorStore interface {
	CalculateInvestmentYearlyReturn(initialInvestment, monthlyContribution, annualReturnRate float64, investmentDurationYears int) (*InvestmentCalculatorResult, error)
}

type InvestmentCalculatorPayload struct {
	InitialInvestment       float64 `json:"initial_investment" validate:"gte=0,lte=100000"`
	MonthlyContribution     float64 `json:"monthly_contribution" validate:"required,gt=0,lte=10000"`
	AnnualReturnRate        float64 `json:"annual_return_rate" validate:"required,gte=0,lte=1"`
	InvestmentDurationYears int     `json:"investment_duration_years" validate:"required,gt=0,lte=100"`
}

type InvestmentCalculatorResult struct {
	TotalInvestment float64           `json:"total_investment"`
	TotalReturn     float64           `json:"total_return"`
	TotalValue      float64           `json:"total_value"`
	YearlyBreakdown []YearlyBreakdown `json:"yearly_breakdown"`
}

type YearlyBreakdown struct {
	Year            int     `json:"year"`
	TotalInvestment float64 `json:"total_investment"`
	TotalReturn     float64 `json:"total_return"`
	TotalValue      float64 `json:"total_value"`
}
