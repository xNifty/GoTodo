package imagehost

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"
)

// Object is a validated image ready to persist.
type Object struct {
	Key         string
	ContentType string
	Data        []byte
}

// Store persists an image and returns its public URL.
type Store interface {
	Put(ctx context.Context, obj Object) (publicURL string, error error)
}

// NewStore builds a Store for the given config. LocalPublicBase should already
// be set by the caller for local uploads.
func NewStore(cfg Config) (Store, error) {
	if msg := cfg.Validate(); msg != "" {
		return nil, fmt.Errorf("%s", msg)
	}
	switch cfg.Provider {
	case ProviderNone:
		return nil, ErrDisabled
	case ProviderLocal:
		return NewLocalStore(cfg.LocalPath, cfg.LocalPublicBase)
	case ProviderS3:
		return NewS3Store(cfg, nil)
	default:
		return nil, fmt.Errorf("unsupported image hosting provider %q", cfg.Provider)
	}
}

// NewObjectKey returns a collision-resistant object name with the right extension.
func NewObjectKey(contentType string) (string, error) {
	ext := ExtForContentType(contentType)
	if ext == "" {
		return "", ErrNotImage
	}
	var b [16]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		return "", fmt.Errorf("generate object key: %w", err)
	}
	// UUID-shaped hex with dashes for readability, no dependency on google/uuid at call sites.
	hexed := hex.EncodeToString(b[:])
	key := hexed[0:8] + "-" + hexed[8:12] + "-" + hexed[12:16] + "-" + hexed[16:20] + "-" + hexed[20:32]
	return key + ext, nil
}

// Prepare validates bytes against type and size rules and assigns a key.
func Prepare(cfg Config, data []byte) (Object, error) {
	if !cfg.Enabled() {
		return Object{}, ErrDisabled
	}
	max := ClampMaxBytes(cfg.MaxBytes)
	if int64(len(data)) > max {
		return Object{}, fmt.Errorf("image exceeds the %d byte limit", max)
	}
	if len(data) == 0 {
		return Object{}, ErrNotImage
	}
	ct, err := DetectImage(data)
	if err != nil {
		return Object{}, err
	}
	key, err := NewObjectKey(ct)
	if err != nil {
		return Object{}, err
	}
	return Object{Key: key, ContentType: ct, Data: data}, nil
}

// SafeObjectKey reports whether key is a single-segment hosted filename.
func SafeObjectKey(key string) bool {
	key = strings.TrimSpace(key)
	if key == "" || strings.Contains(key, "/") || strings.Contains(key, "\\") || strings.Contains(key, "..") {
		return false
	}
	ext := ""
	if i := strings.LastIndex(key, "."); i >= 0 {
		ext = strings.ToLower(key[i:])
	}
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return len(key) <= 80
	default:
		return false
	}
}

// Now is overridable in tests (S3 signing timestamps).
var Now = func() time.Time { return time.Now().UTC() }
