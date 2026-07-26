package secret

import (
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, err := DeriveKey("test-session-key-for-unit-tests-32chars!!")
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	plain := "mailgun-api-key-secret-value"
	enc, err := EncryptWithKey(plain, key)
	if err != nil {
		t.Fatalf("EncryptWithKey: %v", err)
	}
	if enc == "" || enc == plain {
		t.Fatalf("expected non-empty ciphertext distinct from plaintext")
	}
	got, err := DecryptWithKey(enc, key)
	if err != nil {
		t.Fatalf("DecryptWithKey: %v", err)
	}
	if got != plain {
		t.Fatalf("got %q, want %q", got, plain)
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	key1, err := DeriveKey("test-session-key-for-unit-tests-32chars!!")
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	key2, err := DeriveKey("different-session-key-at-least-32chars!!")
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	enc, err := EncryptWithKey("secret", key1)
	if err != nil {
		t.Fatalf("EncryptWithKey: %v", err)
	}
	if _, err := DecryptWithKey(enc, key2); err == nil {
		t.Fatal("expected decrypt with wrong key to fail")
	}
}

func TestEncryptWithSessionKey(t *testing.T) {
	enc, err := Encrypt("hello-world")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	got, err := Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != "hello-world" {
		t.Fatalf("got %q, want hello-world", got)
	}
}
