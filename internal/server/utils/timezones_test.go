package utils

import "testing"

func TestIsValidTimezone(t *testing.T) {
	if !IsValidTimezone("America/New_York") {
		t.Fatal("expected America/New_York valid")
	}
	if !IsValidTimezone("America/Detroit") {
		t.Fatal("expected America/Detroit valid (outside former allowlist)")
	}
	if !IsValidTimezone("UTC") {
		t.Fatal("expected UTC valid")
	}
	if IsValidTimezone("Invalid/Zone") {
		t.Fatal("expected Invalid/Zone invalid")
	}
	if IsValidTimezone("") {
		t.Fatal("expected empty invalid")
	}
	if IsValidTimezone("   ") {
		t.Fatal("expected whitespace-only invalid")
	}
}

func TestValidItemsPerPage(t *testing.T) {
	if !ValidItemsPerPage(25) {
		t.Fatal("25 should be valid")
	}
	if ValidItemsPerPage(99) {
		t.Fatal("99 should be invalid")
	}
}
