package auth

import (
	"encoding/base32"
	"errors"
	"strings"
	"testing"
	"time"
)

// rfcSecret is the SHA1 key from RFC 6238 Appendix B ("12345678901234567890").
var rfcSecret = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("12345678901234567890"))

func TestTOTPCodeMatchesRFC6238Vectors(t *testing.T) {
	// The RFC lists 8-digit codes; a 6-digit code is the last 6 digits.
	vectors := map[int64]string{
		59:         "287082",
		1111111109: "081804",
		1111111111: "050471",
		1234567890: "005924",
		2000000000: "279037",
	}

	for unix, want := range vectors {
		got, err := TOTPCode(rfcSecret, TOTPStep(time.Unix(unix, 0)))
		if err != nil {
			t.Fatalf("TOTPCode(%d): %v", unix, err)
		}
		if got != want {
			t.Errorf("TOTPCode at %d = %s, want %s", unix, got, want)
		}
	}
}

func TestMatchTOTPToleratesOneStepOfSkew(t *testing.T) {
	now := time.Unix(1234567890, 0)
	step := TOTPStep(now)

	for offset, wantOK := range map[int64]bool{-2: false, -1: true, 0: true, 1: true, 2: false} {
		code, _ := TOTPCode(rfcSecret, step+offset)
		matched, err := MatchTOTP(rfcSecret, code, now, 0)
		if ok := err == nil; ok != wantOK {
			t.Errorf("offset %d: ok = %v, want %v", offset, ok, wantOK)
		}
		if err == nil && matched != step+offset {
			t.Errorf("offset %d: matched step %d, want %d", offset, matched, step+offset)
		}
	}
}

func TestMatchTOTPRejectsReplayAndGarbage(t *testing.T) {
	now := time.Unix(1234567890, 0)
	step := TOTPStep(now)
	code, _ := TOTPCode(rfcSecret, step)

	if _, err := MatchTOTP(rfcSecret, code, now, step); !errors.Is(err, ErrTOTPReused) {
		t.Errorf("expected ErrTOTPReused for an already-used step, got %v", err)
	}
	// The previous step's code is also already used once a later step was accepted.
	previous, _ := TOTPCode(rfcSecret, step-1)
	if _, err := MatchTOTP(rfcSecret, previous, now, step); !errors.Is(err, ErrTOTPReused) {
		t.Errorf("expected ErrTOTPReused for an older step, got %v", err)
	}
	for _, bad := range []string{"", "12345", "1234567", "abcdef"} {
		if _, err := MatchTOTP(rfcSecret, bad, now, 0); !errors.Is(err, ErrTOTPInvalid) {
			t.Errorf("expected ErrTOTPInvalid for %q, got %v", bad, err)
		}
	}
	if _, err := MatchTOTP("not base32!", code, now, 0); !errors.Is(err, ErrTOTPInvalid) {
		t.Errorf("expected ErrTOTPInvalid for an invalid secret, got %v", err)
	}
}

func TestGenerateTOTPSecretAndURI(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatal(err)
	}
	if len(secret) != 32 {
		t.Fatalf("expected 32 base32 chars for 20 bytes, got %d", len(secret))
	}
	if _, err := TOTPCode(secret, 1); err != nil {
		t.Fatalf("generated secret must be usable: %v", err)
	}

	uri := TOTPAuthURI("Grão Certo", "john@example.com", secret)
	if !strings.HasPrefix(uri, "otpauth://totp/Gr%C3%A3o%20Certo:john@example.com?") {
		t.Fatalf("unexpected uri label: %s", uri)
	}
	if !strings.Contains(uri, "secret="+secret) || !strings.Contains(uri, "issuer=Gr%C3%A3o+Certo") {
		t.Fatalf("missing secret or issuer: %s", uri)
	}
}
