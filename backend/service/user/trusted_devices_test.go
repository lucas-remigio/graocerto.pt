package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type memoryTrustedDeviceStore struct {
	devices map[int]*types.TrustedDevice
	nextID  int
}

func newMemoryTrustedDeviceStore() *memoryTrustedDeviceStore {
	return &memoryTrustedDeviceStore{devices: map[int]*types.TrustedDevice{}, nextID: 1}
}

func (s *memoryTrustedDeviceStore) CreateTrustedDevice(device *types.TrustedDevice) error {
	copyDevice := *device
	copyDevice.ID = s.nextID
	s.nextID++
	s.devices[copyDevice.ID] = &copyDevice
	return nil
}

func (s *memoryTrustedDeviceStore) GetTrustedDeviceByToken(token string) (*types.TrustedDevice, error) {
	hash := auth.HashSecret(token)
	for _, device := range s.devices {
		if device.TokenHash == hash && time.Now().Before(device.ExpiresAt) {
			return device, nil
		}
	}
	return nil, errTrustedDeviceNotFound
}

func (s *memoryTrustedDeviceStore) ListTrustedDevices(userID int) ([]*types.TrustedDevice, error) {
	devices := []*types.TrustedDevice{}
	for _, device := range s.devices {
		if device.UserID == userID {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func (s *memoryTrustedDeviceStore) TouchTrustedDevice(id int) error {
	s.devices[id].LastUsedAt = time.Now()
	return nil
}

func (s *memoryTrustedDeviceStore) DeleteTrustedDevice(userID, id int) error {
	device, ok := s.devices[id]
	if !ok || device.UserID != userID {
		return errTrustedDeviceNotFound
	}
	delete(s.devices, id)
	return nil
}

func (s *memoryTrustedDeviceStore) DeleteTrustedDevicesByUser(userID int) error {
	for id, device := range s.devices {
		if device.UserID == userID {
			delete(s.devices, id)
		}
	}
	return nil
}

type trustedDeviceFixture struct {
	handler   *Handler
	authStore *memoryAuthTokenStore
	devices   *memoryTrustedDeviceStore
	mailer    *memoryMailer
}

// newTrustedDeviceFixture creates two verified email-OTP users: ids 1 and 2.
func newTrustedDeviceFixture(t *testing.T) *trustedDeviceFixture {
	t.Helper()
	userStore := newMemoryUserStore()
	hashedPassword, _ := auth.HashPassword("Password1!")
	for _, email := range []string{"john@example.com", "jane@example.com"} {
		_ = userStore.CreateUser(context.Background(), &types.User{
			FirstName:     "Test",
			LastName:      "User",
			Email:         email,
			Password:      hashedPassword,
			EmailVerified: true,
			MfaMethod:     types.MfaMethodEmailOTP,
		})
	}

	f := &trustedDeviceFixture{
		authStore: newMemoryAuthTokenStore(),
		devices:   newMemoryTrustedDeviceStore(),
		mailer:    &memoryMailer{},
	}
	f.handler = NewHandler(userStore, f.authStore, f.devices, nil, f.mailer, nil, nil, nil)
	return f
}

// trustDevice stores a device for userID and returns the raw cookie value.
func (f *trustedDeviceFixture) trustDevice(userID int, expiresAt time.Time) string {
	rawToken := fmt.Sprintf("device-token-%d-%d", userID, f.devices.nextID)
	_ = f.devices.CreateTrustedDevice(&types.TrustedDevice{
		UserID:    userID,
		TokenHash: auth.HashSecret(rawToken),
		Label:     "Chrome on macOS",
		ExpiresAt: expiresAt,
	})
	return rawToken
}

func newRequest(method, path string, body any, deviceToken string, userID int) *http.Request {
	marshalled, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(marshalled))
	if deviceToken != "" {
		req.AddCookie(&http.Cookie{Name: trustedDeviceCookie, Value: deviceToken})
	}
	if userID != 0 {
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	}
	return req
}

func serve(handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}

func loginRequest(password, deviceToken string) *http.Request {
	return newRequest(http.MethodPost, "/login", types.LoginUserPayload{
		Email:    "john@example.com",
		Password: password,
	}, deviceToken, 0)
}

func findCookie(rr *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func TestVerifyLoginOTPRemembersDeviceWhenAsked(t *testing.T) {
	for _, remember := range []bool{true, false} {
		t.Run(fmt.Sprintf("remember=%v", remember), func(t *testing.T) {
			f := newTrustedDeviceFixture(t)
			_ = f.authStore.CreateAuthToken(&types.AuthToken{
				ID:          "challenge",
				UserID:      1,
				Purpose:     types.AuthTokenPurposeLoginOTP,
				SecretHash:  auth.HashSecret("123456"),
				ExpiresAt:   time.Now().Add(time.Minute),
				MaxAttempts: loginOTPAttempts,
			})

			req := newRequest(http.MethodPost, "/auth/verify-login-otp", types.VerifyLoginOTPPayload{
				ChallengeID:    "challenge",
				Code:           "123456",
				RememberDevice: remember,
			}, "", 0)
			req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Version/17.0 Mobile Safari/604.1")
			rr := serve(f.handler.handleVerifyLoginOTP, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
			}

			cookie := findCookie(rr, trustedDeviceCookie)
			if !remember {
				if cookie != nil || len(f.devices.devices) != 0 {
					t.Fatal("expected no trusted device without remember_device")
				}
				return
			}

			if cookie == nil || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
				t.Fatalf("expected an HttpOnly SameSite=Strict cookie, got %+v", cookie)
			}
			device, err := f.devices.GetTrustedDeviceByToken(cookie.Value)
			if err != nil {
				t.Fatalf("expected stored device for cookie: %v", err)
			}
			if device.TokenHash == cookie.Value {
				t.Fatal("raw token must not be stored")
			}
			if device.Label != "Safari on iPhone" {
				t.Fatalf("unexpected label %q", device.Label)
			}
		})
	}
}

func TestLoginWithTrustedDeviceSkipsSecondFactor(t *testing.T) {
	f := newTrustedDeviceFixture(t)
	token := f.trustDevice(1, time.Now().Add(time.Hour))

	rr := serve(f.handler.handleLogin, loginRequest("Password1!", token))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(f.mailer.messages) != 0 {
		t.Fatal("expected no OTP email for a trusted device")
	}
	if findCookie(rr, "authToken") == nil {
		t.Fatal("expected a session cookie")
	}
}

func TestLoginFallsBackToChallengeForUntrustedCookies(t *testing.T) {
	f := newTrustedDeviceFixture(t)
	cases := map[string]string{
		"no cookie":      "",
		"unknown token":  "not-a-real-token",
		"expired device": f.trustDevice(1, time.Now().Add(-time.Minute)),
		"other user":     f.trustDevice(2, time.Now().Add(time.Hour)),
	}

	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			rr := serve(f.handler.handleLogin, loginRequest("Password1!", token))
			if rr.Code != http.StatusAccepted {
				t.Fatalf("expected 202 challenge, got %d", rr.Code)
			}
		})
	}
}

