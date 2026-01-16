package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type Config struct {
	BasePath      string `json:"basePath"`
	UseHTTPS      bool   `json:"useHttps"`
	AssetVersion  string `json:"assetVersion,omitempty"`
	FromEmail     string `json:"from_email,omitempty"`
	ShowChangelog bool   `json:"showChangelog,omitempty"`
}

var Cfg Config

func Load() {
	// Try to open repo config file first
	data, err := os.ReadFile("config/config.json")
	if err != nil {
		log.Printf("config: config file not found or unreadable: %v; falling back to env/defaults", err)
		loadFromEnv()
		return
	}
	if err := json.Unmarshal(data, &Cfg); err != nil {
		log.Printf("config: error decoding config file: %v; falling back to env/defaults", err)
		loadFromEnv()
		return
	}

	// If the config file doesn't explicitly include showChangelog, allow env or default
	if !strings.Contains(string(data), "\"showChangelog\"") {
		if os.Getenv("SHOW_CHANGELOG") == "false" {
			Cfg.ShowChangelog = false
		} else if os.Getenv("SHOW_CHANGELOG") == "true" {
			Cfg.ShowChangelog = true
		} else {
			// default to true when not specified
			Cfg.ShowChangelog = true
		}
	}

	// Fill from env where fields are missing
	if Cfg.BasePath == "" {
		Cfg.BasePath = os.Getenv("BASE_PATH")
	}
	if Cfg.BasePath == "" {
		Cfg.BasePath = "/"
	}
	if Cfg.AssetVersion == "" {
		if v := os.Getenv("ASSET_VERSION"); v != "" {
			Cfg.AssetVersion = v
		} else {
			Cfg.AssetVersion = "20251130"
		}
	}
	// FromEmail may be set in config.json; allow env override
	if Cfg.FromEmail == "" {
		if v := os.Getenv("FROM_EMAIL"); v != "" {
			Cfg.FromEmail = v
		} else {
			Cfg.FromEmail = "no-reply@example.com"
		}
	}
	// Also allow env override for UseHTTPS
	if os.Getenv("USE_HTTPS") == "true" {
		Cfg.UseHTTPS = true
	}
}

func loadFromEnv() {
	Cfg.BasePath = os.Getenv("BASE_PATH")
	if Cfg.BasePath == "" {
		Cfg.BasePath = "/"
	}
	if v := os.Getenv("ASSET_VERSION"); v != "" {
		Cfg.AssetVersion = v
	} else {
		Cfg.AssetVersion = "20251130"
	}
	if os.Getenv("USE_HTTPS") == "true" {
		Cfg.UseHTTPS = true
	} else {
		Cfg.UseHTTPS = false
	}
	// Default show changelog to true unless explicitly disabled
	if os.Getenv("SHOW_CHANGELOG") == "false" {
		Cfg.ShowChangelog = false
	} else {
		Cfg.ShowChangelog = true
	}
}
