package recurring_rule

import (
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Handler struct {
	store types.RecurringRuleStore
}

func NewHandler(store types.RecurringRuleStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/recurring-rules", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPost: h.CreateRecurringRule,
			http.MethodGet:  h.GetRecurringRules,
		}),
	))

	router.HandleFunc("/recurring-rules/{id}", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPut:    h.UpdateRecurringRule,
			http.MethodDelete: h.DeleteRecurringRule,
		}),
	))

	router.HandleFunc("/recurring-transfers", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPost: h.CreateRecurringTransfer,
		}),
	))

	router.HandleFunc("/recurring-transfers/{groupID}", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPut: h.UpdateRecurringTransfer,
		}),
	))
}

func (h *Handler) GetRecurringRules(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	recurringRules, err := h.store.GetRecurringRulesByUserID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, map[string]interface{}{
		"recurring_rules": recurringRules,
	})
}

func (h *Handler) CreateRecurringRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	var payload types.CreateRecurringRulePayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	active := true
	if payload.Active != nil {
		active = *payload.Active
	}

	created, err := h.store.CreateRecurringRule(&types.RecurringRule{
		UserID:            userID,
		AccountToken:      payload.AccountToken,
		CategoryID:        payload.CategoryID,
		TransactionTypeID: payload.TransactionTypeID,
		Amount:            payload.Amount,
		Description:       payload.Description,
		Frequency:         payload.Frequency,
		IntervalValue:     payload.IntervalValue,
		NextRunDate:       calculateInitialNextRunDate(payload.Frequency, payload.ExecutionDay, payload.ExecutionWeekday),
		Active:            active,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, map[string]interface{}{
		"recurring_rule": created,
	})
}

func (h *Handler) UpdateRecurringRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	ruleID, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 1)
	if !ok {
		return
	}

	var payload types.UpdateRecurringRulePayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	current, err := h.store.GetRecurringRuleByID(ruleID, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	nextRunDate := current.NextRunDate
	if payload.NextRunDate != "" {
		nextRunDate = payload.NextRunDate
	} else if payload.ExecutionDay != nil || payload.ExecutionWeekday != nil {
		nextRunDate = calculateInitialNextRunDate(
			payload.Frequency,
			payload.ExecutionDay,
			payload.ExecutionWeekday,
		)
	}

	updated, err := h.store.UpdateRecurringRule(&types.RecurringRule{
		ID:                ruleID,
		UserID:            userID,
		AccountToken:      payload.AccountToken,
		CategoryID:        payload.CategoryID,
		TransactionTypeID: payload.TransactionTypeID,
		Amount:            payload.Amount,
		Description:       payload.Description,
		Frequency:         payload.Frequency,
		IntervalValue:     payload.IntervalValue,
		NextRunDate:       nextRunDate,
		Active:            payload.Active,
	}, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, map[string]interface{}{
		"recurring_rule": updated,
	})
}

func (h *Handler) DeleteRecurringRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	ruleID, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 1)
	if !ok {
		return
	}

	if err := h.store.DeleteRecurringRule(ruleID, userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) CreateRecurringTransfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	var payload types.CreateRecurringTransferPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	createdRules, err := h.store.CreateRecurringTransfer(&payload, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, map[string]interface{}{
		"recurring_rules": createdRules,
	})
}

func (h *Handler) UpdateRecurringTransfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	groupID, ok := middleware.ExtractPathParamAndRespond(w, r, 1)
	if !ok {
		return
	}

	var payload types.UpdateRecurringTransferPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	updatedRules, err := h.store.UpdateRecurringTransfer(groupID, &payload, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, map[string]interface{}{
		"recurring_rules": updatedRules,
	})
}
