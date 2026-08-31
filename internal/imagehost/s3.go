package imagehost

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const s3Service = "s3"

// HTTPDoer is the subset of http.Client used by S3Store (tests inject a fake).
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// S3Store uploads objects to an S3-compatible API (AWS, Cloudflare R2, DigitalOcean Spaces, MinIO).
type S3Store struct {
	cfg    Config
	client HTTPDoer
}

// NewS3Store builds an S3-compatible store. client may be nil to use a default timeout client.
func NewS3Store(cfg Config, client HTTPDoer) (*S3Store, error) {
	if msg := cfg.Validate(); msg != "" {
		return nil, fmt.Errorf("%s", msg)
	}
	if cfg.Provider != ProviderS3 {
		return nil, fmt.Errorf("s3 store requires provider %q", ProviderS3)
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &S3Store{cfg: cfg, client: client}, nil
}

// Put uploads obj and returns the configured public URL for the key.
func (s *S3Store) Put(ctx context.Context, obj Object) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if !SafeObjectKey(obj.Key) {
		return "", fmt.Errorf("invalid object key")
	}
	req, err := s.newPutRequest(ctx, obj, Now())
	if err != nil {
		return "", err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("s3 upload: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = resp.Status
		}
		return "", fmt.Errorf("s3 upload failed (%d): %s", resp.StatusCode, msg)
	}
	return JoinPublicURL(s.cfg.S3PublicURL, obj.Key), nil
}

func (s *S3Store) newPutRequest(ctx context.Context, obj Object, now time.Time) (*http.Request, error) {
	endpoint, err := url.Parse(s.cfg.S3Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse s3 endpoint: %w", err)
	}
	if endpoint.Scheme == "" {
		endpoint.Scheme = "https"
	}

	escapedKey := escapePath(obj.Key)
	escapedBucket := escapePath(s.cfg.S3Bucket)
	var requestURL string
	var canonicalURI string
	host := endpoint.Host

	base := strings.TrimRight(endpoint.Scheme+"://"+endpoint.Host, "/")
	if s.cfg.S3ForcePathStyle {
		canonicalURI = "/" + escapedBucket + "/" + escapedKey
		requestURL = base + canonicalURI
	} else {
		host = s.cfg.S3Bucket + "." + endpoint.Host
		canonicalURI = "/" + escapedKey
		requestURL = endpoint.Scheme + "://" + host + canonicalURI
	}

	payloadHash := sha256Hex(obj.Data)
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	contentType := obj.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	canonicalHeaders := "content-type:" + contentType + "\n" +
		"host:" + host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := strings.Join([]string{
		http.MethodPut,
		canonicalURI,
		"",
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	credentialScope := dateStamp + "/" + s.cfg.S3Region + "/" + s3Service + "/aws4_request"
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(s.cfg.S3SecretKey, dateStamp, s.cfg.S3Region, s3Service), []byte(stringToSign)))
	auth := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.cfg.S3AccessKey, credentialScope, signedHeaders, signature)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL, bytes.NewReader(obj.Data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Host", host)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Authorization", auth)
	req.Host = host
	req.ContentLength = int64(len(obj.Data))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(obj.Data)), nil
	}
	return req, nil
}

func escapePath(p string) string {
	// AWS URI encode: encode everything except unreserved chars, keep slashes.
	var b strings.Builder
	for i := 0; i < len(p); i++ {
		c := p[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' || c == '/' {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "%%%02X", c)
	}
	return b.String()
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key, data []byte) []byte {
	m := hmac.New(sha256.New, key)
	_, _ = m.Write(data)
	return m.Sum(nil)
}

func signingKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}
