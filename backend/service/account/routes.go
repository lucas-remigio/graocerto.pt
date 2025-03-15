package account

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	router.HandleFunc("/accounts/{id}", h.ChangeAccountHandler)
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

func (h *Handler) ChangeAccountHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.UpdateAccount(w, r)
	case http.MethodDelete:
		h.DeleteAccount(w, r)
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

func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	// get the account id from the url
	accountId := strings.TrimPrefix(r.URL.Path, "/accounts/")
	if accountId == "" {
		http.Error(w, "Missing account id", http.StatusBadRequest)
		return
	}

	// convert the account id to an int
	accountIdInt, err := strconv.Atoi(accountId)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	// get the JSON payload
	var payload types.UpdateAccountPayload
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

	// update the account
	err = h.store.UpdateAccount(&types.Account{
		ID:          accountIdInt,
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

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	// get the account token from the url
	accountToken := strings.TrimPrefix(r.URL.Path, "/accounts/")
	if accountToken == "" {
		http.Error(w, "Missing account token", http.StatusBadRequest)
		return
	}

	// delete the account
	err = h.store.DeleteAccount(accountToken, userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]string{"status": "success"})
}
