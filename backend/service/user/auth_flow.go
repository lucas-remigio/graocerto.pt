package user

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/service/mailer"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

const (
	emailVerificationTTL      = 24 * time.Hour
	loginOTPTTL               = 10 * time.Minute
	passwordResetTTL          = 30 * time.Minute
	emailVerificationCooldown = 5 * time.Minute
	passwordResetCooldown     = 5 * time.Minute
	loginOTPAttempts          = 5
	genericAuthResponse       = "If the request is valid, check your email for next steps."
)

func (h *Handler) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload types.VerifyEmailPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	token, err := h.authTokenStore.GetAuthTokenByPurposeAndSecret(types.AuthTokenPurposeEmailVerification, payload.Token)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid or expired verification token"))
		return
	}

	if err := h.validateAuthToken(token, types.AuthTokenPurposeEmailVerification); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.MarkEmailVerified(token.UserID, true); err != nil {
		utils.LogWithContext(r.Context()).Error("failed to mark email as verified", "userID", token.UserID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.LogWithContext(r.Context()).Info("email verified successfully", "userID", token.UserID)

	if err := h.authTokenStore.ConsumeAuthToken(token.ID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, types.MessageResponse{Message: "email verified"})
}

func (h *Handler) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload types.ResendVerificationPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil || user.EmailVerified {
		utils.WriteJson(w, http.StatusOK, types.MessageResponse{Message: genericAuthResponse})
		return
	}

	if err := h.sendVerificationEmail(user); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, types.MessageResponse{Message: genericAuthResponse})
}

func (h *Handler) handleVerifyLoginOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload types.VerifyLoginOTPPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	token, err := h.authTokenStore.GetAuthTokenByID(payload.ChallengeID)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid or expired login code"))
		return
	}

	if err := h.validateAuthToken(token, types.AuthTokenPurposeLoginOTP); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	if !auth.CompareSecret(payload.Code, token.SecretHash) {
		_ = h.authTokenStore.IncrementAuthTokenAttempts(token.ID)
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid or expired login code"))
		return
	}

	if err := h.authTokenStore.ConsumeAuthToken(token.ID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	user, err := h.store.GetUserById(token.UserID)
	if err != nil {
		utils.LogWithContext(r.Context()).Error("failed to get user after OTP verification", "userID", token.UserID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.LogWithContext(r.Context()).Info("user logged in via OTP", "userID", user.ID, "email", user.Email)
	h.issueAuthToken(w, r, user)
}

func (h *Handler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload types.ForgotPasswordPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteJson(w, http.StatusOK, types.MessageResponse{Message: genericAuthResponse})
		return
	}

	if err := h.sendPasswordResetEmail(user); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, types.MessageResponse{Message: genericAuthResponse})
}

