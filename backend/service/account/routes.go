package account

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/lucas-remigio/wallet-tracker/config"
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
	router.HandleFunc("/accounts", h.CreateAccount)
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the user id by the token from authorization
	authToken := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authToken, "Bearer ")
	secret := []byte(config.Envs.JWTSecret)
	userId, err := auth.GetUserIdFromToken(secret, token)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
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

	// create a new account
	err = h.store.CreateAccount(&types.Account{
		UserID:      userId,
		AccountName: payload.AccountName,
		Balance:     payload.Balance,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]string{"status": token})
}
