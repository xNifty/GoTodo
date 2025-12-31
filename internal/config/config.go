package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	BasePath     string `json:"basePath"`
	UseHTTPS     bool   `json:"useHttps"`
	AssetVersion string `json:"assetVersion,omitempty"`
	FromEmail    string `json:"from_email,omitempty"`
}

var Cfg Config

func Load() {
	// Try to open repo config file first
	f, err := os.Open("config/config.json")
	if err != nil {
		log.Printf("config: config file not found or unreadable: %v; falling back to env/defaults", err)
		loadFromEnv()
		return
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&Cfg); err != nil {
		log.Printf("config: error decoding config file: %v; falling back to env/defaults", err)
		loadFromEnv()
		return
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
}
