package imagehost

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalStorePutAndRead(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(dir, "https://example.test/uploads")
	if err != nil {
		t.Fatal(err)
	}
	obj, err := Prepare(Config{Provider: ProviderLocal, MaxBytes: DefaultMaxBytes, LocalPath: dir}, tinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	url, err := store.Put(context.Background(), obj)
	if err != nil {
		t.Fatal(err)
	}
	wantURL := "https://example.test/uploads/" + obj.Key
	if url != wantURL {
		t.Fatalf("url=%q want %q", url, wantURL)
	}
	data, ct, err := store.Read(obj.Key)
	if err != nil {
		t.Fatal(err)
	}
	if ct != TypePNG {
		t.Fatalf("content-type=%q", ct)
	}
	if !bytesEqual(data, tinyPNG) {
		t.Fatal("round-trip bytes mismatch")
	}
	if _, err := os.Stat(filepath.Join(dir, obj.Key)); err != nil {
		t.Fatal(err)
	}
}

func TestLocalStoreRejectsTraversal(t *testing.T) {
	store, err := NewLocalStore(t.TempDir(), "/uploads")
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Put(context.Background(), Object{Key: "../x.png", ContentType: TypePNG, Data: tinyPNG})
	if err == nil {
		t.Fatal("expected traversal rejection")
	}
	_, _, err = store.Read("../x.png")
	if err == nil {
		t.Fatal("expected read traversal rejection")
	}
}

func TestS3StorePathStylePut(t *testing.T) {
	var gotPath, gotAuth, gotHost, gotType, gotSHA string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotHost = r.Host
		gotType = r.Header.Get("Content-Type")
		gotSHA = r.Header.Get("X-Amz-Content-Sha256")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	fixed := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	Now = func() time.Time { return fixed }
	t.Cleanup(func() { Now = func() time.Time { return time.Now().UTC() } })

	cfg := Config{
		Provider:         ProviderS3,
		MaxBytes:         DefaultMaxBytes,
		S3Endpoint:       srv.URL,
		S3Region:         "auto",
		S3Bucket:         "media",
		S3AccessKey:      "AKIAEXAMPLE",
		S3SecretKey:      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		S3PublicURL:      "https://cdn.example.com",
		S3ForcePathStyle: true,
	}
	store, err := NewS3Store(cfg, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	obj := Object{Key: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png", ContentType: TypePNG, Data: tinyPNG}
	url, err := store.Put(context.Background(), obj)
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://cdn.example.com/"+obj.Key {
		t.Fatalf("public url=%q", url)
	}
	if gotPath != "/media/"+obj.Key {
		t.Fatalf("path=%q", gotPath)
	}
	if !strings.HasPrefix(gotAuth, "AWS4-HMAC-SHA256 Credential=AKIAEXAMPLE/20260831/auto/s3/aws4_request") {
		t.Fatalf("authorization=%q", gotAuth)
	}
	if !strings.Contains(gotAuth, "SignedHeaders=host;x-amz-content-sha256;x-amz-date") {
		t.Fatalf("signed headers missing: %s", gotAuth)
	}
	if gotType != TypePNG {
		t.Fatalf("content-type=%q", gotType)
	}
	sum := sha256.Sum256(tinyPNG)
	if gotSHA != hex.EncodeToString(sum[:]) {
		t.Fatalf("payload hash=%q", gotSHA)
	}
	if !bytesEqual(gotBody, tinyPNG) {
		t.Fatal("uploaded body mismatch")
	}
	if gotHost == "" {
		t.Fatal("missing host")
	}
}

func TestS3StoreVirtualHostPut(t *testing.T) {
	cfg := Config{
		Provider:         ProviderS3,
		S3Endpoint:       "https://s3.us-east-1.amazonaws.com",
		S3Region:         "us-east-1",
		S3Bucket:         "media",
		S3AccessKey:      "AKID",
		S3SecretKey:      "secret",
		S3PublicURL:      "https://media.example.com",
		S3ForcePathStyle: false,
	}
	store, err := NewS3Store(cfg, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	req, err := store.newPutRequest(context.Background(), Object{
		Key: "pic.jpg", ContentType: TypeJPEG, Data: tinyPNG,
	}, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(req.URL.Host, "media.") {
		t.Fatalf("virtual host=%q", req.URL.Host)
	}
	if req.URL.Path != "/pic.jpg" {
		t.Fatalf("path=%q", req.URL.Path)
	}
}

func TestS3StoreErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><Error><Code>AccessDenied</Code><Message>Access Denied</Message></Error>`))
	}))
	t.Cleanup(srv.Close)
	cfg := Config{
		Provider:         ProviderS3,
		S3Endpoint:       srv.URL,
		S3Region:         "auto",
		S3Bucket:         "media",
		S3AccessKey:      "k",
		S3SecretKey:      "s",
		S3PublicURL:      "https://cdn.example.com",
		S3ForcePathStyle: true,
	}
	store, err := NewS3Store(cfg, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Put(context.Background(), Object{Key: "a.png", ContentType: TypePNG, Data: tinyPNG})
	var ue *UploadError
	if err == nil || !asUploadError(err, &ue) {
		t.Fatalf("err=%v", err)
	}
	if ue.Status != http.StatusForbidden || ue.Code != "AccessDenied" {
		t.Fatalf("upload error=%+v", ue)
	}
	if !ue.ClientError() {
		t.Fatal("403 should be a client error")
	}
	if !strings.Contains(ue.UserMessage(), "AccessDenied") {
		t.Fatalf("user message=%q", ue.UserMessage())
	}
}

func TestS3StoreHTMLGatewayError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html><body>502 Bad Gateway</body></html>"))
	}))
	t.Cleanup(srv.Close)
	cfg := Config{
		Provider:         ProviderS3,
		S3Endpoint:       srv.URL,
		S3Region:         "auto",
		S3Bucket:         "media",
		S3AccessKey:      "k",
		S3SecretKey:      "s",
		S3PublicURL:      "https://cdn.example.com",
		S3ForcePathStyle: true,
	}
	store, err := NewS3Store(cfg, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Put(context.Background(), Object{Key: "a.png", ContentType: TypePNG, Data: tinyPNG})
	var ue *UploadError
	if err == nil || !asUploadError(err, &ue) {
		t.Fatalf("err=%v", err)
	}
	if ue.Status != http.StatusBadGateway {
		t.Fatalf("status=%d", ue.Status)
	}
	if !strings.Contains(ue.UserMessage(), "HTML") {
		t.Fatalf("user message=%q", ue.UserMessage())
	}
}

func TestS3StoreR2DashboardEndpointDoesNotDoubleBucket(t *testing.T) {
	cfg := Config{
		Provider:         ProviderS3,
		S3Endpoint:       "https://abc123.r2.cloudflarestorage.com/media",
		S3Region:         "ENAM",
		S3Bucket:         "media",
		S3AccessKey:      "AKID",
		S3SecretKey:      "secret",
		S3PublicURL:      "https://pub-example.r2.dev",
		S3ForcePathStyle: false,
	}
	store, err := NewS3Store(cfg, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	if store.cfg.S3Endpoint != "https://abc123.r2.cloudflarestorage.com" {
		t.Fatalf("endpoint=%q", store.cfg.S3Endpoint)
	}
	if !store.cfg.S3ForcePathStyle {
		t.Fatal("R2 should force path-style")
	}
	if store.cfg.S3Region != "auto" {
		t.Fatalf("region=%q", store.cfg.S3Region)
	}
	req, err := store.newPutRequest(context.Background(), Object{
		Key: "pic.jpg", ContentType: TypeJPEG, Data: tinyPNG,
	}, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.Path != "/media/pic.jpg" {
		t.Fatalf("path=%q (bucket should appear once)", req.URL.Path)
	}
	if req.URL.Host != "abc123.r2.cloudflarestorage.com" {
		t.Fatalf("host=%q", req.URL.Host)
	}
	auth := req.Header.Get("Authorization")
	if !strings.Contains(auth, "/auto/s3/") {
		t.Fatalf("credential scope should use auto, got %s", auth)
	}
	if strings.Contains(auth, "content-type") {
		t.Fatalf("content-type should not be signed: %s", auth)
	}
}

func TestDefaultS3HTTPClientDisablesHTTP2(t *testing.T) {
	cfg := Config{
		Provider:         ProviderS3,
		S3Endpoint:       "https://abc123.r2.cloudflarestorage.com",
		S3Region:         "auto",
		S3Bucket:         "media",
		S3AccessKey:      "k",
		S3SecretKey:      "s",
		S3PublicURL:      "https://cdn.example.com",
		S3ForcePathStyle: true,
	}
	store, err := NewS3Store(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	client, ok := store.client.(*http.Client)
	if !ok {
		t.Fatalf("client type=%T", store.client)
	}
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type=%T", client.Transport)
	}
	if tr.ForceAttemptHTTP2 {
		t.Fatal("HTTP/2 should be disabled for S3")
	}
	if tr.TLSNextProto == nil {
		t.Fatal("nil TLSNextProto allows the runtime to negotiate HTTP/2")
	}
}

func asUploadError(err error, target **UploadError) bool {
	return errors.As(err, target)
}

func TestNewStore(t *testing.T) {
	if _, err := NewStore(Config{}); err != ErrDisabled {
		t.Fatalf("err=%v", err)
	}
	dir := t.TempDir()
	st, err := NewStore(Config{Provider: ProviderLocal, LocalPath: dir, LocalPublicBase: "/uploads"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.(*LocalStore); !ok {
		t.Fatalf("type=%T", st)
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
