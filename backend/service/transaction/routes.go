package transaction

import (
	"net/http"

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
	router.HandleFunc("/transactions/", middleware.AuthMiddleware(h.GetTransactionsByAccountToken))
	router.HandleFunc("/transactions/{id}", middleware.AuthMiddleware(h.ChangeTransaction))
}

func (h *Handler) ChangeTransaction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.UpdateTransaction(w, r)
	case http.MethodDelete:
		h.DeleteTransaction(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateTransaction(w, r)
	case http.MethodGet:
		h.GetTransactionsByAccountToken(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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
	transactions, err := h.store.GetTransactionsByAccountToken(accountToken)
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

	// get transactions DTO by account token
	transactions, err := h.store.GetTransactionsDTOByAccountToken(accountToken)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
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
