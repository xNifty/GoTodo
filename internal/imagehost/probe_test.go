package imagehost

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadErrorUserMessageHidesInternals(t *testing.T) {
	cases := []struct {
		err      *UploadError
		contains string
		forbid   []string
	}{
		{
			err:      &UploadError{Status: 403, Code: "AccessDenied", Message: "Access Denied"},
			contains: "site admin",
			forbid:   []string{"AccessDenied", "403", "XML"},
		},
		{
			err:      &UploadError{Status: 403, Code: "SignatureDoesNotMatch", Message: "CanonicalRequest\nPUT\n/secret"},
			contains: "site admin",
			forbid:   []string{"SignatureDoesNotMatch", "CanonicalRequest", "PUT"},
		},
		{
			err:      &UploadError{Status: 502, Message: "<html><body>502 Bad Gateway</body></html>"},
			contains: "temporarily unavailable",
			forbid:   []string{"<html", "502", "Bad Gateway"},
		},
		{
			err:      &UploadError{Message: "Could not reach the S3 endpoint.", Cause: errors.New("dial tcp 10.0.0.1:443")},
			contains: "temporarily unavailable",
			forbid:   []string{"10.0.0.1", "dial tcp", "S3"},
		},
	}
	for _, tt := range cases {
		got := tt.err.UserMessage()
		if !strings.Contains(strings.ToLower(got), strings.ToLower(tt.contains)) {
			t.Fatalf("UserMessage=%q want substring %q", got, tt.contains)
		}
		for _, bad := range tt.forbid {
			if strings.Contains(got, bad) {
				t.Fatalf("UserMessage leaked %q: %q", bad, got)
			}
		}
	}
}

func TestUploadErrorAdminMessageUsesCodeNotXML(t *testing.T) {
	err := &UploadError{Status: 403, Code: "SignatureDoesNotMatch", Message: "CanonicalRequest\nPUT\n/bucket/key"}
	got := err.AdminMessage()
	if !strings.Contains(got, "credentials") {
		t.Fatalf("AdminMessage=%q", got)
	}
	if !strings.Contains(got, "SignatureDoesNotMatch") {
		t.Fatalf("admin should see code: %q", got)
	}
	if strings.Contains(got, "CanonicalRequest") || strings.Contains(got, "/bucket/key") {
		t.Fatalf("admin message leaked signing material: %q", got)
	}
}

func TestProbeLocal(t *testing.T) {
	dir := t.TempDir()
	got := Probe(context.Background(), Config{
		Provider:        ProviderLocal,
		LocalPath:       dir,
		LocalPublicBase: "/uploads",
		MaxBytes:        DefaultMaxBytes,
	})
	if !got.OK || !got.PublicURLOK {
		t.Fatalf("probe=%+v", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("probe object should be deleted, leftover=%v", names(entries))
	}
}

func TestProbeS3SuccessAndPublicURL(t *testing.T) {
	var put, del int
	s3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			put++
			body, _ := io.ReadAll(r.Body)
			if !bytesEqual(body, TinyPNG) {
				t.Errorf("unexpected put body")
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			del++
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(s3.Close)
	pub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", TypePNG)
		_, _ = w.Write(TinyPNG)
	}))
	t.Cleanup(pub.Close)

	got := Probe(context.Background(), Config{
		Provider:         ProviderS3,
		S3Endpoint:       s3.URL,
		S3Region:         "auto",
		S3Bucket:         "media",
		S3AccessKey:      "k",
		S3SecretKey:      "s",
		S3PublicURL:      pub.URL,
		S3ForcePathStyle: true,
		MaxBytes:         DefaultMaxBytes,
	})
	if !got.OK {
		t.Fatalf("probe=%+v", got)
	}
	if !got.PublicURLOK {
		t.Fatalf("expected public URL ok: %+v", got)
	}
	if put != 1 || del != 1 {
		t.Fatalf("put=%d del=%d", put, del)
	}
}

func TestProbeS3AccessDenied(t *testing.T) {
	s3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><Error><Code>AccessDenied</Code><Message>Access Denied</Message></Error>`))
	}))
	t.Cleanup(s3.Close)
	got := Probe(context.Background(), Config{
		Provider:         ProviderS3,
		S3Endpoint:       s3.URL,
		S3Region:         "auto",
		S3Bucket:         "media",
		S3AccessKey:      "k",
		S3SecretKey:      "s",
		S3PublicURL:      "https://cdn.example.com",
		S3ForcePathStyle: true,
		MaxBytes:         DefaultMaxBytes,
	})
	if got.OK {
		t.Fatal("expected failure")
	}
	if !strings.Contains(got.Message, "denied") && !strings.Contains(got.Message, "AccessDenied") {
		t.Fatalf("message=%q", got.Message)
	}
	if strings.Contains(got.Message, "<?xml") {
		t.Fatalf("leaked xml: %q", got.Message)
	}
}

func names(entries []os.DirEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

func TestLocalStoreDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(dir, "/uploads")
	if err != nil {
		t.Fatal(err)
	}
	obj := Object{Key: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png", ContentType: TypePNG, Data: TinyPNG}
	if _, err := store.Put(context.Background(), obj); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, obj.Key)); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(context.Background(), obj.Key); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, obj.Key)); !os.IsNotExist(err) {
		t.Fatalf("expected missing file, err=%v", err)
	}
}
