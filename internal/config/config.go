package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	StorageMemory   = "memory"
	StoragePostgres = "postgres"
)

type Config struct {
	HTTPAddr        string
	BaseURL         string
	StorageType     string
	PostgresDSN     string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	RequestTimeout  time.Duration
}

func Load() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.HTTPAddr, "http-addr", envOrDefault("HTTP_ADDR", ":8080"), "http listen address")
	flag.StringVar(&cfg.BaseURL, "base-url", envOrDefault("BASE_URL", "http://localhost:8080"), "external base url for responses")
	flag.StringVar(&cfg.StorageType, "storage", envOrDefault("STORAGE_TYPE", StorageMemory), "storage implementation: memory|postgres")
	flag.StringVar(&cfg.PostgresDSN, "postgres-dsn", envOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@postgres:5432/url_shortener?sslmode=disable"), "postgres connection string")

	flag.DurationVar(&cfg.ShutdownTimeout, "shutdown-timeout", envDurationOrDefault("SHUTDOWN_TIMEOUT", 10*time.Second), "graceful shutdown timeout")
	flag.DurationVar(&cfg.ReadTimeout, "read-timeout", envDurationOrDefault("READ_TIMEOUT", 5*time.Second), "http read timeout")
	flag.DurationVar(&cfg.WriteTimeout, "write-timeout", envDurationOrDefault("WRITE_TIMEOUT", 5*time.Second), "http write timeout")
	flag.DurationVar(&cfg.IdleTimeout, "idle-timeout", envDurationOrDefault("IDLE_TIMEOUT", 30*time.Second), "http idle timeout")
	flag.DurationVar(&cfg.RequestTimeout, "request-timeout", envDurationOrDefault("REQUEST_TIMEOUT", 3*time.Second), "per-request timeout")

	flag.Parse()

	cfg.HTTPAddr = strings.TrimSpace(cfg.HTTPAddr)
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.StorageType = strings.ToLower(strings.TrimSpace(cfg.StorageType))
	cfg.PostgresDSN = strings.TrimSpace(cfg.PostgresDSN)

	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("http address is required")
	}
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return Config{}, err
	}

	switch cfg.StorageType {
	case StorageMemory:
	case StoragePostgres:
		if cfg.PostgresDSN == "" {
			return Config{}, fmt.Errorf("postgres dsn is required when storage=postgres")
		}
	default:
		return Config{}, fmt.Errorf("unsupported storage type %q", cfg.StorageType)
	}

	return cfg, nil
}

func validateBaseURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("base url is required")
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("base url must include scheme and host")
	}

	return nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return d
}
