package user

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

const (
	trustedDeviceCookie = "trustedDevice"
	trustedDeviceTTL    = 90 * 24 * time.Hour
)

// Order matters: Edge and Chrome UAs also contain "Safari", and iOS UAs contain "Mac OS X".
var (
	browserLabels = [][2]string{{"Edg/", "Edge"}, {"Firefox/", "Firefox"}, {"Chrome/", "Chrome"}, {"Safari/", "Safari"}}
	osLabels      = [][2]string{{"iPhone", "iPhone"}, {"iPad", "iPad"}, {"Android", "Android"}, {"Mac OS X", "macOS"}, {"Windows", "Windows"}, {"Linux", "Linux"}}
)

func deviceLabel(userAgent string) string {
	return firstMatch(userAgent, browserLabels, "Browser") + " on " + firstMatch(userAgent, osLabels, "unknown device")
}

func firstMatch(userAgent string, labels [][2]string, fallback string) string {
	for _, label := range labels {
		if strings.Contains(userAgent, label[0]) {
			return label[1]
		}
	}
	return fallback
}

// isSecureRequest also trusts X-Forwarded-Proto because prod terminates TLS at nginx,
// so r.TLS is always nil there.
func isSecureRequest(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

// rememberDevice must run before the response body is written, since it sets a cookie.
func (h *Handler) rememberDevice(w http.ResponseWriter, r *http.Request, userID int) error {
	rawToken, err := auth.GenerateOneTimeToken()
	if err != nil {
		return fmt.Errorf("generate trusted device token: %w", err)
	}

	device := &types.TrustedDevice{
		UserID:    userID,
		TokenHash: auth.HashSecret(rawToken),
		Label:     deviceLabel(r.UserAgent()),
		ExpiresAt: time.Now().Add(trustedDeviceTTL),
	}
	if err := h.trustedDevices.CreateTrustedDevice(device); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     trustedDeviceCookie,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(trustedDeviceTTL.Seconds()),
	})
	return nil
}

// trustedDeviceFor returns the request's trusted device when it belongs to userID, or nil.
func (h *Handler) trustedDeviceFor(r *http.Request, userID int) *types.TrustedDevice {
	if h.trustedDevices == nil {
		return nil
	}

	cookie, err := r.Cookie(trustedDeviceCookie)
	if err != nil || cookie.Value == "" {
		return nil
	}

	device, err := h.trustedDevices.GetTrustedDeviceByToken(cookie.Value)
	if err != nil || device.UserID != userID {
		return nil
	}
	return device
}

func (h *Handler) handleListTrustedDevices(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	devices, err := h.trustedDevices.ListTrustedDevices(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if current := h.trustedDeviceFor(r, userID); current != nil {
		for _, device := range devices {
			device.IsCurrent = device.ID == current.ID
		}
	}

	middleware.WriteDataResponse(w, devices)
}

func (h *Handler) handleRevokeTrustedDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	deviceID, ok := middleware.ExtractPathParamAsIntAndRespond(w, r, 2)
	if !ok {
		return
	}

	err := h.trustedDevices.DeleteTrustedDevice(userID, deviceID)
	if errors.Is(err, errTrustedDeviceNotFound) {
		utils.WriteError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteSuccessResponse(w)
}

func (h *Handler) handleRevokeAllTrustedDevices(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	if err := h.trustedDevices.DeleteTrustedDevicesByUser(userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.LogWithContext(r.Context()).Info("all trusted devices revoked", "userID", userID)
	middleware.WriteSuccessResponse(w)
}
