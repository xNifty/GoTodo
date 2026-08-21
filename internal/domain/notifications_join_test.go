package domain

import "testing"

func TestJoinRequestNotificationBody(t *testing.T) {
	t.Parallel()
	if got := joinRequestNotificationBody("  a@b.com ", ""); got != "a@b.com" {
		t.Fatalf("empty message: got %q", got)
	}
	if got := joinRequestNotificationBody("a@b.com", "hello"); got != "a@b.com\nhello" {
		t.Fatalf("short message: got %q", got)
	}
	long := make([]byte, joinRequestNotifyPreviewMax+10)
	for i := range long {
		long[i] = 'x'
	}
	got := joinRequestNotificationBody("a@b.com", string(long))
	wantSuffix := "..."
	if len(got) <= len("a@b.com\n") || got[len(got)-3:] != wantSuffix {
		t.Fatalf("long message not truncated: %q", got)
	}
}