func TestLoginWithTrustedDeviceStillRequiresPassword(t *testing.T) {
	f := newTrustedDeviceFixture(t)
	token := f.trustDevice(1, time.Now().Add(time.Hour))

	rr := serve(f.handler.handleLogin, loginRequest("WrongPassword1!", token))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestResetPasswordRevokesTrustedDevices(t *testing.T) {
	f := newTrustedDeviceFixture(t)
	f.trustDevice(1, time.Now().Add(time.Hour))
	f.trustDevice(2, time.Now().Add(time.Hour))
	_ = f.authStore.CreateAuthToken(&types.AuthToken{
		ID:          "reset-token",
		UserID:      1,
		Purpose:     types.AuthTokenPurposePasswordReset,
		SecretHash:  auth.HashSecret("reset-token"),
		ExpiresAt:   time.Now().Add(time.Hour),
		MaxAttempts: 3,
	})

	rr := serve(f.handler.handleResetPassword, newRequest(http.MethodPost, "/auth/reset-password", types.ResetPasswordPayload{
		Token:    "reset-token",
		Password: "NewPassword1!",
	}, "", 0))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if devices, _ := f.devices.ListTrustedDevices(1); len(devices) != 0 {
		t.Fatalf("expected user 1 devices revoked, got %d", len(devices))
	}
	if devices, _ := f.devices.ListTrustedDevices(2); len(devices) != 1 {
		t.Fatal("expected other users' devices untouched")
	}
}

func TestListTrustedDevicesFlagsCurrentAndHidesOthers(t *testing.T) {
	f := newTrustedDeviceFixture(t)
	current := f.trustDevice(1, time.Now().Add(time.Hour))
	f.trustDevice(1, time.Now().Add(time.Hour))
	f.trustDevice(2, time.Now().Add(time.Hour))

	rr := serve(f.handler.handleListTrustedDevices, newRequest(http.MethodGet, "/auth/devices", nil, current, 1))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var devices []types.TrustedDevice
	if err := json.Unmarshal(rr.Body.Bytes(), &devices); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	currentCount := 0
	for _, device := range devices {
		if device.IsCurrent {
			currentCount++
		}
	}
	if currentCount != 1 {
		t.Fatalf("expected exactly 1 current device, got %d", currentCount)
	}
}

func TestRevokeTrustedDevice(t *testing.T) {
	f := newTrustedDeviceFixture(t)
	f.trustDevice(1, time.Now().Add(time.Hour)) // id 1
	f.trustDevice(2, time.Now().Add(time.Hour)) // id 2

	own := serve(f.handler.handleRevokeTrustedDevice, newRequest(http.MethodDelete, "/auth/devices/1", nil, "", 1))
	if own.Code != http.StatusOK {
		t.Fatalf("expected 200 revoking own device, got %d", own.Code)
	}

	other := serve(f.handler.handleRevokeTrustedDevice, newRequest(http.MethodDelete, "/auth/devices/2", nil, "", 1))
	if other.Code != http.StatusNotFound {
		t.Fatalf("expected 404 revoking another user's device, got %d", other.Code)
	}
	if _, ok := f.devices.devices[2]; !ok {
		t.Fatal("another user's device must survive")
	}

	invalid := serve(f.handler.handleRevokeTrustedDevice, newRequest(http.MethodDelete, "/auth/devices/abc", nil, "", 1))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d", invalid.Code)
	}
}

