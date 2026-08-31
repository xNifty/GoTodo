package storage

import (
	"testing"

	"GoTodo/internal/imagehost"
)

func TestImageHostingConfigDefaults(t *testing.T) {
	s := &SiteSettings{Image: imagehost.Config{Provider: "local"}}
	cfg, err := s.ImageHostingConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != imagehost.ProviderLocal {
		t.Fatalf("provider=%q", cfg.Provider)
	}
	if cfg.LocalPath != imagehost.DefaultLocalPath {
		t.Fatalf("path=%q", cfg.LocalPath)
	}
	if cfg.MaxBytes != imagehost.DefaultMaxBytes {
		t.Fatalf("max=%d", cfg.MaxBytes)
	}
}

func TestImageHostingConfigDisabled(t *testing.T) {
	s := &SiteSettings{}
	cfg, err := s.ImageHostingConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled() {
		t.Fatal("expected disabled")
	}
}
