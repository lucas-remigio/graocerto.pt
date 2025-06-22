package transaction_types

import (
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Handler struct {
	store types.TransactionTypesStore
}

func NewHandler(store types.TransactionTypesStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/transaction-types", middleware.AuthMiddleware(h.TransactionTypesHandler))
}

func (h *Handler) TransactionTypesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetTransactionTypes(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) GetTransactionTypes(w http.ResponseWriter, r *http.Request) {
	// require authentication
	_, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// get transactionTypes
	transactionTypes, err := h.store.GetTransactionTypes()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"transaction_types": transactionTypes,
	}

	middleware.WriteDataResponse(w, response)
}