func TestRevokeAllTrustedDevices(t *testing.T) {
	f := newTrustedDeviceFixture(t)
	f.trustDevice(1, time.Now().Add(time.Hour))
	f.trustDevice(1, time.Now().Add(time.Hour))
	f.trustDevice(2, time.Now().Add(time.Hour))

	rr := serve(f.handler.handleRevokeAllTrustedDevices, newRequest(http.MethodDelete, "/auth/devices", nil, "", 1))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(f.devices.devices) != 1 {
		t.Fatalf("expected only the other user's device left, got %d", len(f.devices.devices))
	}
}

func TestDeviceLabel(t *testing.T) {
	cases := map[string]string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36":              "Chrome on macOS",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36 Edg/130.0":          "Edge on Windows",
		"Mozilla/5.0 (X11; Linux x86_64; rv:132.0) Gecko/20100101 Firefox/132.0":                                                         "Firefox on Linux",
		"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Mobile Safari/537.36":              "Chrome on Android",
		"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1": "Safari on iPad",
		"": "Browser on unknown device",
	}

	for userAgent, expected := range cases {
		if got := deviceLabel(userAgent); got != expected {
			t.Errorf("deviceLabel(%q) = %q, want %q", userAgent, got, expected)
		}
	}
}

func TestIsSecureRequestHonoursForwardedProto(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	if isSecureRequest(req) {
		t.Fatal("plain http request must not be secure")
	}

	req.Header.Set("X-Forwarded-Proto", "https")
	if !isSecureRequest(req) {
		t.Fatal("expected X-Forwarded-Proto=https to count as secure")
	}
}
