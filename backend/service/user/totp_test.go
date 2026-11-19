package user

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type memoryTOTPStore struct {
	users  *memoryUserStore
	states map[int]*types.TOTPState
}

func (s *memoryTOTPStore) state(userID int) *types.TOTPState {
	if s.states[userID] == nil {
		s.states[userID] = &types.TOTPState{}
	}
	return s.states[userID]
}

func (s *memoryTOTPStore) GetTOTPState(userID int) (*types.TOTPState, error) {
	copyState := *s.state(userID)
	return &copyState, nil
}

func (s *memoryTOTPStore) SetPendingTOTPSecret(userID int, encryptedSecret string) error {
	s.state(userID).PendingSecret = encryptedSecret
	return nil
}

func (s *memoryTOTPStore) EnableTOTP(userID int, lastStep int64) error {
	state := s.state(userID)
	if state.PendingSecret == "" {
		return fmt.Errorf("no pending secret")
	}
	state.Secret, state.PendingSecret, state.LastStep = state.PendingSecret, "", lastStep
	return s.users.UpdateMfaMethod(userID, types.MfaMethodTOTP)
}

func (s *memoryTOTPStore) DisableTOTP(userID int) error {
	s.states[userID] = &types.TOTPState{}
	return s.users.UpdateMfaMethod(userID, types.MfaMethodEmailOTP)
}

func (s *memoryTOTPStore) AdvanceTOTPStep(userID int, step int64) (bool, error) {
	state := s.state(userID)
	if state.LastStep >= step {
		return false, nil
	}
	state.LastStep = step
	return true, nil
}

type totpFixture struct {
	handler   *Handler
	users     *memoryUserStore
	authStore *memoryAuthTokenStore
	totp      *memoryTOTPStore
	devices   *memoryTrustedDeviceStore
	mailer    *memoryMailer
}

// newTOTPFixture creates one verified email-OTP user (id 1) and a valid TOTP_ENC_KEY.
func newTOTPFixture(t *testing.T) *totpFixture {
	t.Helper()
	previousKey := config.Envs.TOTPEncKey
	config.Envs.TOTPEncKey = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32))
	t.Cleanup(func() { config.Envs.TOTPEncKey = previousKey })

	users := newMemoryUserStore()
	hashedPassword, _ := auth.HashPassword("Password1!")
	_ = users.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      hashedPassword,
		EmailVerified: true,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	f := &totpFixture{
		users:     users,
		authStore: newMemoryAuthTokenStore(),
		totp:      &memoryTOTPStore{users: users, states: map[int]*types.TOTPState{}},
		devices:   newMemoryTrustedDeviceStore(),
		mailer:    &memoryMailer{},
	}
	f.handler = NewHandler(users, f.authStore, f.devices, f.totp, f.mailer, nil, nil, nil)
	return f
}

func (f *totpFixture) setup(t *testing.T) types.TOTPSetupResponse {
	t.Helper()
	rr := serve(f.handler.handleTOTPSetup, newRequest(http.MethodPost, "/auth/totp/setup", nil, "", 1))
	if rr.Code != http.StatusOK {
		t.Fatalf("setup: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var response types.TOTPSetupResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	return response
}

// enable runs setup + enable and returns the secret and recovery codes.
func (f *totpFixture) enable(t *testing.T) (string, []string) {
	t.Helper()
	secret := f.setup(t).Secret
	// Enable with the previous step's code so later tests can use the current one
	// without tripping replay protection.
	code, _ := auth.TOTPCode(secret, auth.TOTPStep(time.Now())-1)
	rr := serve(f.handler.handleTOTPEnable, newRequest(http.MethodPost, "/auth/totp/enable", types.MfaCodePayload{Code: code}, "", 1))
	if rr.Code != http.StatusOK {
		t.Fatalf("enable: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var response types.TOTPEnableResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	return secret, response.RecoveryCodes
}

func (f *totpFixture) login(t *testing.T) types.AuthChallengeResponse {
	t.Helper()
	rr := serve(f.handler.handleLogin, loginRequest("Password1!", ""))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("login: expected 202, got %d: %s", rr.Code, rr.Body.String())
	}
	var challenge types.AuthChallengeResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &challenge)
	return challenge
}

func (f *totpFixture) verify(challengeID, code string) int {
	rr := serve(f.handler.handleVerifyLoginOTP, newRequest(http.MethodPost, "/auth/verify-login-otp", types.VerifyLoginOTPPayload{
		ChallengeID: challengeID,
		Code:        code,
	}, "", 0))
	return rr.Code
}

func currentCode(secret string) string {
	code, _ := auth.TOTPCode(secret, auth.TOTPStep(time.Now()))
	return code
}

func TestTOTPSetupStoresEncryptedPendingSecret(t *testing.T) {
	f := newTOTPFixture(t)
	response := f.setup(t)

	if response.Secret == "" || !strings.HasPrefix(response.URI, "otpauth://totp/") {
		t.Fatalf("unexpected setup response: %+v", response)
	}
	state, _ := f.totp.GetTOTPState(1)
	if state.PendingSecret == "" || state.PendingSecret == response.Secret {
		t.Fatal("pending secret must be stored, and encrypted")
	}
	if user, _ := f.users.GetUserById(1); user.MfaMethod != types.MfaMethodEmailOTP {
		t.Fatal("setup alone must not switch the MFA method")
	}
}

func TestTOTPSetupRejectsWhenUnavailable(t *testing.T) {
	t.Run("no encryption key", func(t *testing.T) {
		f := newTOTPFixture(t)
		config.Envs.TOTPEncKey = ""
		rr := serve(f.handler.handleTOTPSetup, newRequest(http.MethodPost, "/auth/totp/setup", nil, "", 1))
		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", rr.Code)
		}
	})

	t.Run("already enabled", func(t *testing.T) {
		f := newTOTPFixture(t)
		f.enable(t)
		rr := serve(f.handler.handleTOTPSetup, newRequest(http.MethodPost, "/auth/totp/setup", nil, "", 1))
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rr.Code)
		}
	})
}

