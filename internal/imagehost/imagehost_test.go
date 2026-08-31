package imagehost

import (
	"bytes"
	"strings"
	"testing"
)

// 1x1 PNG (67 bytes).
var tinyPNG = TinyPNG

func TestDetectImage(t *testing.T) {
	t.Run("png", func(t *testing.T) {
		got, err := DetectImage(tinyPNG)
		if err != nil || got != TypePNG {
			t.Fatalf("png: type=%q err=%v", got, err)
		}
	})
	t.Run("jpeg", func(t *testing.T) {
		// JPEG SOI + APP0 so DetectContentType reports image/jpeg.
		data := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xd9}
		got, err := DetectImage(data)
		if err != nil || got != TypeJPEG {
			t.Fatalf("jpeg: type=%q err=%v", got, err)
		}
	})
	t.Run("gif", func(t *testing.T) {
		data := []byte("GIF89a")
		data = append(data, bytes.Repeat([]byte{0}, 10)...)
		got, err := DetectImage(data)
		if err != nil || got != TypeGIF {
			t.Fatalf("gif: type=%q err=%v", got, err)
		}
	})
	t.Run("webp", func(t *testing.T) {
		data := []byte("RIFF")
		data = append(data, 0, 0, 0, 0)
		data = append(data, []byte("WEBPVP8 ")...)
		got, err := DetectImage(data)
		if err != nil || got != TypeWebP {
			t.Fatalf("webp: type=%q err=%v", got, err)
		}
	})
	t.Run("svg rejected", func(t *testing.T) {
		data := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`)
		if _, err := DetectImage(data); err == nil {
			t.Fatal("expected svg rejection")
		}
	})
	t.Run("html rejected", func(t *testing.T) {
		data := []byte("<!doctype html><html><body>hi</body></html>")
		if _, err := DetectImage(data); err == nil {
			t.Fatal("expected html rejection")
		}
	})
	t.Run("too short", func(t *testing.T) {
		if _, err := DetectImage([]byte("PNG")); err == nil {
			t.Fatal("expected short payload rejection")
		}
	})
}

func TestNormalizeProvider(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ProviderNone},
		{" none ", ProviderNone},
		{"disabled", ProviderNone},
		{"S3", ProviderS3},
		{"local", ProviderLocal},
		{"minio", "minio"},
	}
	for _, tt := range tests {
		if got := NormalizeProvider(tt.in); got != tt.want {
			t.Fatalf("NormalizeProvider(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestClampMaxBytes(t *testing.T) {
	if got := ClampMaxBytes(0); got != DefaultMaxBytes {
		t.Fatalf("zero -> default, got %d", got)
	}
	if got := ClampMaxBytes(MinMaxBytes - 1); got != DefaultMaxBytes {
		t.Fatalf("below min -> default, got %d", got)
	}
	if got := ClampMaxBytes(MinMaxBytes); got != MinMaxBytes {
		t.Fatalf("min kept, got %d", got)
	}
	if got := ClampMaxBytes(MaxMaxBytes + 1); got != MaxMaxBytes {
		t.Fatalf("above max clamped, got %d", got)
	}
}

func TestConfigValidate(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		c := Config{Provider: ""}
		if msg := c.Validate(); msg != "" {
			t.Fatalf("none: %s", msg)
		}
	})
	t.Run("local defaults path", func(t *testing.T) {
		c := Config{Provider: "local"}
		if msg := c.Validate(); msg != "" {
			t.Fatalf("local: %s", msg)
		}
		if c.LocalPath != DefaultLocalPath {
			t.Fatalf("path=%q", c.LocalPath)
		}
	})
	t.Run("local rejects parent path", func(t *testing.T) {
		c := Config{Provider: "local", LocalPath: "../elsewhere"}
		if msg := c.Validate(); msg == "" {
			t.Fatal("expected path traversal rejection")
		}
	})
	t.Run("s3 requires fields", func(t *testing.T) {
		c := Config{Provider: "s3"}
		if msg := c.Validate(); !strings.Contains(msg, "endpoint") {
			t.Fatalf("want endpoint error, got %q", msg)
		}
		c.S3Endpoint = "https://nyc3.digitaloceanspaces.com"
		if msg := c.Validate(); !strings.Contains(msg, "region") {
			t.Fatalf("want region error, got %q", msg)
		}
		c.S3Region = "nyc3"
		c.S3Bucket = "media"
		c.S3AccessKey = "key"
		c.S3SecretKey = "secret"
		if msg := c.Validate(); !strings.Contains(msg, "public_url") {
			t.Fatalf("want public_url error, got %q", msg)
		}
		c.S3PublicURL = "https://media.nyc3.cdn.digitaloceanspaces.com"
		if msg := c.Validate(); msg != "" {
			t.Fatalf("complete s3 config: %s", msg)
		}
	})
	t.Run("s3 rejects relative endpoint", func(t *testing.T) {
		c := Config{
			Provider:    "s3",
			S3Endpoint:  "nyc3.digitaloceanspaces.com",
			S3Region:    "nyc3",
			S3Bucket:    "b",
			S3AccessKey: "k",
			S3SecretKey: "s",
			S3PublicURL: "https://cdn.example.com",
		}
		if msg := c.Validate(); !strings.Contains(msg, "absolute URL") {
			t.Fatalf("got %q", msg)
		}
	})
	t.Run("r2 dashboard endpoint and location code", func(t *testing.T) {
		c := Config{
			Provider:         "s3",
			S3Endpoint:       "https://abc123.r2.cloudflarestorage.com/ordryn-testing",
			S3Region:         "ENAM",
			S3Bucket:         "ordryn-testing",
			S3AccessKey:      "k",
			S3SecretKey:      " s ",
			S3PublicURL:      "https://pub-example.r2.dev",
			S3ForcePathStyle: false,
		}
		if msg := c.Validate(); msg != "" {
			t.Fatalf("validate: %s", msg)
		}
		if c.S3Endpoint != "https://abc123.r2.cloudflarestorage.com" {
			t.Fatalf("endpoint=%q", c.S3Endpoint)
		}
		if c.S3Region != "auto" {
			t.Fatalf("region=%q", c.S3Region)
		}
		if !c.S3ForcePathStyle {
			t.Fatal("expected path-style for R2")
		}
		if c.S3SecretKey != "s" {
			t.Fatalf("secret not trimmed: %q", c.S3SecretKey)
		}
	})
	t.Run("unknown provider", func(t *testing.T) {
		c := Config{Provider: "gcs"}
		if msg := c.Validate(); !strings.Contains(msg, "must be none, s3, or local") {
			t.Fatalf("got %q", msg)
		}
	})
}

func TestPublicOrigin(t *testing.T) {
	c := Config{Provider: ProviderS3, S3PublicURL: "https://cdn.example.com/uploads/"}
	if got := c.PublicOrigin(); got != "https://cdn.example.com" {
		t.Fatalf("origin=%q", got)
	}
	c = Config{Provider: ProviderLocal, LocalPublicBase: "/uploads"}
	if got := c.PublicOrigin(); got != "" {
		t.Fatalf("relative base should not yield origin, got %q", got)
	}
}

func TestSafeObjectKey(t *testing.T) {
	if !SafeObjectKey("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png") {
		t.Fatal("uuid png should be safe")
	}
	if SafeObjectKey("../etc/passwd") {
		t.Fatal("traversal should be rejected")
	}
	if SafeObjectKey("dir/file.png") {
		t.Fatal("slash should be rejected")
	}
	if SafeObjectKey("file.exe") {
		t.Fatal("exe should be rejected")
	}
}

func TestPrepareSizeLimit(t *testing.T) {
	cfg := Config{Provider: ProviderLocal, MaxBytes: MinMaxBytes, LocalPath: t.TempDir()}
	big := make([]byte, MinMaxBytes+1)
	copy(big, tinyPNG)
	if _, err := Prepare(cfg, big); err == nil {
		t.Fatal("expected oversize rejection")
	}
	ok, err := Prepare(cfg, tinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	if ok.ContentType != TypePNG || !strings.HasSuffix(ok.Key, ".png") {
		t.Fatalf("object=%+v", ok)
	}
}

func TestPrepareDisabled(t *testing.T) {
	if _, err := Prepare(Config{}, tinyPNG); err != ErrDisabled {
		t.Fatalf("err=%v", err)
	}
}

func TestJoinPublicURL(t *testing.T) {
	if got := JoinPublicURL("https://cdn.example.com/", "a.png"); got != "https://cdn.example.com/a.png" {
		t.Fatalf("got %q", got)
	}
	if got := JoinPublicURL("/uploads", "a.png"); got != "/uploads/a.png" {
		t.Fatalf("got %q", got)
	}
}

func TestExtForContentType(t *testing.T) {
	if ExtForContentType(TypeJPEG) != ".jpg" {
		t.Fatal(ExtForContentType(TypeJPEG))
	}
	if ExtForContentType("image/svg+xml") != "" {
		t.Fatal("svg should have no hosted extension")
	}
}
