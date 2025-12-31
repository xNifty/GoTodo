package utils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	cfg "GoTodo/internal/config"

	"github.com/mailgun/mailgun-go/v5"
)

// SendEmail sends an email using Mailgun. Parameters: subject, message, toEmail
func SendEmail(subject, message, toEmail string) error {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")
	if apiKey == "" || domain == "" {
		return fmt.Errorf("mailgun credentials not configured")
	}

	mg := mailgun.NewMailgun(apiKey)
	from := cfg.Cfg.FromEmail
	if from == "" {
		from = "no-reply@ryanmalacina.com"
	}

	// In mailgun-go v5 the domain is provided to NewMessage
	m := mailgun.NewMessage(
		domain,
		from,
		subject,
		message,
		toEmail,
	)
	// Preserve line breaks for HTML by converting newlines to <br/>
	htmlBody := strings.ReplaceAll(message, "\n", "<br/>")
	m.SetHTML(htmlBody)
	m.SetReplyTo(from)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err := mg.Send(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
