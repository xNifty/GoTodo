package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"GoTodo/internal/config"
	"GoTodo/internal/crypto/secret"
	"GoTodo/internal/domain"
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
	"GoTodo/internal/version"
)

type adminSettingsJSON struct {
	SiteName                 string `json:"site_name"`
	DefaultTimezone          string `json:"default_timezone"`
	ShowChangelog            bool   `json:"show_changelog"`
	SiteVersion              string `json:"site_version"`
	EnableRegistration       bool   `json:"enable_registration"`
	InviteOnly               bool   `json:"invite_only"`
	MetaDescription          string `json:"meta_description"`
	EnableGlobalAnnouncement bool   `json:"enable_global_announcement"`
	GlobalAnnouncementText   string `json:"global_announcement_text"`
	EnableAPI                bool   `json:"enable_api"`

	EmailProvider         string `json:"email_provider"`
	EmailFromAddress      string `json:"email_from_address"`
	EmailFromName         string `json:"email_from_name"`
	EmailMailgunDomain    string `json:"email_mailgun_domain"`
	EmailMailgunAPIKeySet bool   `json:"email_mailgun_api_key_set"`
	EmailSMTPHost         string `json:"email_smtp_host"`
	EmailSMTPPort         int    `json:"email_smtp_port"`
	EmailSMTPUsername     string `json:"email_smtp_username"`
	EmailSMTPPasswordSet  bool   `json:"email_smtp_password_set"`
	EmailSMTPTLS          bool   `json:"email_smtp_tls"`

	GitHubOAuthClientID        string `json:"github_oauth_client_id"`
	GitHubOAuthClientSecretSet bool   `json:"github_oauth_client_secret_set"`
	GitHubOAuthConfigured      bool   `json:"github_oauth_configured"`
}

type adminSettingsPatch struct {
	SiteName                 *string `json:"site_name"`
	DefaultTimezone          *string `json:"default_timezone"`
	ShowChangelog            *bool   `json:"show_changelog"`
	EnableRegistration       *bool   `json:"enable_registration"`
	InviteOnly               *bool   `json:"invite_only"`
	MetaDescription          *string `json:"meta_description"`
	EnableGlobalAnnouncement *bool   `json:"enable_global_announcement"`
	GlobalAnnouncementText   *string `json:"global_announcement_text"`
	EnableAPI                *bool   `json:"enable_api"`

	EmailProvider      *string `json:"email_provider"`
	EmailFromAddress   *string `json:"email_from_address"`
	EmailFromName      *string `json:"email_from_name"`
	EmailMailgunDomain *string `json:"email_mailgun_domain"`
	EmailMailgunAPIKey *string `json:"email_mailgun_api_key"`
	EmailSMTPHost      *string `json:"email_smtp_host"`
	EmailSMTPPort      *int    `json:"email_smtp_port"`
	EmailSMTPUsername  *string `json:"email_smtp_username"`
	EmailSMTPPassword  *string `json:"email_smtp_password"`
	EmailSMTPTLS       *bool   `json:"email_smtp_tls"`

	GitHubOAuthClientID     *string `json:"github_oauth_client_id"`
	GitHubOAuthClientSecret *string `json:"github_oauth_client_secret"`
}

// APIV1AdminSettings handles GET/PATCH /api/v1/admin/settings.
func APIV1AdminSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		apiV1GetAdminSettings(w, r)
	case http.MethodPatch:
		apiV1PatchAdminSettings(w, r)
	default:
		utils.APIJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed.")
	}
}

func apiV1GetAdminSettings(w http.ResponseWriter, r *http.Request) {
	s, err := storage.GetSiteSettings()
	if err != nil || s == nil {
		utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to load settings.")
		return
	}
	writeAdminSettings(w, s)
}

