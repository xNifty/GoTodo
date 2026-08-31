package imagehost

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
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

const (
	userMsgUnavailable   = "Image hosting is temporarily unavailable. Try again later."
	userMsgMisconfigured = "Couldn't upload that image. A site admin needs to check image hosting settings."
	userMsgGeneric       = "Couldn't upload that image. Try a different file, or try again later."
)

type uploadKind int

const (
	uploadKindUnknown uploadKind = iota
	uploadKindUnreachable
	uploadKindUnavailable
	uploadKindCredentials
	uploadKindDenied
	uploadKindNotFound
)

func (e *UploadError) kind() uploadKind {
	if e == nil {
		return uploadKindUnknown
	}
	if e.Cause != nil {
		return uploadKindUnreachable
	}
	switch strings.ToLower(strings.TrimSpace(e.Code)) {
	case "signaturedoesnotmatch", "invalidaccesskeyid", "invalidsecurity",
		"authfailure", "authorizationheadermalformed", "expiredtoken", "invalidtoken":
		return uploadKindCredentials
	case "accessdenied", "accessforbidden", "allaccessdisabled":
		return uploadKindDenied
	case "nosuchbucket", "nosuchkey", "invalidbucketname":
		return uploadKindNotFound
	}
	switch {
	case e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden:
		return uploadKindCredentials
	case e.Status == http.StatusNotFound:
		return uploadKindNotFound
	case e.Status >= 500 || e.Status == 0:
		return uploadKindUnavailable
	default:
		return uploadKindUnknown
	}
}

// UserMessage is safe to show to anyone uploading an image. It never includes
// provider XML, codes, URLs, or other internals.
func (e *UploadError) UserMessage() string {
	switch e.kind() {
	case uploadKindUnreachable, uploadKindUnavailable:
		return userMsgUnavailable
	case uploadKindCredentials, uploadKindDenied, uploadKindNotFound:
		return userMsgMisconfigured
	default:
		return userMsgGeneric
	}
}

// AdminMessage explains a storage failure to a site admin without dumping
// provider XML (SignatureDoesNotMatch bodies include signing material).
func (e *UploadError) AdminMessage() string {
	if e == nil {
		return "The storage backend rejected the test upload."
	}
	code := safeAdminCode(e.Code)
	status := e.Status
	switch e.kind() {
	case uploadKindUnreachable:
		return "Could not reach the storage API. Check the endpoint URL and that this server can make outbound HTTPS requests."
	case uploadKindCredentials:
		return adminCodeSuffix("The storage API rejected the credentials. Check the access key, secret key, and region (use auto for Cloudflare R2). The API endpoint should be the S3 host only, without the bucket in the path.", code, status)
	case uploadKindDenied:
		return adminCodeSuffix("The storage API denied the upload. Check that this access key can write to the bucket.", code, status)
	case uploadKindNotFound:
		return adminCodeSuffix("The bucket or object was not found. Check the bucket name and API endpoint.", code, status)
	case uploadKindUnavailable:
		return adminCodeSuffix("The storage API returned a gateway or server error. For Cloudflare R2, confirm the endpoint host and that region is auto.", code, status)
	default:
		return adminCodeSuffix("The storage API rejected the upload.", code, status)
	}
}

func safeAdminCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" || len(code) > 64 {
		return ""
	}
	for i := 0; i < len(code); i++ {
		c := code[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}
		return ""
	}
	return code
}

func adminCodeSuffix(msg, code string, status int) string {
	if code != "" && status != 0 {
		return fmt.Sprintf("%s (HTTP %d, %s)", msg, status, code)
	}
	if code != "" {
		return msg + " (" + code + ")"
	}
	if status != 0 {
		return fmt.Sprintf("%s (HTTP %d)", msg, status)
	}
	return msg
}

// AdminMessageFor maps any store error to an admin-safe diagnostic.
func AdminMessageFor(err error) string {
	if err == nil {
		return ""
	}
	var ue *UploadError
	if errors.As(err, &ue) {
		return ue.AdminMessage()
	}
	return "The storage backend rejected the test upload. Check the configuration and try again."
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
