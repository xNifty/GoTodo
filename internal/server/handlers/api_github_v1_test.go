package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"GoTodo/internal/domain"
)

func TestVerifyGitHubWebhookHMAC(t *testing.T) {
	secret := "whsec_test"
	body := []byte(`{"action":"closed"}`)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !verifyGitHubWebhookHMAC(secret, body, sig) {
		t.Fatal("expected valid signature to verify")
	}
	if verifyGitHubWebhookHMAC(secret, body, "sha256=deadbeef") {
		t.Fatal("expected invalid signature to fail")
	}
	if verifyGitHubWebhookHMAC(secret, body, "md5=abc") {
		t.Fatal("expected wrong prefix to fail")
	}
}

func TestWriteGitHubDomainError(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"validation", domain.ErrValidation, http.StatusBadRequest, "invalid_request"},
		{"not_found", domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"conflict", domain.ErrConflict, http.StatusConflict, "conflict"},
		{"other", errors.New("boom"), http.StatusInternalServerError, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeGitHubDomainError(rec, tc.err)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("json: %v body=%s", err, rec.Body.String())
			}
			if body["error"] != tc.code {
				t.Fatalf("error = %q, want %q", body["error"], tc.code)
			}
		})
	}
}
