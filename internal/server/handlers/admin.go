package handlers

import (
	"GoTodo/internal/config"
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
	"GoTodo/internal/version"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// AdminPageHandler shows the admin settings page
func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	_, _, permissions, _, loggedIn, _ := utils.GetSessionUserWithTimezone(r)
	if !loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Prefer DB-backed settings for mutable fields; site version is always baked into the binary
	siteName := config.Cfg.SiteName
	siteVersion := version.Version
	defaultTz := config.Cfg.DefaultTimezone
	showChangelog := config.Cfg.ShowChangelog
	if s, err := storage.GetSiteSettings(); err == nil && s != nil {
		if s.SiteName != "" {
			siteName = s.SiteName
		}
		if s.DefaultTimezone != "" {
			defaultTz = s.DefaultTimezone
		}
		showChangelog = s.ShowChangelog
	}

	context := map[string]interface{}{
		"LoggedIn":        loggedIn,
		"Permissions":     permissions,
		"SiteName":        siteName,
		"SiteVersion":     siteVersion,
		"DefaultTimezone": defaultTz,
		"ShowChangelog":   showChangelog,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := utils.RenderTemplate(w, r, "admin.html", context); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

// APIUpdateSiteSettings updates site-wide settings (only for admins)
func APIUpdateSiteSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// We expect the route to be protected by RequirePermission("admin", ...)

	siteName := strings.TrimSpace(r.FormValue("site_name"))
	defaultTz := strings.TrimSpace(r.FormValue("default_timezone"))
	showChangelogStr := r.FormValue("show_changelog")

	if siteName == "" {
		http.Error(w, "Site name is required", http.StatusBadRequest)
		return
	}
	if defaultTz == "" {
		http.Error(w, "Default timezone is required", http.StatusBadRequest)
		return
	}

	// Update in-memory config
	config.Cfg.SiteName = siteName
	config.Cfg.DefaultTimezone = defaultTz
	if showChangelogStr == "true" || showChangelogStr == "on" {
		config.Cfg.ShowChangelog = true
	} else {
		config.Cfg.ShowChangelog = false
	}

	// Persist to DB when possible; fall back to config file if DB unavailable
	ss := storage.SiteSettings{
		SiteName:        siteName,
		DefaultTimezone: defaultTz,
		ShowChangelog:   config.Cfg.ShowChangelog,
		// Do NOT persist site version from the app; site version is baked into the binary only.
		SiteVersion: "",
	}
	if err := storage.UpsertSiteSettings(ss); err != nil {
		// fallback: persist to config file
		out, err := json.MarshalIndent(config.Cfg, "", "  ")
		if err != nil {
			http.Error(w, "Failed to encode config: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile("config/config.json", out, 0644); err != nil {
			http.Error(w, "Failed to write config file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect back to admin page
	http.Redirect(w, r, utils.GetBasePath()+"/admin?status=success", http.StatusSeeOther)
}

// Note: bumping site version is intentionally disabled from within the site.
