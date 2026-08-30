package storage

import "testing"

func TestClampEmailAuditRetentionDays(t *testing.T) {
	tests := []struct {
		in   int
		want int
	}{
		{in: 0, want: DefaultEmailAuditRetentionDays},
		{in: -1, want: DefaultEmailAuditRetentionDays},
		{in: 1, want: 1},
		{in: 7, want: 7},
		{in: 90, want: 90},
		{in: 91, want: MaxEmailAuditRetentionDays},
	}
	for _, tt := range tests {
		if got := ClampEmailAuditRetentionDays(tt.in); got != tt.want {
			t.Fatalf("ClampEmailAuditRetentionDays(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestSiteEmailConfigNil(t *testing.T) {
	cfg := SiteEmailConfig(nil)
	if cfg.Provider != "" || cfg.FromAddress != "" {
		t.Fatalf("nil settings should yield empty config: %+v", cfg)
	}
	s := &SiteSettings{Email: cfg}
	s.Email.Provider = EmailProviderSMTP
	if SiteEmailConfig(s).Provider != EmailProviderSMTP {
		t.Fatal("expected smtp provider from settings")
	}
}
