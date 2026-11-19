package auth

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func testKey(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, encryptionKeySize)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := testKey(1)
	sealed, err := EncryptSecret(key, "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatal(err)
	}
	if sealed == "JBSWY3DPEHPK3PXP" {
		t.Fatal("secret must not be stored in plaintext")
	}

	again, _ := EncryptSecret(key, "JBSWY3DPEHPK3PXP")
	if again == sealed {
		t.Fatal("expected a fresh nonce per encryption")
	}

	plain, err := DecryptSecret(key, sealed)
	if err != nil || plain != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("round trip failed: %q, %v", plain, err)
	}
}

func TestDecryptSecretRejectsWrongKeyAndTampering(t *testing.T) {
	sealed, _ := EncryptSecret(testKey(1), "secret")

	if _, err := DecryptSecret(testKey(2), sealed); err == nil {
		t.Error("expected wrong key to fail")
	}

	raw, _ := base64.StdEncoding.DecodeString(sealed)
	raw[len(raw)-1] ^= 0xff
	if _, err := DecryptSecret(testKey(1), base64.StdEncoding.EncodeToString(raw)); err == nil {
		t.Error("expected tampered ciphertext to fail")
	}

	for _, bad := range []string{"", "!!notbase64", "c2hvcnQ="} {
		if _, err := DecryptSecret(testKey(1), bad); err == nil {
			t.Errorf("expected %q to fail", bad)
		}
	}
}

func TestParseEncryptionKey(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString(testKey(7))
	if key, err := ParseEncryptionKey(valid); err != nil || len(key) != encryptionKeySize {
		t.Fatalf("expected valid key, got %v", err)
	}

	short := base64.StdEncoding.EncodeToString([]byte("too short"))
	for _, bad := range []string{"", "not base64!", short} {
		if _, err := ParseEncryptionKey(bad); err == nil {
			t.Errorf("expected %q to be rejected", bad)
		}
	}
}
