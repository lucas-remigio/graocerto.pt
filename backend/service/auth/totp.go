package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// RFC 6238 defaults; every mainstream authenticator app (Ente Auth, Aegis,
// Google Authenticator, 1Password) assumes exactly these.
const (
	totpPeriod     = 30
	totpDigits     = 6
	totpSecretSize = 20
	// totpSkewSteps accepts the previous and next code to tolerate phone clock drift.
	totpSkewSteps = 1
)

var totpEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

var (
	ErrTOTPInvalid = errors.New("invalid authenticator code")
	// ErrTOTPReused means the code is right but its time step was already used,
	// typically the code still on screen right after a successful login.
	ErrTOTPReused = errors.New("authenticator code already used")
)

func GenerateTOTPSecret() (string, error) {
	secret := make([]byte, totpSecretSize)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("generate totp secret: %w", err)
	}
	return totpEncoding.EncodeToString(secret), nil
}

// TOTPAuthURI is the otpauth:// link that authenticator apps read from the QR code.
func TOTPAuthURI(issuer, account, secret string) string {
	label := url.PathEscape(issuer + ":" + account)
	query := url.Values{"secret": {secret}, "issuer": {issuer}}
	return "otpauth://totp/" + label + "?" + query.Encode()
}

func TOTPStep(now time.Time) int64 {
	return now.Unix() / totpPeriod
}

func TOTPCode(secret string, step int64) (string, error) {
	key, err := totpEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("decode totp secret: %w", err)
	}

	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(step))
	mac := hmac.New(sha1.New, key)
	mac.Write(counter[:])
	sum := mac.Sum(nil)

	// Dynamic truncation, RFC 4226 section 5.3.
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", value%1_000_000), nil
}

// MatchTOTP returns the time step the code belongs to. Steps at or before
// lastUsedStep return ErrTOTPReused so an observed code can't be replayed.
func MatchTOTP(secret, code string, now time.Time, lastUsedStep int64) (int64, error) {
	step, ok := findTOTPStep(secret, code, now)
	if !ok {
		return 0, ErrTOTPInvalid
	}
	if step <= lastUsedStep {
		return 0, ErrTOTPReused
	}
	return step, nil
}

// findTOTPStep checks the current step and one either side for clock drift.
func findTOTPStep(secret, code string, now time.Time) (int64, bool) {
	if len(code) != totpDigits {
		return 0, false
	}

	current := TOTPStep(now)
	for step := current - totpSkewSteps; step <= current+totpSkewSteps; step++ {
		expected, err := TOTPCode(secret, step)
		if err != nil {
			return 0, false
		}
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return step, true
		}
	}
	return 0, false
}
