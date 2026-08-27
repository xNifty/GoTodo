package mailer

import (
	"strings"
	"testing"
)

func TestSendEmailValidation(t *testing.T) {
	from := "noreply@example.com"
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "empty provider",
			cfg:     Config{FromAddress: from},
			wantErr: "email not configured",
		},
		{
			name:    "none provider",
			cfg:     Config{Provider: "none", FromAddress: from},
			wantErr: "email not configured",
		},
		{
			name:    "whitespace none provider",
			cfg:     Config{Provider: " NONE ", FromAddress: from},
			wantErr: "email not configured",
		},
		{
			name:    "empty from address",
			cfg:     Config{Provider: ProviderSMTP},
			wantErr: "email from address not configured",
		},
		{
			name:    "whitespace from address",
			cfg:     Config{Provider: ProviderMailgun, FromAddress: "  "},
			wantErr: "email from address not configured",
		},
		{
			name:    "unsupported provider",
			cfg:     Config{Provider: "sendgrid", FromAddress: from},
			wantErr: "unsupported email provider",
		},
		{
			name:    "mailgun missing domain",
			cfg:     Config{Provider: ProviderMailgun, FromAddress: from, MailgunAPIKeyEnc: "enc"},
			wantErr: "mailgun credentials not configured",
		},
		{
			name:    "mailgun missing api key",
			cfg:     Config{Provider: ProviderMailgun, FromAddress: from, MailgunDomain: "mg.example.com"},
			wantErr: "mailgun credentials not configured",
		},
		{
			name: "smtp missing host",
			cfg: Config{
				Provider:        ProviderSMTP,
				FromAddress:     from,
				SMTPPort:        587,
				SMTPUsername:    "user",
				SMTPPasswordEnc: "enc",
			},
			wantErr: "smtp credentials not configured",
		},
		{
			name: "smtp missing port",
			cfg: Config{
				Provider:        ProviderSMTP,
				FromAddress:     from,
				SMTPHost:        "smtp.example.com",
				SMTPUsername:    "user",
				SMTPPasswordEnc: "enc",
			},
			wantErr: "smtp credentials not configured",
		},
		{
			name: "smtp missing username",
			cfg: Config{
				Provider:        ProviderSMTP,
				FromAddress:     from,
				SMTPHost:        "smtp.example.com",
				SMTPPort:        587,
				SMTPPasswordEnc: "enc",
			},
			wantErr: "smtp credentials not configured",
		},
		{
			name: "smtp missing password",
			cfg: Config{
				Provider:     ProviderSMTP,
				FromAddress:  from,
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUsername: "user",
			},
			wantErr: "smtp credentials not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendEmail(tt.cfg, "subject", "body", "to@example.com")
			if err == nil {
				t.Fatalf("SendEmail() error = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("SendEmail() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestFormatFrom(t *testing.T) {
	addr := "noreply@example.com"
	tests := []struct {
		name string
		from string
		want string
	}{
		{name: "", from: addr, want: addr},
		{name: "  ", from: addr, want: addr},
		{name: "GoTodo", from: addr, want: "GoTodo <noreply@example.com>"},
	}
	for _, tt := range tests {
		got := formatFrom(tt.name, tt.from)
		if got != tt.want {
			t.Fatalf("formatFrom(%q, %q) = %q, want %q", tt.name, tt.from, got, tt.want)
		}
	}
}
