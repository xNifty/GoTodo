package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"GoTodo/internal/crypto/secret"
	"GoTodo/internal/storage"

	"github.com/mailgun/mailgun-go/v5"
)

// SendEmail sends an email using the admin-configured provider (Mailgun or SMTP).
func SendEmail(subject, message, toEmail string) error {
	settings, err := storage.GetSiteSettings()
	if err != nil || settings == nil {
		return fmt.Errorf("email not configured")
	}

	provider := strings.ToLower(strings.TrimSpace(settings.EmailProvider))
	if provider == "" || provider == "none" {
		return fmt.Errorf("email not configured")
	}

	fromAddr := strings.TrimSpace(settings.EmailFromAddress)
	if fromAddr == "" {
		return fmt.Errorf("email from address not configured")
	}
	from := formatFrom(settings.EmailFromName, fromAddr)

	switch provider {
	case storage.EmailProviderMailgun:
		return sendViaMailgun(settings, subject, message, toEmail, from, fromAddr)
	case storage.EmailProviderSMTP:
		return sendViaSMTP(settings, subject, message, toEmail, from, fromAddr)
	default:
		return fmt.Errorf("unsupported email provider %q", provider)
	}
}

func formatFrom(name, address string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return address
	}
	return fmt.Sprintf("%s <%s>", name, address)
}

func sendViaMailgun(settings *storage.SiteSettings, subject, message, toEmail, from, fromAddr string) error {
	domain := strings.TrimSpace(settings.EmailMailgunDomain)
	if domain == "" || settings.EmailMailgunAPIKeyEnc == "" {
		return fmt.Errorf("mailgun credentials not configured")
	}
	apiKey, err := secret.Decrypt(settings.EmailMailgunAPIKeyEnc)
	if err != nil {
		return fmt.Errorf("decrypt mailgun api key: %w", err)
	}

	mg := mailgun.NewMailgun(apiKey)
	m := mailgun.NewMessage(domain, from, subject, message, toEmail)
	htmlBody := strings.ReplaceAll(message, "\n", "<br/>")
	m.SetHTML(htmlBody)
	m.SetReplyTo(fromAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := mg.Send(ctx, m); err != nil {
		fmt.Println("Email send error:", err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func sendViaSMTP(settings *storage.SiteSettings, subject, message, toEmail, from, fromAddr string) error {
	host := strings.TrimSpace(settings.EmailSMTPHost)
	port := settings.EmailSMTPPort
	username := strings.TrimSpace(settings.EmailSMTPUsername)
	if host == "" || port <= 0 || username == "" || settings.EmailSMTPPasswordEnc == "" {
		return fmt.Errorf("smtp credentials not configured")
	}
	password, err := secret.Decrypt(settings.EmailSMTPPasswordEnc)
	if err != nil {
		return fmt.Errorf("decrypt smtp password: %w", err)
	}

	htmlBody := strings.ReplaceAll(message, "\n", "<br/>")
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + toEmail,
		"Reply-To: " + fromAddr,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	if port == 465 {
		return sendSMTPWithTLS(addr, host, auth, fromAddr, []string{toEmail}, []byte(msg))
	}
	if settings.EmailSMTPTLS {
		return sendSMTPWithStartTLS(addr, host, auth, fromAddr, []string{toEmail}, []byte(msg))
	}
	if err := smtp.SendMail(addr, auth, fromAddr, []string{toEmail}, []byte(msg)); err != nil {
		fmt.Println("Email send error:", err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func sendSMTPWithTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer client.Close()

	return smtpClientSend(client, auth, from, to, msg)
}

func sendSMTPWithStartTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer client.Close()

	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}
	return smtpClientSend(client, auth, from, to, msg)
}

func smtpClientSend(client *smtp.Client, auth smtp.Auth, from string, to []string, msg []byte) error {
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				fmt.Println("Email send error:", err)
				return fmt.Errorf("failed to send email: %w", err)
			}
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return fmt.Errorf("failed to send email: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	if err := client.Quit(); err != nil {
		fmt.Println("Email send error:", err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
