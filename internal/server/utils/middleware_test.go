package utils

import (
	"strings"
	"testing"

	"GoTodo/internal/imagehost"
	"GoTodo/internal/storage"
)

func TestImageSrcCSPDefault(t *testing.T) {
	got := ImageSrcCSP()
	if !strings.HasPrefix(got, "img-src 'self' data: blob:") {
		t.Fatalf("got %q", got)
	}
	if !strings.HasSuffix(strings.TrimSpace(got), ";") {
		t.Fatalf("directive should end with semicolon: %q", got)
	}
}

func TestImagePublicOriginInCSP(t *testing.T) {
	s := &storage.SiteSettings{
		Image: imagehost.Config{
			Provider:    imagehost.ProviderS3,
			S3PublicURL: "https://cdn.example.com/uploads",
		},
	}
	if origin := s.Image.PublicOrigin(); origin != "https://cdn.example.com" {
		t.Fatalf("origin=%q", origin)
	}
}
