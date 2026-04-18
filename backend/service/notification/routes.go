package notification

import (
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Handler struct {
	store types.NotificationStore
}

func NewHandler(store types.NotificationStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/notifications", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodGet: h.GetNotifications,
		}),
	))

	router.HandleFunc("/notifications/unread-count", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodGet: h.GetUnreadNotificationCount,
		}),
	))

	router.HandleFunc("/notifications/{id}/read", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPatch: h.MarkNotificationAsRead,
		}),
	))

	router.HandleFunc("/notifications/preferences", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodGet: h.GetNotificationPreferences,
			http.MethodPut: h.UpdateNotificationPreferences,
		}),
	))

	router.HandleFunc("/notifications/push-subscriptions", middleware.AuthMiddleware(
		middleware.MethodRouter(map[string]http.HandlerFunc{
			http.MethodPost: h.CreatePushSubscription,
			http.MethodDelete: h.DeletePushSubscription,
		}),
	))
}

func (h *Handler) CreatePushSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	var payload types.PushSubscription
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	if err := h.store.CreatePushSubscription(userID, &payload); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) DeletePushSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Endpoint string `json:"endpoint" validate:"required"`
	}
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	if err := h.store.DeletePushSubscription(userID, payload.Endpoint); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	notifications, err := h.store.GetNotificationsByUserID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, types.NotificationsResponse{
		Notifications: notifications,
	})
}

func (h *Handler) GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	count, err := h.store.GetUnreadNotificationCount(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, types.UnreadNotificationCountResponse{
		Count: count,
	})
}

func (h *Handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	notificationID, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 1)
	if !ok {
		return
	}

	if err := h.store.MarkNotificationAsRead(notificationID, userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	preferences, err := h.store.GetNotificationPreferences(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, types.NotificationPreferencesResponse{
		Preferences: preferences,
	})
}

func (h *Handler) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	var payload types.UpdateNotificationPreferencesPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	preferences, err := h.store.UpdateNotificationPreferences(userID, &payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, types.NotificationPreferencesResponse{
		Preferences: preferences,
	})
}
