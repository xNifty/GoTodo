package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"GoTodo/internal/config"

	"golang.org/x/crypto/hkdf"
)

const hkdfInfo = "gotodo-email-secrets-v1"

// DeriveKey returns a 32-byte AES key derived from the given root secret via HKDF-SHA256.
func DeriveKey(root string) ([]byte, error) {
	if root == "" {
		return nil, fmt.Errorf("empty root secret")
	}
	r := hkdf.New(sha256.New, []byte(root), nil, []byte(hkdfInfo))
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}
	return key, nil
}

func sessionDerivedKey() ([]byte, error) {
	root, err := config.SessionKey()
	if err != nil {
		return nil, err
	}
	return DeriveKey(root)
}

// Encrypt encrypts plaintext with AES-GCM using a key derived from SESSION_KEY.
// Returns base64(nonce||ciphertext).
func Encrypt(plaintext string) (string, error) {
	key, err := sessionDerivedKey()
	if err != nil {
		return "", err
	}
	return EncryptWithKey(plaintext, key)
}

// Decrypt decrypts a value produced by Encrypt.
func Decrypt(encoded string) (string, error) {
	key, err := sessionDerivedKey()
	if err != nil {
		return "", err
	}
	return DecryptWithKey(encoded, key)
}

// EncryptWithKey encrypts plaintext with the given 32-byte AES key.
func EncryptWithKey(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithKey decrypts a value produced by EncryptWithKey.
func DecryptWithKey(encoded string, key []byte) (string, error) {
	if encoded == "" {
		return "", fmt.Errorf("empty ciphertext")
	}
	if len(key) != 32 {
		return "", fmt.Errorf("key must be 32 bytes")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}
