package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"GoTodo/internal/crypto/secret"
	"GoTodo/internal/mailer"

	"github.com/jackc/pgx/v5"
)

// Email provider values stored in site_settings.email_provider.
const (
	EmailProviderNone    = ""
	EmailProviderMailgun = "mailgun"
	EmailProviderSMTP    = "smtp"
)

// SiteSettings represents site-wide settings stored in the database.
type SiteSettings struct {
	SiteName                 string
	DefaultTimezone          string
	ShowChangelog            bool
	SiteVersion              string
	EnableRegistration       bool
	InviteOnly               bool
	EnableJoinRequests       bool
	MetaDescription          string
	EnableGlobalAnnouncement bool
	GlobalAnnouncementText   string
	EnableAPI                bool

	EmailProvider         string
	EmailFromAddress      string
	EmailFromName         string
	EmailMailgunDomain    string
	EmailMailgunAPIKeyEnc string
	EmailSMTPHost         string
	EmailSMTPPort         int
	EmailSMTPUsername     string
	EmailSMTPPasswordEnc  string
	EmailSMTPTLS          bool

	GitHubOAuthClientID        string
	GitHubOAuthClientSecretEnc string
}

// MailerConfig maps email-related site settings into a mailer.Config.
func (s SiteSettings) MailerConfig() mailer.Config {
	return mailer.Config{
		Provider:         s.EmailProvider,
		FromAddress:      s.EmailFromAddress,
		FromName:         s.EmailFromName,
		MailgunDomain:    s.EmailMailgunDomain,
		MailgunAPIKeyEnc: s.EmailMailgunAPIKeyEnc,
		SMTPHost:         s.EmailSMTPHost,
		SMTPPort:         s.EmailSMTPPort,
		SMTPUsername:     s.EmailSMTPUsername,
		SMTPPasswordEnc:  s.EmailSMTPPasswordEnc,
		SMTPTLS:          s.EmailSMTPTLS,
	}
}

// CreateSiteSettingsTable ensures the site_settings table exists.
func CreateSiteSettingsTable() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	// id is a single-row table; use id=1 for the single settings row
	_, err = pool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS site_settings (
            id INTEGER PRIMARY KEY DEFAULT 1,
            site_name TEXT,
            default_timezone TEXT,
            show_changelog BOOLEAN DEFAULT TRUE,
			site_version TEXT,
			enable_registration BOOLEAN DEFAULT TRUE,
			invite_only BOOLEAN DEFAULT TRUE,
			meta_description TEXT,
			enable_global_announcement BOOLEAN DEFAULT FALSE,
			global_announcement_text TEXT
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create site_settings table: %v", err)
	}
	return nil
}

