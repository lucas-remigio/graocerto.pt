package types

import "time"

type MfaMethod string

const (
	MfaMethodEmailOTP MfaMethod = "email_otp"
	MfaMethodTOTP     MfaMethod = "totp"
	MfaMethodOff      MfaMethod = "off"
)

type AuthTokenPurpose string

const (
	AuthTokenPurposeEmailVerification AuthTokenPurpose = "email_verification"
	AuthTokenPurposeLoginOTP          AuthTokenPurpose = "login_otp"
	AuthTokenPurposePasswordReset     AuthTokenPurpose = "password_reset"
	AuthTokenPurposeTelegramLink      AuthTokenPurpose = "telegram_link"
)

type AuthTokenStore interface {
	CreateAuthToken(token *AuthToken) error
	GetAuthTokenByID(id string) (*AuthToken, error)
	GetAuthTokenByPurposeAndSecret(purpose AuthTokenPurpose, secret string) (*AuthToken, error)
	GetLatestAuthTokenByUserAndPurpose(userID int, purpose AuthTokenPurpose) (*AuthToken, error)
	ConsumeAuthToken(id string) error
	IncrementAuthTokenAttempts(id string) error
	DeleteAuthTokensByUserAndPurpose(userID int, purpose AuthTokenPurpose) error
}

type AuthToken struct {
	ID          string           `json:"id"`
	UserID      int              `json:"user_id"`
	Purpose     AuthTokenPurpose `json:"purpose"`
	SecretHash  string           `json:"-"`
	ExpiresAt   time.Time        `json:"expires_at"`
	ConsumedAt  *time.Time       `json:"consumed_at,omitempty"`
	Attempts    int              `json:"attempts"`
	MaxAttempts int              `json:"max_attempts"`
	CreatedAt   time.Time        `json:"created_at"`
}

type RegisterUserPayload struct {
	FirstName string `json:"first_name" validate:"required,max=32"`
	LastName  string `json:"last_name" validate:"required,max=32"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=64"`
}

type GoogleLoginPayload struct {
	Token string `json:"token" validate:"required"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type VerifyEmailPayload struct {
	Token string `json:"token" validate:"required"`
}

type ResendVerificationPayload struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type VerifyLoginOTPPayload struct {
	ChallengeID string `json:"challenge_id" validate:"required"`
	Code        string `json:"code" validate:"required,min=6,max=6"`
}

type ForgotPasswordPayload struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type ResetPasswordPayload struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type AuthChallengeResponse struct {
	ChallengeID string    `json:"challenge_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	Message     string    `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
