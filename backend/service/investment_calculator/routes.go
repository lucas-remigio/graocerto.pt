package investment_calculator

import (
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Handler struct {
	store types.InvestmentCalculatorStore
}

func NewHandler(store types.InvestmentCalculatorStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/investment-calculator", h.handleCalculate)
}

func (h *Handler) handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse and validate JSON payload
	var payload types.InvestmentCalculatorPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// call the store to calculate the investment
	result, err := h.store.CalculateInvestmentYearlyReturn(
		payload.InitialInvestment,
		payload.MonthlyContribution,
		payload.AnnualReturnRate,
		payload.InvestmentDurationYears,
	)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, result)
}
