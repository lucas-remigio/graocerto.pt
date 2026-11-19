package user

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

const (
	totpIssuer          = "Grão Certo"
	recoveryCodeCount   = 8
	recoveryCodeLength  = 10
	recoveryCodeTTL     = 10 * 365 * 24 * time.Hour
	loginTOTPTTL        = 5 * time.Minute
	errTOTPNotAvailable = "authenticator app sign-in is not configured"
)

var errTOTPNotConfigured = errors.New(errTOTPNotAvailable)

func totpKey() ([]byte, error) {
	if config.Envs.TOTPEncKey == "" {
		return nil, errTOTPNotConfigured
	}
	return auth.ParseEncryptionKey(config.Envs.TOTPEncKey)
}

// normalizeRecoveryCode lets users type codes with or without the dash, in any case.
func normalizeRecoveryCode(code string) string {
	return strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(code))
}

func (h *Handler) createLoginTOTPChallenge(user *types.User) (*types.AuthChallengeResponse, error) {
	if err := h.authTokenStore.DeleteAuthTokensByUserAndPurpose(user.ID, types.AuthTokenPurposeLoginTOTP); err != nil {
		return nil, err
	}

	challengeID, err := auth.GenerateOneTimeToken()
	if err != nil {
		return nil, err
	}

	// The challenge carries no secret: it only binds the code attempt to a user
	// and reuses auth_tokens' expiry and attempt limiting.
	token := &types.AuthToken{
		ID:          challengeID,
		UserID:      user.ID,
		Purpose:     types.AuthTokenPurposeLoginTOTP,
		ExpiresAt:   time.Now().Add(loginTOTPTTL),
		MaxAttempts: loginOTPAttempts,
	}
	if err := h.authTokenStore.CreateAuthToken(token); err != nil {
		return nil, err
	}

	return &types.AuthChallengeResponse{
		ChallengeID: challengeID,
		Method:      types.MfaMethodTOTP,
		ExpiresAt:   token.ExpiresAt,
		Message:     "enter the code from your authenticator app",
	}, nil
}

// verifyTOTPOrRecovery accepts either the current app code or an unused recovery code.
func (h *Handler) verifyTOTPOrRecovery(userID int, code string) (bool, error) {
	key, err := totpKey()
	if err != nil {
		return false, err
	}

	state, err := h.totp.GetTOTPState(userID)
	if err != nil {
		return false, err
	}
	if state.Secret == "" {
		return false, nil
	}

	secret, err := auth.DecryptSecret(key, state.Secret)
	if err != nil {
		return false, err
	}

	step, err := auth.MatchTOTP(secret, code, time.Now(), state.LastStep)
	if errors.Is(err, auth.ErrTOTPReused) {
		return false, err
	}
	if err != nil {
		return h.redeemRecoveryCode(userID, code)
	}

	advanced, err := h.totp.AdvanceTOTPStep(userID, step)
	if err != nil {
		return false, err
	}
	if !advanced {
		// A concurrent request with the same code won the race.
		return false, auth.ErrTOTPReused
	}
	return true, nil
}

func (h *Handler) redeemRecoveryCode(userID int, code string) (bool, error) {
	token, err := h.authTokenStore.GetAuthTokenByPurposeAndSecret(types.AuthTokenPurposeTOTPRecovery, normalizeRecoveryCode(code))
	if err != nil || token.UserID != userID || h.validateAuthToken(token, types.AuthTokenPurposeTOTPRecovery) != nil {
		return false, nil
	}

	if err := h.authTokenStore.ConsumeAuthToken(token.ID); err != nil {
		return false, err
	}
	return true, nil
}

// issueRecoveryCodes replaces any previous codes; only hashes are stored.
func (h *Handler) issueRecoveryCodes(userID int) ([]string, error) {
	if err := h.authTokenStore.DeleteAuthTokensByUserAndPurpose(userID, types.AuthTokenPurposeTOTPRecovery); err != nil {
		return nil, err
	}

	codes := make([]string, 0, recoveryCodeCount)
	for range recoveryCodeCount {
		code, err := auth.GenerateReadableCode(recoveryCodeLength)
		if err != nil {
			return nil, err
		}

		id, err := auth.GenerateOneTimeToken()
		if err != nil {
			return nil, err
		}

		if err := h.authTokenStore.CreateAuthToken(&types.AuthToken{
			ID:          id,
			UserID:      userID,
			Purpose:     types.AuthTokenPurposeTOTPRecovery,
			SecretHash:  auth.HashSecret(code),
			ExpiresAt:   time.Now().Add(recoveryCodeTTL),
			MaxAttempts: 1,
		}); err != nil {
			return nil, err
		}

		codes = append(codes, code[:5]+"-"+code[5:])
	}
	return codes, nil
}