func apiV1PatchAdminSettings(w http.ResponseWriter, r *http.Request) {
	current, err := storage.GetSiteSettings()
	if err != nil || current == nil {
		utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to load settings.")
		return
	}
	var req adminSettingsPatch
	if err := decodeJSONBody(r, &req); err != nil {
		utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body.")
		return
	}
	next := *current
	if req.SiteName != nil {
		next.SiteName = strings.TrimSpace(*req.SiteName)
	}
	if req.DefaultTimezone != nil {
		next.DefaultTimezone = strings.TrimSpace(*req.DefaultTimezone)
	}
	if req.ShowChangelog != nil {
		next.ShowChangelog = *req.ShowChangelog
	}
	if req.EnableRegistration != nil {
		next.EnableRegistration = *req.EnableRegistration
	}
	if req.InviteOnly != nil {
		next.InviteOnly = *req.InviteOnly
	}
	if req.MetaDescription != nil {
		next.MetaDescription = strings.TrimSpace(*req.MetaDescription)
	}
	if req.EnableGlobalAnnouncement != nil {
		next.EnableGlobalAnnouncement = *req.EnableGlobalAnnouncement
	}
	if req.GlobalAnnouncementText != nil {
		next.GlobalAnnouncementText = strings.TrimSpace(*req.GlobalAnnouncementText)
	}
	if req.EnableAPI != nil {
		next.EnableAPI = *req.EnableAPI
	}
	if req.EmailProvider != nil {
		next.EmailProvider = normalizeEmailProvider(*req.EmailProvider)
	}
	if req.EmailFromAddress != nil {
		next.EmailFromAddress = strings.TrimSpace(*req.EmailFromAddress)
	}
	if req.EmailFromName != nil {
		next.EmailFromName = strings.TrimSpace(*req.EmailFromName)
	}
	if req.EmailMailgunDomain != nil {
		next.EmailMailgunDomain = strings.TrimSpace(*req.EmailMailgunDomain)
	}
	if req.EmailMailgunAPIKey != nil {
		key := *req.EmailMailgunAPIKey
		if key == "" {
			next.EmailMailgunAPIKeyEnc = ""
		} else {
			enc, err := secret.Encrypt(key)
			if err != nil {
				utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to encrypt mailgun API key.")
				return
			}
			next.EmailMailgunAPIKeyEnc = enc
		}
	}
	if req.EmailSMTPHost != nil {
		next.EmailSMTPHost = strings.TrimSpace(*req.EmailSMTPHost)
	}
	if req.EmailSMTPPort != nil {
		next.EmailSMTPPort = *req.EmailSMTPPort
	}
	if req.EmailSMTPUsername != nil {
		next.EmailSMTPUsername = strings.TrimSpace(*req.EmailSMTPUsername)
	}
	if req.EmailSMTPPassword != nil {
		pass := *req.EmailSMTPPassword
		if pass == "" {
			next.EmailSMTPPasswordEnc = ""
		} else {
			enc, err := secret.Encrypt(pass)
			if err != nil {
				utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to encrypt SMTP password.")
				return
			}
			next.EmailSMTPPasswordEnc = enc
		}
	}
	if req.EmailSMTPTLS != nil {
		next.EmailSMTPTLS = *req.EmailSMTPTLS
	}
	if req.GitHubOAuthClientID != nil {
		next.GitHubOAuthClientID = strings.TrimSpace(*req.GitHubOAuthClientID)
	}
	if req.GitHubOAuthClientSecret != nil {
		sec := *req.GitHubOAuthClientSecret
		if sec == "" {
			next.GitHubOAuthClientSecretEnc = ""
		} else {
			enc, err := secret.Encrypt(sec)
			if err != nil {
				utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to encrypt GitHub OAuth client secret.")
				return
			}
			next.GitHubOAuthClientSecretEnc = enc
		}
	}
	if next.SiteName == "" || next.DefaultTimezone == "" {
		utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", "site_name and default_timezone are required.")
		return
	}
	if !utils.IsValidTimezone(next.DefaultTimezone) {
		utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid timezone.")
		return
	}
	if len(next.GlobalAnnouncementText) > 500 {
		utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", "Global announcement text must be 500 characters or less.")
		return
	}
	if next.EnableAPI && !utils.RedisAvailable() {
		utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", "Cannot enable REST API without Redis.")
		return
	}
	if errMsg := validateEmailSettings(&next); errMsg != "" {
		utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", errMsg)
		return
	}
	next.SiteVersion = "" // never persist binary version from API
	if err := storage.UpsertSiteSettings(next); err != nil {
		utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to save settings.")
		return
	}
	config.Cfg.SiteName = next.SiteName
	config.Cfg.DefaultTimezone = next.DefaultTimezone
	config.Cfg.ShowChangelog = next.ShowChangelog
	saved, _ := storage.GetSiteSettings()
	if saved == nil {
		saved = &next
	}
	writeAdminSettings(w, saved)
}

func normalizeEmailProvider(p string) string {
	p = strings.ToLower(strings.TrimSpace(p))
	switch p {
	case storage.EmailProviderMailgun, storage.EmailProviderSMTP:
		return p
	case "none", "":
		return storage.EmailProviderNone
	default:
		return p
	}
}

func validateEmailSettings(s *storage.SiteSettings) string {
	provider := normalizeEmailProvider(s.EmailProvider)
	s.EmailProvider = provider
	switch provider {
	case storage.EmailProviderNone:
		return ""
	case storage.EmailProviderMailgun:
		if s.EmailFromAddress == "" {
			return "email_from_address is required for Mailgun."
		}
		if s.EmailMailgunDomain == "" {
			return "email_mailgun_domain is required for Mailgun."
		}
		if s.EmailMailgunAPIKeyEnc == "" {
			return "email_mailgun_api_key is required for Mailgun."
		}
	case storage.EmailProviderSMTP:
		if s.EmailFromAddress == "" {
			return "email_from_address is required for SMTP."
		}
		if s.EmailSMTPHost == "" {
			return "email_smtp_host is required for SMTP."
		}
		if s.EmailSMTPPort <= 0 {
			return "email_smtp_port must be a positive integer."
		}
		if s.EmailSMTPUsername == "" {
			return "email_smtp_username is required for SMTP."
		}
		if s.EmailSMTPPasswordEnc == "" {
			return "email_smtp_password is required for SMTP."
		}
	default:
		return "email_provider must be none, mailgun, or smtp."
	}
	return ""
}