func TestTOTPEnableRequiresValidCodeFromPendingSecret(t *testing.T) {
	f := newTOTPFixture(t)

	noSetup := serve(f.handler.handleTOTPEnable, newRequest(http.MethodPost, "/auth/totp/enable", types.MfaCodePayload{Code: "123456"}, "", 1))
	if noSetup.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without setup, got %d", noSetup.Code)
	}

	secret := f.setup(t).Secret
	wrong := "000000"
	if currentCode(secret) == wrong {
		wrong = "111111"
	}
	rr := serve(f.handler.handleTOTPEnable, newRequest(http.MethodPost, "/auth/totp/enable", types.MfaCodePayload{Code: wrong}, "", 1))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for wrong code, got %d", rr.Code)
	}
	if user, _ := f.users.GetUserById(1); user.MfaMethod != types.MfaMethodEmailOTP {
		t.Fatal("a wrong code must not switch the MFA method")
	}
}

func TestTOTPEnableSwitchesMethodIssuesCodesAndRevokesDevices(t *testing.T) {
	f := newTOTPFixture(t)
	_ = f.devices.CreateTrustedDevice(&types.TrustedDevice{UserID: 1, TokenHash: "h", ExpiresAt: time.Now().Add(time.Hour)})

	_, codes := f.enable(t)

	if user, _ := f.users.GetUserById(1); user.MfaMethod != types.MfaMethodTOTP {
		t.Fatalf("expected totp method, got %s", user.MfaMethod)
	}
	if len(codes) != recoveryCodeCount {
		t.Fatalf("expected %d recovery codes, got %d", recoveryCodeCount, len(codes))
	}
	if len(codes[0]) != 11 || codes[0][5] != '-' {
		t.Fatalf("expected XXXXX-XXXXX format, got %q", codes[0])
	}
	if len(f.devices.devices) != 0 {
		t.Fatal("expected trusted devices to be revoked after enabling TOTP")
	}
}

func TestLoginWithTOTPUsesAppCodeNotEmail(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enable(t)

	challenge := f.login(t)
	if challenge.Method != types.MfaMethodTOTP {
		t.Fatalf("expected totp challenge, got %s", challenge.Method)
	}
	if len(f.mailer.messages) != 0 {
		t.Fatal("a TOTP login must not send an email")
	}

	if code := f.verify(challenge.ChallengeID, currentCode(secret)); code != http.StatusOK {
		t.Fatalf("expected 200 for valid app code, got %d", code)
	}
}

func TestLoginWithTOTPRejectsReplayedCode(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enable(t)
	code := currentCode(secret)

	if status := f.verify(f.login(t).ChallengeID, code); status != http.StatusOK {
		t.Fatalf("first use: expected 200, got %d", status)
	}
	if status := f.verify(f.login(t).ChallengeID, code); status != http.StatusConflict {
		t.Fatalf("replay: expected 409 (already used), got %d", status)
	}
}

