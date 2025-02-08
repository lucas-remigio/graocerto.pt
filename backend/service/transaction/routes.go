package transaction

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
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
	router.HandleFunc("/transactions", h.CreateTransaction)
	router.HandleFunc("/transactions/dto/", h.GetTransactionsDTOByAccountToken)
	router.HandleFunc("/transactions/", h.GetTransactionsByAccountToken)
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get JSON payload
	var payload types.CreateTransactionPayload
	if err := utils.ParseJson(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		error := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", error))
		return
	}

	// get the user id by the token from authorization
	authToken := r.Header.Get("Authorization")
	_, err := auth.GetUserIdFromToken(authToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	// create a new account
	err = h.store.CreateTransaction(&types.Transaction{
		AccountToken: payload.AccountToken,
		Amount:       payload.Amount,
		CategoryId:   payload.CategoryID,
		Description:  payload.Description,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetTransactionsByAccountToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the user id by the token from authorization
	authToken := r.Header.Get("Authorization")
	_, err := auth.GetUserIdFromToken(authToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	// Example URL: /transactions/{account_token}
	// Remove the prefix to get the account token.
	accountToken := strings.TrimPrefix(r.URL.Path, "/transactions/")
	if accountToken == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing account token"))
		return
	}

	// get categories by user id
	transactions, err := h.store.GetTransactionsByAccountToken(accountToken)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
	}

	utils.WriteJson(w, http.StatusOK, response)
}

func (h *Handler) GetTransactionsDTOByAccountToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the user id by the token from authorization
	authToken := r.Header.Get("Authorization")
	_, err := auth.GetUserIdFromToken(authToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	// Example URL: /transactions/{account_token}
	// Remove the prefix to get the account token.
	accountToken := strings.TrimPrefix(r.URL.Path, "/transactions/dto/")
	if accountToken == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing account token"))
		return
	}

	// get categories by user id
	transactions, err := h.store.GetTransactionsDTOByAccountToken(accountToken)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
	}

	utils.WriteJson(w, http.StatusOK, response)
}
