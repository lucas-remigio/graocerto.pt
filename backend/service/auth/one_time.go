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
