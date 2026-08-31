package imagehost

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

const (
	ProviderNone  = ""
	ProviderS3    = "s3"
	ProviderLocal = "local"

	DefaultMaxBytes int64 = 5 << 20  // 5 MiB
	MinMaxBytes     int64 = 64 << 10 // 64 KiB
	MaxMaxBytes     int64 = 50 << 20 // 50 MiB

	DefaultLocalPath = "data/uploads"
)

// Config is the runtime image-hosting configuration (secrets decrypted).
type Config struct {
	Provider string
	MaxBytes int64

	S3Endpoint       string
	S3Region         string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
	S3PublicURL      string
	S3ForcePathStyle bool

	LocalPath       string
	LocalPublicBase string
}

// NormalizeProvider maps UI/API values onto the stored provider constants.
func NormalizeProvider(p string) string {
	p = strings.ToLower(strings.TrimSpace(p))
	switch p {
	case ProviderS3, ProviderLocal:
		return p
	case "none", "disabled", "":
		return ProviderNone
	default:
		return p
	}
}

// ClampMaxBytes returns a safe size limit. Values below the minimum
// (including 0 from uninitialized structs) become the default.
func ClampMaxBytes(n int64) int64 {
	if n < MinMaxBytes {
		return DefaultMaxBytes
	}
	if n > MaxMaxBytes {
		return MaxMaxBytes
	}
	return n
}

// NormalizeMaxBytes accepts either a byte count or a UI value in MiB (1–50).
// The admin form displays megabytes; a payload of 5 must not be rejected as
// "below 64 KiB".
func NormalizeMaxBytes(n int64) int64 {
	if n >= 1 && n <= 50 {
		n = n << 20
	}
	return ClampMaxBytes(n)
}

// Enabled reports whether uploads should be accepted.
func (c Config) Enabled() bool {
	p := NormalizeProvider(c.Provider)
	return p == ProviderS3 || p == ProviderLocal
}

// Validate returns a user-facing error message, or "" if the config is usable.
// It trims whitespace (and a trailing slash on URLs) but does not rewrite the
// admin-entered S3 endpoint, region, or path-style flag. Those values are
// persisted as submitted so the settings form does not appear to revert.
func (c *Config) Validate() string {
	c.Provider = NormalizeProvider(c.Provider)
	c.MaxBytes = ClampMaxBytes(c.MaxBytes)
	c.S3Endpoint = strings.TrimSpace(c.S3Endpoint)
	c.S3Region = strings.TrimSpace(c.S3Region)
	c.S3Bucket = strings.TrimSpace(c.S3Bucket)
	c.S3AccessKey = strings.TrimSpace(c.S3AccessKey)
	c.S3SecretKey = strings.TrimSpace(c.S3SecretKey)
	c.S3PublicURL = strings.TrimRight(strings.TrimSpace(c.S3PublicURL), "/")
	c.LocalPath = strings.TrimSpace(c.LocalPath)
	c.LocalPublicBase = strings.TrimRight(strings.TrimSpace(c.LocalPublicBase), "/")
	c.S3Endpoint = strings.TrimRight(c.S3Endpoint, "/")

	switch c.Provider {
	case ProviderNone:
		return ""
	case ProviderLocal:
		if c.LocalPath == "" {
			c.LocalPath = DefaultLocalPath
		}
		if path.IsAbs(c.LocalPath) {
			return ""
		}
		if strings.Contains(c.LocalPath, "..") {
			return "image_local_path must not contain '..'."
		}
		return ""
	case ProviderS3:
		if c.S3Endpoint == "" {
			return "image_s3_endpoint is required for S3 (use the S3, R2, or Spaces API endpoint)."
		}
		if _, err := url.ParseRequestURI(c.S3Endpoint); err != nil {
			return "image_s3_endpoint must be an absolute URL (https://…)."
		}
		if c.S3Region == "" {
			return "image_s3_region is required for S3 (use auto for Cloudflare R2)."
		}
		if c.S3Bucket == "" {
			return "image_s3_bucket is required for S3."
		}
		if c.S3AccessKey == "" {
			return "image_s3_access_key is required for S3."
		}
		if c.S3SecretKey == "" {
			return "image_s3_secret_key is required for S3."
		}
		if c.S3PublicURL == "" {
			return "image_s3_public_url is required for S3 (CDN or public bucket URL)."
		}
		if _, err := url.ParseRequestURI(c.S3PublicURL); err != nil {
			return "image_s3_public_url must be an absolute URL (https://…)."
		}
		return ""
	default:
		return "image_hosting_provider must be none, s3, or local."
	}
}

// PublicOrigin returns the origin (scheme://host) of the public image URL, if any.
// Used to extend Content-Security-Policy img-src for S3/CDN hosts.
func (c Config) PublicOrigin() string {
	raw := ""
	switch NormalizeProvider(c.Provider) {
	case ProviderS3:
		raw = c.S3PublicURL
	case ProviderLocal:
		raw = c.LocalPublicBase
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	scheme := u.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return scheme + "://" + u.Host
}

// JoinPublicURL concatenates a public base and object key without double slashes.
func JoinPublicURL(base, key string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if base == "" {
		return "/" + key
	}
	return base + "/" + key
}

// ErrDisabled is returned when image hosting is not configured.
var ErrDisabled = fmt.Errorf("image hosting is not configured")
