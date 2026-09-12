package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port           string
	MongoURI       string
	MongoDatabase  string
	AccessSecret   string
	RefreshSecret  string
	FrontendOrigin string
	Environment    string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:           value("PORT", "8080"),
		MongoURI:       value("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:  value("MONGODB_DATABASE", "kabadiconnect"),
		AccessSecret:   os.Getenv("JWT_ACCESS_SECRET"),
		RefreshSecret:  os.Getenv("JWT_REFRESH_SECRET"),
		FrontendOrigin: value("FRONTEND_ORIGIN", "http://localhost:5173"),
		Environment:    value("ENVIRONMENT", "production"),
		AccessTTL:      15 * time.Minute,
		RefreshTTL:     7 * 24 * time.Hour,
	}
	if len(cfg.AccessSecret) < 32 || len(cfg.RefreshSecret) < 32 {
		return Config{}, fmt.Errorf("JWT secrets must each be at least 32 characters")
	}
	if strings.EqualFold(cfg.AccessSecret, cfg.RefreshSecret) {
		return Config{}, fmt.Errorf("JWT access and refresh secrets must be different")
	}
	if cfg.Environment != "development" && cfg.Environment != "test" && cfg.Environment != "production" {
		return Config{}, fmt.Errorf("ENVIRONMENT must be development, test or production")
	}
	origin, err := url.Parse(cfg.FrontendOrigin)
	if err != nil || origin.Host == "" || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" || (origin.Scheme != "http" && origin.Scheme != "https") {
		return Config{}, fmt.Errorf("FRONTEND_ORIGIN must be an exact HTTP(S) origin")
	}
	if cfg.Environment != "development" && (origin.Scheme != "https" || strings.Contains(cfg.AccessSecret, "replace-with") || strings.Contains(cfg.RefreshSecret, "replace-with")) {
		return Config{}, fmt.Errorf("non-development environments require HTTPS and non-placeholder JWT secrets")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
