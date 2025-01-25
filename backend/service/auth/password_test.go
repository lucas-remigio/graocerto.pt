package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Errorf("error hashing password %v", err)
	}

	if len(hash) == 0 {
		t.Error("expected hash to not be empty")
	}

	if hash == "password" {
		t.Error("expected hash to not be the same as the password")
	}
}

func TestComparePasswords(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Errorf("error hashing password %v", err)
	}

	if !CheckPasswordHash([]byte("password"), hash) {
		t.Error("expected passwords to match")
	}

	if CheckPasswordHash([]byte("password1"), hash) {
		t.Error("expected passwords to not match")
	}
}
