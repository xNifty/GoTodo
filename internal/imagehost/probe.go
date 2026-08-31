package imagehost

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"time"
)

// TinyPNG is a 1×1 PNG used for connection tests.
var TinyPNG = mustDecodePNG("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==")

func mustDecodePNG(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// ProbeResult is an admin-facing connection test outcome.
type ProbeResult struct {
	OK          bool
	PublicURLOK bool
	Message     string
}

type objectDeleter interface {
	Delete(ctx context.Context, key string) error
}

var probePublicClient = &http.Client{Timeout: 12 * time.Second}

// Probe uploads a tiny image with the given config, checks the public URL when
// possible, then deletes the probe object.
func Probe(ctx context.Context, cfg Config) ProbeResult {
	if msg := cfg.Validate(); msg != "" {
		return ProbeResult{Message: msg}
	}
	if !cfg.Enabled() {
		return ProbeResult{Message: "Select S3-compatible or local uploads first."}
	}
	store, err := NewStore(cfg)
	if err != nil {
		if err == ErrDisabled {
			return ProbeResult{Message: "Select S3-compatible or local uploads first."}
		}
		return ProbeResult{Message: AdminMessageFor(err)}
	}
	key, err := NewObjectKey(TypePNG)
	if err != nil {
		return ProbeResult{Message: "Could not generate a test object key."}
	}
	obj := Object{Key: key, ContentType: TypePNG, Data: TinyPNG}
	publicURL, err := store.Put(ctx, obj)
	if err != nil {
		return ProbeResult{Message: AdminMessageFor(err)}
	}
	result := ProbeResult{OK: true, Message: "Connected. A test image was uploaded and removed."}
	if NormalizeProvider(cfg.Provider) == ProviderLocal {
		result.PublicURLOK = true
	} else {
		result.PublicURLOK = probePublicURL(ctx, publicURL)
		if !result.PublicURLOK {
			result.Message = "Upload succeeded, but the public URL did not return the image. Check the public URL and that the bucket allows public reads (R2 public development URL or a custom domain)."
		}
	}
	if d, ok := store.(objectDeleter); ok {
		if err := d.Delete(ctx, obj.Key); err != nil {
			result.Message += " The test object could not be deleted automatically."
		}
	}
	return result
}

func probePublicURL(ctx context.Context, rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" || !strings.HasPrefix(strings.ToLower(rawURL), "http") {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return false
	}
	resp, err := probePublicClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	return ct == "" || strings.HasPrefix(ct, "image/") || strings.HasPrefix(ct, "application/octet-stream")
}