func writeAdminSettings(w http.ResponseWriter, s *storage.SiteSettings) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(adminSettingsJSON{
		SiteName:                   s.SiteName,
		DefaultTimezone:            s.DefaultTimezone,
		ShowChangelog:              s.ShowChangelog,
		SiteVersion:                version.Version, // baked into binary; not from DB
		EnableRegistration:         s.EnableRegistration,
		InviteOnly:                 s.InviteOnly,
		MetaDescription:            s.MetaDescription,
		EnableGlobalAnnouncement:   s.EnableGlobalAnnouncement,
		GlobalAnnouncementText:     s.GlobalAnnouncementText,
		EnableAPI:                  s.EnableAPI,
		EmailProvider:              s.EmailProvider,
		EmailFromAddress:           s.EmailFromAddress,
		EmailFromName:              s.EmailFromName,
		EmailMailgunDomain:         s.EmailMailgunDomain,
		EmailMailgunAPIKeySet:      s.EmailMailgunAPIKeyEnc != "",
		EmailSMTPHost:              s.EmailSMTPHost,
		EmailSMTPPort:              s.EmailSMTPPort,
		EmailSMTPUsername:          s.EmailSMTPUsername,
		EmailSMTPPasswordSet:       s.EmailSMTPPasswordEnc != "",
		EmailSMTPTLS:               s.EmailSMTPTLS,
		GitHubOAuthClientID:        s.GitHubOAuthClientID,
		GitHubOAuthClientSecretSet: s.GitHubOAuthClientSecretEnc != "",
		GitHubOAuthConfigured:      strings.TrimSpace(s.GitHubOAuthClientID) != "" && s.GitHubOAuthClientSecretEnc != "",
	})
}

// APIV1AdminUsersRouter handles /api/v1/admin/users and ban/unban.
func APIV1AdminUsersRouter(w http.ResponseWriter, r *http.Request) {
	sub := utils.ParseAPIV1Subpath(r, "admin/users")
	if sub == "" {
		if r.Method != http.MethodGet {
			utils.APIJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed.")
			return
		}
		users, err := storage.ListUsers()
		if err != nil {
			utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to list users.")
			return
		}
		out := make([]map[string]interface{}, 0, len(users))
		for _, u := range users {
			out = append(out, map[string]interface{}{
				"id":        u.ID,
				"email":     u.Email,
				"user_name": u.UserName,
				"is_banned": u.IsBanned,
			})
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	parts := strings.Split(sub, "/")
	if len(parts) == 2 {
		id, err := strconv.Atoi(parts[0])
		if err != nil || id <= 0 {
			utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid user id.")
			return
		}
		switch parts[1] {
		case "ban":
			if r.Method != http.MethodPost {
				utils.APIJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed.")
				return
			}
			if err := storage.SetUserBanned(id, true); err != nil {
				utils.APIJSONError(w, http.StatusNotFound, "not_found", "User not found.")
				return
			}
		case "unban":
			if r.Method != http.MethodPost {
				utils.APIJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed.")
				return
			}
			if err := storage.SetUserBanned(id, false); err != nil {
				utils.APIJSONError(w, http.StatusNotFound, "not_found", "User not found.")
				return
			}
		case "username":
			if r.Method != http.MethodPatch {
				utils.APIJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed.")
				return
			}
			apiV1AdminSetUsername(w, r, id)
			return
		default:
			utils.APIJSONError(w, http.StatusNotFound, "not_found", "Unknown action.")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		return
	}

	utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid user path.")
}

type adminSetUsernameRequest struct {
	UserName string `json:"user_name"`
}

func apiV1AdminSetUsername(w http.ResponseWriter, r *http.Request, userID int) {
	var req adminSetUsernameRequest
	if err := decodeJSONBody(r, &req); err != nil {
		utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body.")
		return
	}
	profile, err := domain.AdminSetUsername(r.Context(), userID, req.UserName)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			if strings.Contains(err.Error(), "already taken") {
				utils.APIJSONError(w, http.StatusConflict, "username_taken", "That username is already taken.")
				return
			}
			utils.APIJSONError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to update username.")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":        true,
		"id":        profile.ID,
		"user_name": profile.UserName,
	})
}
