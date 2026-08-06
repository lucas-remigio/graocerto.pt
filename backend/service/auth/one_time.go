package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/lucas-remigio/wallet-tracker/utils"
)

func GenerateOneTimeToken() (string, error) {
	return utils.GenerateToken(16)
}

// readableAlphabet omits characters that are easily confused when a human has
// to retype a code: 0/O, 1/I/L.
const readableAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// GenerateReadableCode returns a code a user can type without ambiguity.
// At 8 characters it carries ~39 bits of entropy, which is why callers must
// still expire it and rate limit redemption attempts.
func GenerateReadableCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("code length must be greater than zero")
	}

	max := big.NewInt(int64(len(readableAlphabet)))
	code := make([]byte, length)

	for i := range code {
		index, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = readableAlphabet[index.Int64()]
	}

	return string(code), nil
}

func GenerateOTPCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("otp length must be greater than zero")
	}

	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", length, value.Int64()), nil
}

func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func CompareSecret(secret, expectedHash string) bool {
	return HashSecret(secret) == expectedHash
}
