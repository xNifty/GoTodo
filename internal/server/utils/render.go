package utils

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"GoTodo/internal/config"
	"GoTodo/internal/storage"
	"GoTodo/internal/version"
)

var Templates *template.Template
var BasePath string

func InitializeTemplates() error {
	var err error
	// Load repo config (fallbacks to env/defaults internally)
	config.Load()
	BasePath = config.Cfg.BasePath
	if BasePath == "" {
		BasePath = "/"
	}

	BasePath = strings.TrimSuffix(BasePath, "/")
	if BasePath == "" {
		BasePath = "/"
	}

	Templates, err = template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"hasPermission": func(permissions []string, permission string) bool {
			for _, p := range permissions {
				if p == permission {
					return true
				}
			}
			return false
		},
		"basePath": func() string {
			return GetBasePath()
		},
		"themeIs": func(theme interface{}, want string) bool {
			return fmt.Sprintf("%v", theme) == want
		},
	}).ParseGlob("internal/server/templates/*.html")
	if err != nil {
		return err
	}
	_, err = Templates.ParseGlob("internal/server/templates/partials/*.html")
	return err
}

// RenderTemplate renders templates and injects AssetVersion and optional theme from cookie.
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) error {
	envAsset := os.Getenv("ASSET_VERSION")
	fileAsset := ""
	if b, err := os.ReadFile("internal/server/public/.asset_version"); err == nil {
		if v := strings.TrimSpace(string(b)); v != "" {
			fileAsset = v
		}
	}

	// assetVersion preference: env -> .asset_version file -> config -> default
	assetVersion := envAsset
	if assetVersion == "" {
		assetVersion = fileAsset
	}
	if assetVersion == "" {
		if config.Cfg.AssetVersion != "" {
			assetVersion = config.Cfg.AssetVersion
		} else {
			assetVersion = "20251130"
		}
	}

	// Use minified assets only when an explicit env or .asset_version file is present
	// and the minified files actually exist next to the sources. This prevents
	// accidentally serving .min files on dev if they were committed or missing.
	useMinified := false
	if envAsset != "" {
		useMinified = true
	} else if fileAsset != "" {
		// Check that both minified outputs exist before enabling.
		if _, errJs := os.Stat("internal/server/public/js/site.min.js"); errJs == nil {
			if _, errCss := os.Stat("internal/server/public/css/site.min.css"); errCss == nil {
				useMinified = true
			}
		}
	}

	var execErr error
	// If data is a map, inject AssetVersion and theme
	if ctx, ok := data.(map[string]interface{}); ok {
		ctx["AssetVersion"] = assetVersion
		ctx["UseMinifiedAssets"] = useMinified
		// Inject site config values. Prefer DB-backed settings when available.
		ctx["SiteName"] = config.Cfg.SiteName
		ctx["DefaultTimezone"] = config.Cfg.DefaultTimezone
		ctx["ShowChangelog"] = config.Cfg.ShowChangelog
		// Site version comes only from the baked-in binary; never from DB
		ctx["SiteVersion"] = version.Version
		if s, err := storage.GetSiteSettings(); err == nil && s != nil {
			if s.SiteName != "" {
				ctx["SiteName"] = s.SiteName
			}
			if s.DefaultTimezone != "" {
				ctx["DefaultTimezone"] = s.DefaultTimezone
			}
			ctx["ShowChangelog"] = s.ShowChangelog
		}
		// Inject theme from cookie if present
		if r != nil {
			if c, err := r.Cookie("theme"); err == nil {
				ctx["Theme"] = c.Value
			}
		}
		execErr = Templates.ExecuteTemplate(w, tmpl, ctx)
	} else {
		ctx := map[string]interface{}{
			"Data":              data,
			"AssetVersion":      assetVersion,
			"UseMinifiedAssets": useMinified,
		}
		// Inject site config values. Prefer DB for mutable fields; site version is baked-in only.
		ctx["SiteName"] = config.Cfg.SiteName
		ctx["DefaultTimezone"] = config.Cfg.DefaultTimezone
		ctx["ShowChangelog"] = config.Cfg.ShowChangelog
		ctx["SiteVersion"] = version.Version
		if s, err := storage.GetSiteSettings(); err == nil && s != nil {
			if s.SiteName != "" {
				ctx["SiteName"] = s.SiteName
			}
			if s.DefaultTimezone != "" {
				ctx["DefaultTimezone"] = s.DefaultTimezone
			}
			ctx["ShowChangelog"] = s.ShowChangelog
		}
		if r != nil {
			if c, err := r.Cookie("theme"); err == nil {
				ctx["Theme"] = c.Value
			}
		}
		execErr = Templates.ExecuteTemplate(w, tmpl, ctx)
	}
	if execErr != nil {
		fmt.Println("Error parsing template: ", execErr)
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.Error(w, execErr.Error(), http.StatusInternalServerError)
		return execErr
	}
	return nil
}

// GetBasePath returns the base path for use in templates
func GetBasePath() string {
	return BasePath
}
