package account

import (
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/middleware"
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
	router.HandleFunc("/accounts", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPost: h.CreateAccount,
			http.MethodGet:  h.GetAccountsByUserId,
		})))
	router.HandleFunc("/accounts/reorder", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPost: h.ReorderAccounts,
		}),
	))
	router.HandleFunc("/accounts/{id}", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPut:    h.UpdateAccount,
			http.MethodDelete: h.DeleteAccount,
		}),
	))
	router.HandleFunc("/accounts/{token}/favorite", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPatch: h.FavoriteAccount,
		}),
	))
	router.HandleFunc("/accounts/{id}/feedback-month", middleware.AuthMiddleware(h.GetAccountFeedbackMonthly))

}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	// parse and validate JSON payload
	var payload types.CreateAccountPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// create a new account
	err := h.store.CreateAccount(&types.Account{
		UserID:      userId,
		AccountName: payload.AccountName,
		Balance:     *payload.Balance,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) GetAccountsByUserId(w http.ResponseWriter, r *http.Request) {
	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
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

	middleware.WriteDataResponse(w, response)
}

func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// extract account ID from URL path (/accounts/{id})
	accountIdInt, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 1)
	if !ok {
		return
	}

	// parse and validate JSON payload
	var payload types.UpdateAccountPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// update the account
	err := h.store.UpdateAccount(&types.Account{
		ID:          accountIdInt,
		UserID:      userId,
		AccountName: payload.AccountName,
		Balance:     *payload.Balance,
	}, userId)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// extract account token from URL path (/accounts/{token})
	accountToken, ok := middleware.ExtractPathParamAndRespond(w, r, 1)
	if !ok {
		return
	}

	// delete the account
	err := h.store.DeleteAccount(accountToken, userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) GetAccountFeedbackMonthly(w http.ResponseWriter, r *http.Request) {
	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// extract account token from URL path (/accounts/{accountId}/feedback-month)
	accountToken, ok := middleware.ExtractPathParamAndRespond(w, r, 1)
	if !ok {
		return
	}

	// get month and year from query parameters
	month, err := utils.GetIntFromQuery(r, "month")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	year, err := utils.GetIntFromQuery(r, "year")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	language, err := utils.GetStringFromQuery(r, "language")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// get the account feedback monthly
	feedback, err := h.store.GetAccountFeedbackMonthly(userId, accountToken, language, month, year)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, feedback)
}

func (h *Handler) ReorderAccounts(w http.ResponseWriter, r *http.Request) {
	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// parse and validate JSON payload
	var payload types.ReorderAccountsPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// call store to update order indexes
	err := h.store.ReorderAccounts(userId, payload.Accounts)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) FavoriteAccount(w http.ResponseWriter, r *http.Request) {
	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// extract account token from URL path (/accounts/{token}/favorite)
	accountToken, ok := middleware.ExtractPathParamAndRespond(w, r, 1)
	if !ok {
		return
	}

	// parse and validate JSON payload
	var payload types.FavoriteAccountPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// call store to update favorite status
	err := h.store.FavoriteAccount(accountToken, userId, payload.IsFavorite)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}