func TestTOTPDisableWithJustUsedCodeSaysAlreadyUsed(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enable(t)
	code := currentCode(secret)
	if status := f.verify(f.login(t).ChallengeID, code); status != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", status)
	}

	rr := serve(f.handler.handleTOTPDisable, newRequest(http.MethodPost, "/auth/totp/disable", types.MfaCodePayload{Code: code}, "", 1))
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 for the code still on screen, got %d", rr.Code)
	}
	if user, _ := f.users.GetUserById(1); user.MfaMethod != types.MfaMethodTOTP {
		t.Fatal("a reused code must not disable TOTP")
	}
}

func TestTOTPSetupReusesPendingSecretUntilConfirmed(t *testing.T) {
	f := newTOTPFixture(t)
	first := f.setup(t)
	if again := f.setup(t); again.Secret != first.Secret || again.URI != first.URI {
		t.Fatal("reopening setup must show the same QR until it is confirmed")
	}

	// After enabling and disabling, a new setup must not resurrect the old secret.
	secret, _ := f.enable(t)
	rr := serve(f.handler.handleTOTPDisable, newRequest(http.MethodPost, "/auth/totp/disable", types.MfaCodePayload{Code: currentCode(secret)}, "", 1))
	if rr.Code != http.StatusOK {
		t.Fatalf("disable: expected 200, got %d", rr.Code)
	}
	if fresh := f.setup(t); fresh.Secret == first.Secret {
		t.Fatal("expected a fresh secret after disabling")
	}
}

func TestLoginWithTOTPLocksChallengeAfterTooManyAttempts(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enable(t)
	challenge := f.login(t)

	wrong := "000000"
	if currentCode(secret) == wrong {
		wrong = "111111"
	}
	for range loginOTPAttempts {
		f.verify(challenge.ChallengeID, wrong)
	}

	if status := f.verify(challenge.ChallengeID, currentCode(secret)); status != http.StatusUnauthorized {
		t.Fatalf("expected locked challenge to reject even a valid code, got %d", status)
	}
}

func TestRecoveryCodeWorksOnceInAnyFormat(t *testing.T) {
	f := newTOTPFixture(t)
	_, codes := f.enable(t)
	sloppy := strings.ToLower(strings.ReplaceAll(codes[0], "-", ""))

	if status := f.verify(f.login(t).ChallengeID, sloppy); status != http.StatusOK {
		t.Fatalf("expected recovery code to log in, got %d", status)
	}
	if status := f.verify(f.login(t).ChallengeID, codes[0]); status != http.StatusUnauthorized {
		t.Fatalf("expected used recovery code to be rejected, got %d", status)
	}
}

func TestTrustedDeviceSkipsTOTP(t *testing.T) {
	f := newTOTPFixture(t)
	f.enable(t)
	_ = f.devices.CreateTrustedDevice(&types.TrustedDevice{
		UserID:    1,
		TokenHash: auth.HashSecret("device-token"),
		ExpiresAt: time.Now().Add(time.Hour),
	})

	rr := serve(f.handler.handleLogin, loginRequest("Password1!", "device-token"))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected trusted device to skip TOTP, got %d", rr.Code)
	}
}

func TestTOTPDisableRequiresCodeAndRestoresEmail(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enable(t)

	wrong := "000000"
	if currentCode(secret) == wrong {
		wrong = "111111"
	}
	rejected := serve(f.handler.handleTOTPDisable, newRequest(http.MethodPost, "/auth/totp/disable", types.MfaCodePayload{Code: wrong}, "", 1))
	if rejected.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for wrong code, got %d", rejected.Code)
	}

	rr := serve(f.handler.handleTOTPDisable, newRequest(http.MethodPost, "/auth/totp/disable", types.MfaCodePayload{Code: currentCode(secret)}, "", 1))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if user, _ := f.users.GetUserById(1); user.MfaMethod != types.MfaMethodEmailOTP {
		t.Fatal("expected email OTP after disabling")
	}
	for _, token := range f.authStore.tokens {
		if token.Purpose == types.AuthTokenPurposeTOTPRecovery && token.ConsumedAt == nil {
			t.Fatal("expected recovery codes to be deleted")
		}
	}

	challenge := f.login(t)
	if challenge.Method != types.MfaMethodEmailOTP || len(f.mailer.messages) != 1 {
		t.Fatal("expected login to be back on email codes")
	}
}

func TestTOTPStatus(t *testing.T) {
	f := newTOTPFixture(t)
	status := func() bool {
		rr := serve(f.handler.handleTOTPStatus, newRequest(http.MethodGet, "/auth/totp", nil, "", 1))
		var response types.TOTPStatusResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &response)
		return response.Enabled
	}

	if status() {
		t.Fatal("expected disabled by default")
	}
	f.enable(t)
	if !status() {
		t.Fatal("expected enabled after enable")
	}
}
