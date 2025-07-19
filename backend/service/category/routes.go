package category

import (
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Handler struct {
	store types.CategoryStore
}

func NewHandler(store types.CategoryStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/categories", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPost: h.CreateCategory,
			http.MethodGet:  h.GetCategoriesByUserId,
		}),
	))
	router.HandleFunc("/categories/dto", middleware.AuthMiddleware(h.GetCategoriesDtoByUserId))
	router.HandleFunc("/categories/{id}", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPut:    h.UpdateCategory,
			http.MethodDelete: h.DeleteCategory,
		}),
	))
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	// parse and validate JSON payload
	var payload types.CreateCategoryPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// create a new category
	err := h.store.CreateCategory(&types.Category{
		UserID:            userId,
		TransactionTypeID: payload.TransactionTypeId,
		CategoryName:      payload.CategoryName,
		Color:             payload.Color,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) GetCategoriesByUserId(w http.ResponseWriter, r *http.Request) {
	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// get categories by user id
	categories, err := h.store.GetCategoriesByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"categories": categories,
	}

	middleware.WriteDataResponse(w, response)
}

func (h *Handler) GetCategoriesDtoByUserId(w http.ResponseWriter, r *http.Request) {
	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// get categories by user id
	categories, err := h.store.GetCategoryDtoByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"categories": categories,
	}

	middleware.WriteDataResponse(w, response)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	// get the category id from the url using path segment extraction
	categoryIdInt, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 1)
	if !ok {
		return
	}

	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// parse and validate JSON payload
	var payload types.UpdateCategoryPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	err := h.store.UpdateCategory(&types.Category{
		ID:           categoryIdInt,
		UserID:       userId,
		CategoryName: payload.CategoryName,
		Color:        payload.Color,
	}, userId)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	// get the category id from the url using path segment extraction
	categoryIdInt, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 1)
	if !ok {
		return
	}

	// require authentication
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	err := h.store.DeleteCategory(categoryIdInt, userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}
