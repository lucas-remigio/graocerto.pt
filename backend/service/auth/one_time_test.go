package auth

import "testing"

func TestGenerateOTPCode(t *testing.T) {
	code, err := GenerateOTPCode(6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(code) != 6 {
		t.Fatalf("expected 6 digits, got %q", code)
	}

	for _, r := range code {
		if r < '0' || r > '9' {
			t.Fatalf("expected numeric otp, got %q", code)
		}
	}
}

func TestHashSecret(t *testing.T) {
	secret := "hello-world"
	hash := HashSecret(secret)

	if hash == secret {
		t.Fatal("expected hashed secret to differ from the raw secret")
	}

	if !CompareSecret(secret, hash) {
		t.Fatal("expected secret comparison to succeed")
	}

	if CompareSecret("wrong", hash) {
		t.Fatal("expected wrong secret comparison to fail")
	}
}
