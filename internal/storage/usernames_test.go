package storage

import "testing"

func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"alice", "alice"},
		{"a_b", `a\_b`},
		{"100%", `100\%`},
		{`path\to`, `path\\to`},
		{`_%\`, `\_\%\\`},
	}
	for _, tt := range tests {
		if got := EscapeLikePattern(tt.in); got != tt.want {
			t.Errorf("EscapeLikePattern(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
