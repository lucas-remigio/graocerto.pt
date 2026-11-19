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
	AuthTokenPurposeLoginTOTP         AuthTokenPurpose = "login_totp"
	AuthTokenPurposeTOTPRecovery      AuthTokenPurpose = "totp_recovery"
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

// TrustedDeviceStore persists browsers that may skip the second login factor.
// Tokens are looked up by their raw value; implementations store only the hash.
type TrustedDeviceStore interface {
	CreateTrustedDevice(device *TrustedDevice) error
	GetTrustedDeviceByToken(token string) (*TrustedDevice, error)
	ListTrustedDevices(userID int) ([]*TrustedDevice, error)
	TouchTrustedDevice(id int) error
	DeleteTrustedDevice(userID, id int) error
	DeleteTrustedDevicesByUser(userID int) error
}

type TrustedDevice struct {
	ID         int       `json:"id"`
	UserID     int       `json:"-"`
	TokenHash  string    `json:"-"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsCurrent  bool      `json:"is_current"`
}

// TOTPStore persists authenticator-app secrets. Secrets are stored encrypted;
// callers encrypt before writing and decrypt after reading.
type TOTPStore interface {
	GetTOTPState(userID int) (*TOTPState, error)
	SetPendingTOTPSecret(userID int, encryptedSecret string) error
	// EnableTOTP promotes the pending secret and switches the user's MFA method to TOTP.
	EnableTOTP(userID int, lastStep int64) error
	DisableTOTP(userID int) error
	// AdvanceTOTPStep records a used time step; false means it was already used (replay).
	AdvanceTOTPStep(userID int, step int64) (bool, error)
}

type TOTPState struct {
	Secret        string
	PendingSecret string
	LastStep      int64
}

type MfaCodePayload struct {
	// Up to 16 chars so a recovery code (XXXXX-XXXXX) fits as well as a 6-digit code.
	Code string `json:"code" validate:"required,min=6,max=16"`
}

type TOTPStatusResponse struct {
	Enabled bool `json:"enabled"`
}

type TOTPSetupResponse struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

type TOTPEnableResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
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
	ChallengeID    string `json:"challenge_id" validate:"required"`
	Code           string `json:"code" validate:"required,min=6,max=16"`
	RememberDevice bool   `json:"remember_device"`
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
	Method      MfaMethod `json:"method"`
	ExpiresAt   time.Time `json:"expires_at"`
	Message     string    `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
