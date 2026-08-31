package imagehost

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
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

// defaultS3HTTPClient talks HTTP/1.1 only. HTTP/2 to Cloudflare R2
// (*.r2.cloudflarestorage.com) is a documented source of 502 Bad Gateway.
func defaultS3HTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ForceAttemptHTTP2 = false
	transport.DisableCompression = true
	transport.ExpectContinueTimeout = 0
	// An empty map disables HTTP/2 even when ALPN would otherwise negotiate it.
	// A nil map lets the runtime enable HTTP/2.
	transport.TLSNextProto = map[string]func(authority string, c *tls.Conn) http.RoundTripper{}
	return &http.Client{Timeout: 45 * time.Second, Transport: transport}
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
		client = defaultS3HTTPClient()
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
		return "", &UploadError{Message: "Could not reach the S3 endpoint.", Cause: err}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", s3StatusError(resp.StatusCode, body)
	}
	return JoinPublicURL(s.cfg.S3PublicURL, obj.Key), nil
}

type s3ErrorBody struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}

func s3StatusError(status int, body []byte) error {
	code, message := parseS3XMLError(body)
	if message == "" {
		message = sanitizeS3ErrorMessage(string(body))
	}
	if message == "" {
		message = fmt.Sprintf("S3 returned HTTP %d.", status)
	}
	return &UploadError{Status: status, Code: code, Message: message}
}

func parseS3XMLError(body []byte) (code, message string) {
	var parsed s3ErrorBody
	if err := xml.Unmarshal(body, &parsed); err != nil {
		return "", ""
	}
	return strings.TrimSpace(parsed.Code), strings.TrimSpace(parsed.Message)
}

func sanitizeS3ErrorMessage(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	lower := strings.ToLower(raw)
	if strings.Contains(lower, "<html") {
		return "the storage provider returned an HTML error page (often HTTP/2 or a gateway failure)"
	}
	if len(raw) > 300 {
		return raw[:300]
	}
	return raw
}

// Delete removes an object. A missing object is treated as success.
func (s *S3Store) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !SafeObjectKey(key) {
		return fmt.Errorf("invalid object key")
	}
	req, err := s.newDeleteRequest(ctx, key, Now())
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return &UploadError{Message: "Could not reach the S3 endpoint.", Cause: err}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode == http.StatusNotFound || (resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil
	}
	return s3StatusError(resp.StatusCode, body)
}

func (s *S3Store) objectTarget(key string) (requestURL, canonicalURI, host string, err error) {
	endpoint, err := url.Parse(s.cfg.S3Endpoint)
	if err != nil {
		return "", "", "", fmt.Errorf("parse s3 endpoint: %w", err)
	}
	if endpoint.Scheme == "" {
		endpoint.Scheme = "https"
	}
	escapedKey := escapePath(key)
	escapedBucket := escapePath(s.cfg.S3Bucket)
	host = endpoint.Host
	base := endpoint.Scheme + "://" + endpoint.Host
	if s.cfg.S3ForcePathStyle {
		canonicalURI = "/" + escapedBucket + "/" + escapedKey
		requestURL = base + canonicalURI
		return requestURL, canonicalURI, host, nil
	}
	host = s.cfg.S3Bucket + "." + endpoint.Host
	canonicalURI = "/" + escapedKey
	requestURL = endpoint.Scheme + "://" + host + canonicalURI
	return requestURL, canonicalURI, host, nil
}

func (s *S3Store) authorize(req *http.Request, canonicalURI, host, payloadHash string, now time.Time) {
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	canonicalHeaders := "host:" + host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := strings.Join([]string{
		req.Method,
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
	req.Header.Set("Host", host)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Authorization", auth)
	req.Host = host
}

func (s *S3Store) newDeleteRequest(ctx context.Context, key string, now time.Time) (*http.Request, error) {
	requestURL, canonicalURI, host, err := s.objectTarget(key)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
	if err != nil {
		return nil, err
	}
	s.authorize(req, canonicalURI, host, sha256Hex(nil), now)
	return req, nil
}

func (s *S3Store) newPutRequest(ctx context.Context, obj Object, now time.Time) (*http.Request, error) {
	requestURL, canonicalURI, host, err := s.objectTarget(obj.Key)
	if err != nil {
		return nil, err
	}
	contentType := obj.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	payloadHash := sha256Hex(obj.Data)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL, bytes.NewReader(obj.Data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(obj.Data))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(obj.Data)), nil
	}
	s.authorize(req, canonicalURI, host, payloadHash, now)
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
