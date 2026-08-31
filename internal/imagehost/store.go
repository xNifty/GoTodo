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

// UploadError is a storage-provider failure with an HTTP status when known.
type UploadError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *UploadError) Error() string {
	if e == nil {
		return "upload failed"
	}
	switch {
	case e.Code != "" && e.Status != 0:
		return fmt.Sprintf("s3 upload failed (%d): %s: %s", e.Status, e.Code, e.Message)
	case e.Status != 0 && e.Message != "":
		return fmt.Sprintf("s3 upload failed (%d): %s", e.Status, e.Message)
	case e.Status != 0:
		return fmt.Sprintf("s3 upload failed (%d)", e.Status)
	case e.Cause != nil:
		return fmt.Sprintf("s3 upload: %v", e.Cause)
	case e.Message != "":
		return e.Message
	default:
		return "upload failed"
	}
}

func (e *UploadError) Unwrap() error { return e.Cause }

// ClientError reports a 4xx response from the storage provider.
func (e *UploadError) ClientError() bool {
	return e != nil && e.Status >= 400 && e.Status < 500
}

// UserMessage is safe to show in an API error body.
func (e *UploadError) UserMessage() string {
	if e == nil {
		return "Failed to store image."
	}
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		if e.Cause != nil {
			return "Could not reach the S3 endpoint."
		}
		return "Failed to store image."
	}
	if e.Code != "" && !strings.HasPrefix(msg, e.Code+":") {
		return e.Code + ": " + msg
	}
	return msg
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