// GetSiteSettings returns the first (and only) settings row from site_settings.
func GetSiteSettings() (*SiteSettings, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var s SiteSettings
	row := pool.QueryRow(context.Background(), `
		SELECT
			site_name,
			default_timezone,
			show_changelog,
			COALESCE(site_version, ''),
			enable_registration,
			invite_only,
			COALESCE(enable_join_requests, FALSE),
			COALESCE(meta_description, ''),
			COALESCE(enable_global_announcement, FALSE),
			COALESCE(global_announcement_text, ''),
			COALESCE(enable_api, FALSE),
			COALESCE(email_provider, ''),
			COALESCE(email_from_address, ''),
			COALESCE(email_from_name, ''),
			COALESCE(email_mailgun_domain, ''),
			COALESCE(email_mailgun_api_key_enc, ''),
			COALESCE(email_smtp_host, ''),
			COALESCE(email_smtp_port, 587),
			COALESCE(email_smtp_username, ''),
			COALESCE(email_smtp_password_enc, ''),
			COALESCE(email_smtp_tls, TRUE),
			COALESCE(github_oauth_client_id, ''),
			COALESCE(github_oauth_client_secret_enc, '')
		FROM site_settings WHERE id = 1`)
	if err := row.Scan(
		&s.SiteName, &s.DefaultTimezone, &s.ShowChangelog, &s.SiteVersion,
		&s.EnableRegistration, &s.InviteOnly, &s.EnableJoinRequests, &s.MetaDescription,
		&s.EnableGlobalAnnouncement, &s.GlobalAnnouncementText, &s.EnableAPI,
		&s.EmailProvider, &s.EmailFromAddress, &s.EmailFromName,
		&s.EmailMailgunDomain, &s.EmailMailgunAPIKeyEnc,
		&s.EmailSMTPHost, &s.EmailSMTPPort, &s.EmailSMTPUsername,
		&s.EmailSMTPPasswordEnc, &s.EmailSMTPTLS,
		&s.GitHubOAuthClientID, &s.GitHubOAuthClientSecretEnc,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

// UpsertSiteSettings inserts or updates the singleton settings row (id=1).
func UpsertSiteSettings(s SiteSettings) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	if s.EmailSMTPPort <= 0 {
		s.EmailSMTPPort = 587
	}

	_, err = pool.Exec(context.Background(), `
        INSERT INTO site_settings (
			id, site_name, default_timezone, show_changelog, site_version,
			enable_registration, invite_only, enable_join_requests, meta_description,
			enable_global_announcement, global_announcement_text, enable_api,
			email_provider, email_from_address, email_from_name,
			email_mailgun_domain, email_mailgun_api_key_enc,
			email_smtp_host, email_smtp_port, email_smtp_username,
			email_smtp_password_enc, email_smtp_tls,
			github_oauth_client_id, github_oauth_client_secret_enc
		)
        VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
        ON CONFLICT (id) DO UPDATE SET
            site_name = EXCLUDED.site_name,
            default_timezone = EXCLUDED.default_timezone,
            show_changelog = EXCLUDED.show_changelog,
            site_version = EXCLUDED.site_version,
            enable_registration = EXCLUDED.enable_registration,
            invite_only = EXCLUDED.invite_only,
            enable_join_requests = EXCLUDED.enable_join_requests,
            meta_description = EXCLUDED.meta_description,
            enable_global_announcement = EXCLUDED.enable_global_announcement,
            global_announcement_text = EXCLUDED.global_announcement_text,
            enable_api = EXCLUDED.enable_api,
			email_provider = EXCLUDED.email_provider,
			email_from_address = EXCLUDED.email_from_address,
			email_from_name = EXCLUDED.email_from_name,
			email_mailgun_domain = EXCLUDED.email_mailgun_domain,
			email_mailgun_api_key_enc = EXCLUDED.email_mailgun_api_key_enc,
			email_smtp_host = EXCLUDED.email_smtp_host,
			email_smtp_port = EXCLUDED.email_smtp_port,
			email_smtp_username = EXCLUDED.email_smtp_username,
			email_smtp_password_enc = EXCLUDED.email_smtp_password_enc,
			email_smtp_tls = EXCLUDED.email_smtp_tls,
			github_oauth_client_id = EXCLUDED.github_oauth_client_id,
			github_oauth_client_secret_enc = EXCLUDED.github_oauth_client_secret_enc
    `, s.SiteName, s.DefaultTimezone, s.ShowChangelog, s.SiteVersion,
		s.EnableRegistration, s.InviteOnly, s.EnableJoinRequests, s.MetaDescription,
		s.EnableGlobalAnnouncement, s.GlobalAnnouncementText, s.EnableAPI,
		s.EmailProvider, s.EmailFromAddress, s.EmailFromName,
		s.EmailMailgunDomain, s.EmailMailgunAPIKeyEnc,
		s.EmailSMTPHost, s.EmailSMTPPort, s.EmailSMTPUsername,
		s.EmailSMTPPasswordEnc, s.EmailSMTPTLS,
		s.GitHubOAuthClientID, s.GitHubOAuthClientSecretEnc)
	if err != nil {
		return fmt.Errorf("failed to upsert site_settings: %v", err)
	}
	return nil
}

// MigrateSiteSettingsAddRegistrationOptions adds registration settings columns if they don't exist.
func MigrateSiteSettingsAddRegistrationOptions() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	if _, err := pool.Exec(context.Background(), "ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS enable_registration BOOLEAN DEFAULT TRUE"); err != nil {
		return fmt.Errorf("failed to add enable_registration column to site_settings: %v", err)
	}
	if _, err := pool.Exec(context.Background(), "ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS invite_only BOOLEAN DEFAULT TRUE"); err != nil {
		return fmt.Errorf("failed to add invite_only column to site_settings: %v", err)
	}
	return nil
}

// MigrateSiteSettingsAddMetaDescription adds meta_description column if it doesn't exist.
func MigrateSiteSettingsAddMetaDescription() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	if _, err := pool.Exec(context.Background(), "ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS meta_description TEXT"); err != nil {
		return fmt.Errorf("failed to add meta_description column to site_settings: %v", err)
	}
	return nil
}