func (h *Handler) handleTOTPStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	user, err := h.store.GetUserById(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, types.TOTPStatusResponse{Enabled: user.MfaMethod == types.MfaMethodTOTP})
}

func (h *Handler) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	key, err := totpKey()
	if err != nil {
		utils.WriteError(w, http.StatusServiceUnavailable, errTOTPNotConfigured)
		return
	}

	user, err := h.store.GetUserById(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if user.MfaMethod == types.MfaMethodTOTP {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("authenticator app is already enabled"))
		return
	}

	secret, err := h.pendingOrNewTOTPSecret(key, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteDataResponse(w, types.TOTPSetupResponse{
		Secret: secret,
		URI:    auth.TOTPAuthURI(totpIssuer, user.Email, secret),
	})
}

// pendingOrNewTOTPSecret keeps showing the same QR until setup is confirmed, so an
// authenticator entry scanned during an earlier, unfinished attempt keeps working.
func (h *Handler) pendingOrNewTOTPSecret(key []byte, userID int) (string, error) {
	state, err := h.totp.GetTOTPState(userID)
	if err != nil {
		return "", err
	}
	if state.PendingSecret != "" {
		// Undecryptable (e.g. the key changed) just falls through to a fresh secret.
		if secret, err := auth.DecryptSecret(key, state.PendingSecret); err == nil {
			return secret, nil
		}
	}

	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		return "", err
	}
	encrypted, err := auth.EncryptSecret(key, secret)
	if err != nil {
		return "", err
	}
	return secret, h.totp.SetPendingTOTPSecret(userID, encrypted)
}

func (h *Handler) handleTOTPEnable(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	var payload types.MfaCodePayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	step, err := h.matchPendingTOTP(userID, payload.Code)
	if errors.Is(err, errTOTPNotConfigured) {
		utils.WriteError(w, http.StatusServiceUnavailable, err)
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.totp.EnableTOTP(userID, step); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	codes, err := h.issueRecoveryCodes(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.revokeTrustedDevicesAfterMfaChange(r, userID)
	utils.LogWithContext(r.Context()).Info("totp enabled", "userID", userID)
	middleware.WriteDataResponse(w, types.TOTPEnableResponse{RecoveryCodes: codes})
}

// matchPendingTOTP proves the user's app is set up with the pending secret.
func (h *Handler) matchPendingTOTP(userID int, code string) (int64, error) {
	key, err := totpKey()
	if err != nil {
		return 0, err
	}

	state, err := h.totp.GetTOTPState(userID)
	if err != nil {
		return 0, err
	}
	if state.PendingSecret == "" {
		return 0, fmt.Errorf("start the authenticator setup first")
	}

	secret, err := auth.DecryptSecret(key, state.PendingSecret)
	if err != nil {
		return 0, err
	}

	return auth.MatchTOTP(secret, code, time.Now(), 0)
}

func (h *Handler) handleTOTPDisable(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	var payload types.MfaCodePayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	valid, err := h.verifyTOTPOrRecovery(userID, payload.Code)
	if errors.Is(err, errTOTPNotConfigured) {
		utils.WriteError(w, http.StatusServiceUnavailable, err)
		return
	}
	if errors.Is(err, auth.ErrTOTPReused) {
		utils.WriteError(w, http.StatusConflict, err)
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if !valid {
		utils.WriteError(w, http.StatusBadRequest, auth.ErrTOTPInvalid)
		return
	}

	if err := h.totp.DisableTOTP(userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if err := h.authTokenStore.DeleteAuthTokensByUserAndPurpose(userID, types.AuthTokenPurposeTOTPRecovery); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.revokeTrustedDevicesAfterMfaChange(r, userID)
	utils.LogWithContext(r.Context()).Info("totp disabled", "userID", userID)
	middleware.WriteSuccessResponse(w)
}

// revokeTrustedDevicesAfterMfaChange forces every browser through the new
// second factor once. Failure is logged, not fatal: the MFA change already happened.
func (h *Handler) revokeTrustedDevicesAfterMfaChange(r *http.Request, userID int) {
	if h.trustedDevices == nil {
		return
	}
	if err := h.trustedDevices.DeleteTrustedDevicesByUser(userID); err != nil {
		utils.LogWithContext(r.Context()).Error("failed to revoke trusted devices after mfa change", "userID", userID, "error", err)
	}
}
