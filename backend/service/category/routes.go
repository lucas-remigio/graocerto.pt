package category

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
	store types.CategoryStore
}

func NewHandler(store types.CategoryStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/categories", h.CategoryHandler)
	router.HandleFunc("/categories/dto", h.GetCategoriesDtoByUserId)
	router.HandleFunc("/categories/{id}", h.ChangeCategory)
}

func (h *Handler) ChangeCategory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.UpdateCategory(w, r)
	case http.MethodDelete:
		h.DeleteCategory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) CategoryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateCategory(w, r)
	case http.MethodGet:
		h.GetCategoriesByUserId(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get JSON payload
	var payload types.CreateCategoryPayload
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
	err = h.store.CreateCategory(&types.Category{
		UserID:            userId,
		TransactionTypeID: payload.TransactionTypeId,
		CategoryName:      payload.CategoryName,
		Color:             payload.Color,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetCategoriesByUserId(w http.ResponseWriter, r *http.Request) {
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

	// get categories by user id
	categories, err := h.store.GetCategoriesByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"categories": categories,
	}

	utils.WriteJson(w, http.StatusOK, response)
}

func (h *Handler) GetCategoriesDtoByUserId(w http.ResponseWriter, r *http.Request) {
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

	// get categories by user id
	categories, err := h.store.GetCategoryDtoByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"categories": categories,
	}

	utils.WriteJson(w, http.StatusOK, response)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the category id from the url
	categoryId := strings.TrimPrefix(r.URL.Path, "/categories/")
	if categoryId == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	// convert the category id to int
	categoryIdInt, err := strconv.Atoi(categoryId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// get the user id by the token from authorization
	authToken := r.Header.Get("Authorization")
	userId, err := auth.GetUserIdFromToken(authToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	// get JSON payload
	var payload types.UpdateCategoryPayload
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

	err = h.store.UpdateCategory(&types.Category{
		ID:           categoryIdInt,
		UserID:       userId,
		CategoryName: payload.CategoryName,
		Color:        payload.Color,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the category id from the url
	categoryId := strings.TrimPrefix(r.URL.Path, "/categories/")
	if categoryId == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	// convert the category id to int
	categoryIdInt, err := strconv.Atoi(categoryId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// get the user id by the token from authorization
	authToken := r.Header.Get("Authorization")
	userId, err := auth.GetUserIdFromToken(authToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	err = h.store.DeleteCategory(categoryIdInt, userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]string{"status": "success"})
}
