package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"GoTodo/internal/storage"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func TestNormalizeAndFormatRecoveryCode(t *testing.T) {
	if got := NormalizeRecoveryCode(" ab12c-de3fg "); got != "AB12CDE3FG" {
		t.Fatalf("normalize = %q", got)
	}
	if got := FormatRecoveryCode("AB12CDE3FG"); got != "AB12C-DE3FG" {
		t.Fatalf("format = %q", got)
	}
}

func TestGenerateAndMatchRecoveryCodes(t *testing.T) {
	codes, err := GenerateRecoveryCodes(5)
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 5 {
		t.Fatalf("len=%d", len(codes))
	}
	seen := map[string]struct{}{}
	var hashes []storage.RecoveryCodeHash
	for i, c := range codes {
		norm := NormalizeRecoveryCode(c)
		if len(norm) != 10 {
			t.Fatalf("code %q not 10 chars normalized", c)
		}
		if strings.Count(c, "-") != 1 {
			t.Fatalf("code %q missing dash", c)
		}
		if _, ok := seen[norm]; ok {
			t.Fatalf("duplicate %q", c)
		}
		seen[norm] = struct{}{}
		h, err := HashRecoveryCode(c)
		if err != nil {
			t.Fatal(err)
		}
		hashes = append(hashes, storage.RecoveryCodeHash{ID: i + 1, Hash: h})
	}
	id, ok := MatchRecoveryCode(codes[2], hashes)
	if !ok || id != 3 {
		t.Fatalf("match id=%d ok=%v", id, ok)
	}
	if _, ok := MatchRecoveryCode("ZZZZZ-ZZZZZ", hashes); ok {
		t.Fatal("expected no match")
	}
}

func TestVerifyTOTPWindowAndReplay(t *testing.T) {
	setup, err := GenerateTOTPSetup("GoTodo", "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	opts := totp.ValidateOpts{
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
	code, err := totp.GenerateCodeCustom(setup.Secret, now, opts)
	if err != nil {
		t.Fatal(err)
	}
	step, ok := VerifyTOTP(setup.Secret, code, 0, now)
	if !ok {
		t.Fatal("expected valid TOTP")
	}
	if _, ok := VerifyTOTP(setup.Secret, code, step, now); ok {
		t.Fatal("replay should fail")
	}
	if _, ok := VerifyTOTP(setup.Secret, "000000", 0, now); ok {
		t.Fatal("wrong code should fail")
	}
	prevCode, err := totp.GenerateCodeCustom(setup.Secret, now.Add(-30*time.Second), opts)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := VerifyTOTP(setup.Secret, prevCode, 0, now); !ok {
		t.Fatal("expected previous period to be accepted (skew)")
	}
}

func TestEnableDisableMFARoundTrip(t *testing.T) {
	const userID = 1
	_ = storage.DisableUserMFA(userID)

	setup, err := GenerateTOTPSetup("GoTodo", "owner@example.com")
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCodeCustom(setup.Secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EnableMFA(userID, setup.Secret, "000000"); !errors.Is(err, ErrValidation) {
		t.Fatalf("bad code: %v", err)
	}
	codes, err := EnableMFA(userID, setup.Secret, code)
	if err != nil {
		t.Fatalf("enable: %v", err)
	}
	if len(codes) != RecoveryCodeCount {
		t.Fatalf("codes=%d", len(codes))
	}
	status, err := GetMFAStatus(userID)
	if err != nil || !status.Enabled || status.RecoveryCodesRemaining != 5 {
		t.Fatalf("status=%+v err=%v", status, err)
	}
	if err := VerifyLoginMFA(userID, codes[0]); err != nil {
		t.Fatalf("recovery login: %v", err)
	}
	if err := VerifyLoginMFA(userID, codes[0]); !errors.Is(err, ErrValidation) {
		t.Fatalf("reuse recovery: %v", err)
	}
	status, err = GetMFAStatus(userID)
	if err != nil || status.RecoveryCodesRemaining != 4 {
		t.Fatalf("remaining=%+v err=%v", status, err)
	}
	if err := DisableMFA(userID, codes[1]); err != nil {
		t.Fatalf("disable: %v", err)
	}
	status, err = GetMFAStatus(userID)
	if err != nil || status.Enabled {
		t.Fatalf("after disable %+v err=%v", status, err)
	}
}

func TestDisableMFARequiresCode(t *testing.T) {
	err := DisableMFA(1, "")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err=%v", err)
	}
}
