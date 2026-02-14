package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
)

func RequireHTMX(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") != "true" {
			basePath := GetBasePath()
			http.Redirect(w, r, basePath+"/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

type contextKey string

const cspNonceKey contextKey = "csp-nonce"

func setCSPNonce(r *http.Request, nonce string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), cspNonceKey, nonce))
}

func GetCSPNonce(r *http.Request) string {
	if r == nil {
		return ""
	}
	if v := r.Context().Value(cspNonceKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// generateNonce returns a base64-encoded random value for CSP nonces.
func generateNonce() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b[:]), nil
}

// SecurityHeadersMiddleware adds common security headers. Adjust CSP sources to match asset hosts.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := generateNonce()
		if err == nil {
			r = setCSPNonce(r, nonce)
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		if os.Getenv("ENV") == "production" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		cspNonce := ""
		if nonce != "" {
			cspNonce = "'nonce-" + nonce + "' "
		}

		// CSP with specific whitelisted CDN resources
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"base-uri 'self'; "+
				"frame-ancestors 'none'; "+
				"script-src 'self' "+cspNonce+
				"https://cdn.jsdelivr.net/npm/@popperjs/core@2.11.6/dist/umd/popper.min.js "+
				"https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha1/dist/js/bootstrap.min.js "+
				"https://cdn.jsdelivr.net/npm/sortablejs@1.15.0/Sortable.min.js "+
				"https://unpkg.com/htmx.org@2.0.3; "+
				"style-src 'self' 'unsafe-inline' "+
				"https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha1/dist/css/bootstrap.min.css "+
				"https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css; "+
				"img-src 'self' data:; "+
				"font-src 'self' data: https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/fonts/; "+
				"connect-src 'self'; "+
				"object-src 'none'",
		)

		next.ServeHTTP(w, r)
	})
}