// MigrateSiteSettingsAddGlobalAnnouncement adds global announcement columns if they don't exist.
func MigrateSiteSettingsAddGlobalAnnouncement() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	if _, err := pool.Exec(context.Background(), "ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS enable_global_announcement BOOLEAN DEFAULT FALSE"); err != nil {
		return fmt.Errorf("failed to add enable_global_announcement column to site_settings: %v", err)
	}
	if _, err := pool.Exec(context.Background(), "ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS global_announcement_text TEXT"); err != nil {
		return fmt.Errorf("failed to add global_announcement_text column to site_settings: %v", err)
	}
	return nil
}

// MigrateSiteSettingsAddEnableAPI adds enable_api column if it doesn't exist.
func MigrateSiteSettingsAddEnableAPI() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	if _, err := pool.Exec(context.Background(), "ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS enable_api BOOLEAN DEFAULT FALSE"); err != nil {
		return fmt.Errorf("failed to add enable_api column to site_settings: %v", err)
	}
	return nil
}

// MigrateSiteSettingsAddEmailSettings adds outbound email configuration columns.
func MigrateSiteSettingsAddEmailSettings() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	alters := []string{
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_provider TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_from_address TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_from_name TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_mailgun_domain TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_mailgun_api_key_enc TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_smtp_host TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_smtp_port INTEGER DEFAULT 587",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_smtp_username TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_smtp_password_enc TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS email_smtp_tls BOOLEAN DEFAULT TRUE",
	}
	for _, q := range alters {
		if _, err := pool.Exec(context.Background(), q); err != nil {
			return fmt.Errorf("failed to migrate site_settings email columns: %v", err)
		}
	}
	return nil
}

// MigrateSiteSettingsAddJoinRequests adds enable_join_requests if missing.
func MigrateSiteSettingsAddJoinRequests() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	if _, err := pool.Exec(context.Background(), "ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS enable_join_requests BOOLEAN DEFAULT FALSE"); err != nil {
		return fmt.Errorf("failed to add enable_join_requests column to site_settings: %v", err)
	}
	return nil
}

// MigrateSiteSettingsAddGitHubOAuth adds GitHub OAuth app credential columns.
func MigrateSiteSettingsAddGitHubOAuth() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	alters := []string{
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS github_oauth_client_id TEXT DEFAULT ''",
		"ALTER TABLE site_settings ADD COLUMN IF NOT EXISTS github_oauth_client_secret_enc TEXT DEFAULT ''",
	}
	for _, q := range alters {
		if _, err := pool.Exec(context.Background(), q); err != nil {
			return fmt.Errorf("failed to migrate site_settings github oauth columns: %v", err)
		}
	}
	return nil
}

// MaybeImportEmailSettingsFromEnv seeds Mailgun settings from legacy env vars when
// the DB has no email provider configured yet. Safe to call repeatedly.
func MaybeImportEmailSettingsFromEnv() error {
	apiKey := strings.TrimSpace(os.Getenv("MAILGUN_API_KEY"))
	domain := strings.TrimSpace(os.Getenv("MAILGUN_DOMAIN"))
	if apiKey == "" || domain == "" {
		return nil
	}

	current, err := GetSiteSettings()
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		current = &SiteSettings{
			SiteName:           "GoTodo",
			DefaultTimezone:    "America/New_York",
			ShowChangelog:      true,
			EnableRegistration: true,
			InviteOnly:         true,
			EmailSMTPPort:      587,
			EmailSMTPTLS:       true,
		}
	}
	if current == nil {
		return nil
	}
	if strings.TrimSpace(current.EmailProvider) != "" && strings.TrimSpace(current.EmailProvider) != "none" {
		return nil
	}
	if current.EmailMailgunAPIKeyEnc != "" {
		return nil
	}

	enc, err := secret.Encrypt(apiKey)
	if err != nil {
		return fmt.Errorf("encrypt mailgun api key from env: %w", err)
	}

	from := strings.TrimSpace(os.Getenv("FROM_EMAIL"))
	if from == "" {
		from = current.EmailFromAddress
	}

	next := *current
	next.EmailProvider = EmailProviderMailgun
	next.EmailMailgunDomain = domain
	next.EmailMailgunAPIKeyEnc = enc
	next.EmailFromAddress = from
	if next.EmailSMTPPort <= 0 {
		next.EmailSMTPPort = 587
	}
	if err := UpsertSiteSettings(next); err != nil {
		return err
	}
	fmt.Println("migration: imported Mailgun email settings from MAILGUN_* / FROM_EMAIL env into site_settings; configure email in Admin going forward (env vars are deprecated)")
	return nil
}
