package auth

import (
	"strings"
	"testing"
)

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

func TestGenerateReadableCode(t *testing.T) {
	code, err := GenerateReadableCode(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(code) != 8 {
		t.Fatalf("expected 8 characters, got %q", code)
	}

	for _, r := range code {
		if !strings.ContainsRune(readableAlphabet, r) {
			t.Fatalf("code %q contains ambiguous character %q", code, r)
		}
	}

	for _, ambiguous := range []rune{'0', 'O', '1', 'I', 'L'} {
		if strings.ContainsRune(readableAlphabet, ambiguous) {
			t.Fatalf("alphabet must not contain the easily confused %q", ambiguous)
		}
	}

	if _, err := GenerateReadableCode(0); err == nil {
		t.Fatal("expected an error for a zero length code")
	}
}

func TestGenerateReadableCodeIsRandom(t *testing.T) {
	seen := map[string]bool{}

	for i := 0; i < 50; i++ {
		code, err := GenerateReadableCode(8)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[code] {
			t.Fatalf("generated duplicate code %q", code)
		}
		seen[code] = true
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
