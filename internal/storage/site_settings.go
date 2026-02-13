package storage

import (
	"context"
	"fmt"
)

// SiteSettings represents site-wide settings stored in the database.
type SiteSettings struct {
	SiteName           string
	DefaultTimezone    string
	ShowChangelog      bool
	SiteVersion        string
	EnableRegistration bool
	InviteOnly         bool
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
			invite_only BOOLEAN DEFAULT TRUE
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
	row := pool.QueryRow(context.Background(), "SELECT site_name, default_timezone, show_changelog, site_version, enable_registration, invite_only FROM site_settings WHERE id = 1")
	if err := row.Scan(&s.SiteName, &s.DefaultTimezone, &s.ShowChangelog, &s.SiteVersion, &s.EnableRegistration, &s.InviteOnly); err != nil {
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

	_, err = pool.Exec(context.Background(), `
        INSERT INTO site_settings (id, site_name, default_timezone, show_changelog, site_version, enable_registration, invite_only)
        VALUES (1, $1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO UPDATE SET
            site_name = EXCLUDED.site_name,
            default_timezone = EXCLUDED.default_timezone,
            show_changelog = EXCLUDED.show_changelog,
            site_version = EXCLUDED.site_version,
            enable_registration = EXCLUDED.enable_registration,
            invite_only = EXCLUDED.invite_only
    `, s.SiteName, s.DefaultTimezone, s.ShowChangelog, s.SiteVersion, s.EnableRegistration, s.InviteOnly)
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
