package transaction

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Handler struct {
	store types.TransactionStore
}

func NewHandler(store types.TransactionStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/transactions", middleware.AuthMiddleware(h.CreateTransaction))
	router.HandleFunc("/transactions/dto/", middleware.AuthMiddleware(h.GetTransactionsDTOByAccountToken))
	router.HandleFunc("/transactions/statistics/", middleware.AuthMiddleware(h.GetTransactionStatistics))
	router.HandleFunc("/transactions/", middleware.AuthMiddleware(h.GetTransactionsByAccountToken))
	router.HandleFunc("/transactions/{id}", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPut:    h.UpdateTransaction,
			http.MethodDelete: h.DeleteTransaction,
		})))
	router.HandleFunc("/transactions/months/{accountToken}", middleware.AuthMiddleware(h.GetTransactionsMonthsAndYears))
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	// parse and validate JSON payload
	var payload types.CreateTransactionPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// require authentication
	_, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// create a new transaction
	err := h.store.CreateTransaction(&types.Transaction{
		AccountToken: payload.AccountToken,
		Amount:       payload.Amount,
		CategoryId:   payload.CategoryID,
		Description:  payload.Description,
		Date:         payload.Date,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) GetTransactionsByAccountToken(w http.ResponseWriter, r *http.Request) {
	// require authentication
	_, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// extract account token from URL path (/transactions/{token})
	accountToken, ok := middleware.ExtractPathParamAndRespond(w, r, 1)
	if !ok {
		return
	}

	// get transactions by account token
	transactions, err := h.store.GetTransactionsByAccountToken(accountToken, nil, nil)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
	}

	middleware.WriteDataResponse(w, response)
}

func (h *Handler) GetTransactionsDTOByAccountToken(w http.ResponseWriter, r *http.Request) {
	// require authentication
	_, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// extract account token from URL path (/transactions/dto/{token})
	accountToken, ok := middleware.ExtractPathParamAndRespond(w, r, 2)
	if !ok {
		return
	}

	// Parse query parameters for month and year filtering
	query := r.URL.Query()
	monthStr := query.Get("month")
	yearStr := query.Get("year")

	var month, year *int
	if monthStr != "" && yearStr != "" {
		monthVal, monthErr := strconv.Atoi(monthStr)
		yearVal, yearErr := strconv.Atoi(yearStr)

		if monthErr != nil || yearErr != nil || monthVal < 1 || monthVal > 12 {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid month or year parameters"))
			return
		}

		month = &monthVal
		year = &yearVal
	}

	// Always return grouped transactions (with optional month/year filter)
	groupedResponse, err := h.store.GetGroupedTransactionsDTOByAccountToken(accountToken, month, year)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, groupedResponse)
}

func (h *Handler) GetTransactionsMonthsAndYears(w http.ResponseWriter, r *http.Request) {
	// require authentication
	_, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// extract account token from URL path (/transactions/months/{token})
	accountToken, ok := middleware.ExtractPathParamAndRespond(w, r, 2)
	if !ok {
		return
	}

	// get transactions months and years by account token
	months, err := h.store.GetAvailableTransactionMonthsByAccountToken(accountToken)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"months": months,
	}

	middleware.WriteDataResponse(w, response)
}

func (h *Handler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	// extract transaction ID from URL path (/transactions/{id})
	transactionIdInt, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 1)
	if !ok {
		return
	}

	// parse and validate JSON payload
	var payload types.UpdateTransactionPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// require authentication
	_, ok = middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	err := h.store.UpdateTransaction(&types.UpdateTransactionPayload{
		ID:          transactionIdInt,
		Amount:      payload.Amount,
		CategoryID:  payload.CategoryID,
		Description: payload.Description,
		Date:        payload.Date,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	// extract transaction ID from URL path (/transactions/{id})
	transactionIdInt, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 1)
	if !ok {
		return
	}

	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	err := h.store.DeleteTransaction(transactionIdInt, userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) GetTransactionStatistics(w http.ResponseWriter, r *http.Request) {
	// extract account token from URL path (/transactions/statistics/{accountToken})
	accountToken, ok := middleware.ExtractPathParamAndRespond(w, r, 2)
	if !ok {
		return
	}

	// require authentication
	_, ok = middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// Parse query parameters for month and year
	var month, year *int

	if monthStr := r.URL.Query().Get("month"); monthStr != "" {
		if monthInt, err := strconv.Atoi(monthStr); err == nil {
			month = &monthInt
		}
	}

	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if yearInt, err := strconv.Atoi(yearStr); err == nil {
			year = &yearInt
		}
	}

	// Get statistics from store
	statistics, err := h.store.GetTransactionStatistics(accountToken, month, year)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, statistics)
}
