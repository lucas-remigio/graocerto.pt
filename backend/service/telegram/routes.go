package telegram

import (
	"fmt"
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

// Handler exposes the account-owner side of the integration. The bot itself
// runs in-process and never comes through here.
type Handler struct {
	links *LinkService
}

func NewHandler(links *LinkService) *Handler {
	return &Handler{links: links}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/telegram", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodGet:    h.GetStatus,
			http.MethodDelete: h.Unlink,
		}),
	))

	router.HandleFunc("/telegram/link-token", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPost: h.CreateLinkToken,
		}),
	))
}

func (h *Handler) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	code, expiresAt, err := h.links.IssueCode(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create link code"))
		return
	}

	middleware.WriteDataResponse(w, types.TelegramLinkTokenResponse{
		Token:     code,
		ExpiresAt: expiresAt,
	})
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	linked, err := h.links.IsLinked(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to read telegram status"))
		return
	}

	middleware.WriteDataResponse(w, types.TelegramStatusResponse{Linked: linked})
}

func (h *Handler) Unlink(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	if err := h.links.Unlink(userId); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to unlink telegram"))
		return
	}

	middleware.WriteSuccessResponse(w)
}