func (h *Handler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload types.ResetPasswordPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	token, err := h.authTokenStore.GetAuthTokenByPurposeAndSecret(types.AuthTokenPurposePasswordReset, payload.Token)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid or expired reset token"))
		return
	}

	if err := h.validateAuthToken(token, types.AuthTokenPurposePasswordReset); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.ValidatePassword(payload.Password); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.store.UpdatePassword(token.UserID, hashedPassword); err != nil {
		utils.LogWithContext(r.Context()).Error("failed to update password", "userID", token.UserID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.LogWithContext(r.Context()).Info("password reset successfully", "userID", token.UserID)

	if err := h.authTokenStore.ConsumeAuthToken(token.ID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusOK, types.MessageResponse{Message: "password updated"})
}

func (h *Handler) createLoginOTPChallenge(user *types.User) (*types.AuthChallengeResponse, error) {
	if h.authTokenStore == nil || h.mailer == nil {
		return nil, fmt.Errorf("mfa dependencies are not configured")
	}

	if err := h.authTokenStore.DeleteAuthTokensByUserAndPurpose(user.ID, types.AuthTokenPurposeLoginOTP); err != nil {
		return nil, err
	}

	challengeID, err := auth.GenerateOneTimeToken()
	if err != nil {
		return nil, err
	}

	code, err := auth.GenerateOTPCode(6)
	if err != nil {
		return nil, err
	}

	token := &types.AuthToken{
		ID:          challengeID,
		UserID:      user.ID,
		Purpose:     types.AuthTokenPurposeLoginOTP,
		SecretHash:  auth.HashSecret(code),
		ExpiresAt:   time.Now().Add(loginOTPTTL),
		Attempts:    0,
		MaxAttempts: loginOTPAttempts,
	}

	if err := h.authTokenStore.CreateAuthToken(token); err != nil {
		return nil, err
	}

	if err := h.mailer.Send(mailer.Message{
		To:      user.Email,
		Subject: "Your Grão Certo login code",
		Body:    fmt.Sprintf("Your login code is %s.\n\nIt expires in 10 minutes.", code),
	}); err != nil {
		return nil, err
	}

	return &types.AuthChallengeResponse{
		ChallengeID: challengeID,
		ExpiresAt:   token.ExpiresAt,
		Message:     "we sent a login code to your email",
	}, nil
}

func (h *Handler) sendVerificationEmail(user *types.User) error {
	return h.sendVerificationEmailWithCooldown(user, emailVerificationCooldown)
}

func (h *Handler) sendVerificationEmailWithCooldown(user *types.User, cooldown time.Duration) error {
	if h.authTokenStore == nil || h.mailer == nil {
		return fmt.Errorf("email verification is not configured")
	}

	shouldSkipSend, err := h.shouldSkipAuthEmail(user.ID, types.AuthTokenPurposeEmailVerification, cooldown)
	if err != nil {
		return err
	}

	if shouldSkipSend {
		return nil
	}

	if err := h.authTokenStore.DeleteAuthTokensByUserAndPurpose(user.ID, types.AuthTokenPurposeEmailVerification); err != nil {
		return err
	}

	rawToken, err := auth.GenerateOneTimeToken()
	if err != nil {
		return err
	}

	token := &types.AuthToken{
		ID:          rawToken,
		UserID:      user.ID,
		Purpose:     types.AuthTokenPurposeEmailVerification,
		SecretHash:  auth.HashSecret(rawToken),
		ExpiresAt:   time.Now().Add(emailVerificationTTL),
		Attempts:    0,
		MaxAttempts: 3,
	}

	if err := h.authTokenStore.CreateAuthToken(token); err != nil {
		return err
	}

	return h.mailer.Send(mailer.Message{
		To:      user.Email,
		Subject: "Verify your Grão Certo email",
		Body: fmt.Sprintf(
			"Please verify your email by opening this link:\n%s/verify-email?token=%s\n\nIf you did not create this account, you can ignore this email.",
			strings.TrimRight(config.Envs.FrontendUrl, "/"),
			rawToken,
		),
	})
}

func (h *Handler) sendPasswordResetEmail(user *types.User) error {
	return h.sendPasswordResetEmailWithCooldown(user, passwordResetCooldown)
}

func (h *Handler) sendPasswordResetEmailWithCooldown(user *types.User, cooldown time.Duration) error {
	if h.authTokenStore == nil || h.mailer == nil {
		return fmt.Errorf("password reset is not configured")
	}

	shouldSkipSend, err := h.shouldSkipAuthEmail(user.ID, types.AuthTokenPurposePasswordReset, cooldown)
	if err != nil {
		return err
	}

	if shouldSkipSend {
		return nil
	}

	if err := h.authTokenStore.DeleteAuthTokensByUserAndPurpose(user.ID, types.AuthTokenPurposePasswordReset); err != nil {
		return err
	}

	rawToken, err := auth.GenerateOneTimeToken()
	if err != nil {
		return err
	}

	token := &types.AuthToken{
		ID:          rawToken,
		UserID:      user.ID,
		Purpose:     types.AuthTokenPurposePasswordReset,
		SecretHash:  auth.HashSecret(rawToken),
		ExpiresAt:   time.Now().Add(passwordResetTTL),
		Attempts:    0,
		MaxAttempts: 3,
	}

	if err := h.authTokenStore.CreateAuthToken(token); err != nil {
		return err
	}

	return h.mailer.Send(mailer.Message{
		To:      user.Email,
		Subject: "Reset your Grão Certo password",
		Body: fmt.Sprintf(
			"Reset your password with this link:\n%s/reset-password?token=%s\n\nIf you did not request this, you can ignore this email.",
			strings.TrimRight(config.Envs.FrontendUrl, "/"),
			rawToken,
		),
	})
}

func (h *Handler) shouldSkipAuthEmail(userID int, purpose types.AuthTokenPurpose, cooldown time.Duration) (bool, error) {
	if h.authTokenStore == nil {
		return false, nil
	}

	token, err := h.authTokenStore.GetLatestAuthTokenByUserAndPurpose(userID, purpose)
	if err != nil {
		if strings.Contains(err.Error(), "auth token not found") {
			return false, nil
		}
		return false, err
	}

	return time.Since(token.CreatedAt) < cooldown, nil
}

func (h *Handler) validateAuthToken(token *types.AuthToken, expectedPurpose types.AuthTokenPurpose) error {
	if token.Purpose != expectedPurpose {
		return fmt.Errorf("invalid or expired token")
	}

	if token.ConsumedAt != nil {
		return fmt.Errorf("invalid or expired token")
	}

	if time.Now().After(token.ExpiresAt) {
		return fmt.Errorf("invalid or expired token")
	}

	if token.Attempts >= token.MaxAttempts {
		return fmt.Errorf("invalid or expired token")
	}

	return nil
}
