package account

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Handler struct {
	store types.AccountStore
}

func NewHandler(store types.AccountStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/accounts", h.AccountsHandler)
}

func (h *Handler) AccountsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateAccount(w, r)
	case http.MethodGet:
		h.GetAccountsByUserId(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get JSON payload
	var payload types.CreateAccountPayload
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
	userId, err := auth.GetUserIdFromToken(authToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	// create a new account
	err = h.store.CreateAccount(&types.Account{
		UserID:      userId,
		AccountName: payload.AccountName,
		Balance:     *payload.Balance,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetAccountsByUserId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the user id by the token from authorization
	authToken := r.Header.Get("Authorization")
	userId, err := auth.GetUserIdFromToken(authToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	// get accounts by user id
	accounts, err := h.store.GetAccountsByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"accounts": accounts,
	}

	utils.WriteJson(w, http.StatusOK, response)
}
