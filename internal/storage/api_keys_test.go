package storage

import (
	"strings"
	"testing"
)

func TestUpdateAPIKeyNameRejectsBlankAndTooLong(t *testing.T) {
	if _, err := UpdateAPIKeyName(1, 1, "   "); err == nil {
		t.Fatal("whitespace-only name should be rejected")
	}
	if _, err := UpdateAPIKeyName(1, 1, "\u00a0\u00a0"); err == nil {
		t.Fatal("NBSP-only name should be rejected")
	}
	if _, err := UpdateAPIKeyName(1, 1, strings.Repeat("a", 81)); err == nil {
		t.Fatal("name longer than 80 should be rejected")
	}
}
