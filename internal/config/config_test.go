package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()
	if cfg.ServerPort != "8080" {
		t.Errorf("expected default port 8080, got %q", cfg.ServerPort)
	}
	if cfg.RedisHost != "localhost" {
		t.Errorf("expected default redis host localhost, got %q", cfg.RedisHost)
	}
	if cfg.CacheTTL != 3600 {
		t.Errorf("expected default cache TTL 3600, got %d", cfg.CacheTTL)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("REDIS_HOST", "myredis")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("REDIS_HOST")
	}()
	cfg := Load()
	if cfg.ServerPort != "9090" {
		t.Errorf("expected 9090, got %q", cfg.ServerPort)
	}
	if cfg.RedisHost != "myredis" {
		t.Errorf("expected myredis, got %q", cfg.RedisHost)
	}
}

func TestLoad_RedisDB(t *testing.T) {
	os.Setenv("REDIS_DB", "1")
	defer os.Unsetenv("REDIS_DB")
	cfg := Load()
	if cfg.RedisDB != 1 {
		t.Errorf("expected redis DB 1, got %d", cfg.RedisDB)
	}
}

func TestDatabaseURL_Default(t *testing.T) {
	cfg := Load()
	url := cfg.DatabaseURL()
	if url == "" {
		t.Fatal("empty database URL")
	}
	if url != "postgres://urlshortener:urlshortener_secret@localhost:5432/urlshortener?sslmode=disable" {
		t.Errorf("unexpected database URL: %q", url)
	}
}

func TestDatabaseURL_Custom(t *testing.T) {
	os.Setenv("POSTGRES_USER", "admin")
	os.Setenv("POSTGRES_PASSWORD", "pass")
	os.Setenv("POSTGRES_HOST", "pg.example.com")
	defer func() {
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_HOST")
	}()
	cfg := Load()
	url := cfg.DatabaseURL()
	expected := "postgres://admin:pass@pg.example.com:5432/urlshortener?sslmode=disable"
	if url != expected {
		t.Errorf("got %q, want %q", url, expected)
	}
}

func TestLoad_BaseURL(t *testing.T) {
	os.Setenv("BASE_URL", "https://short.example.com")
	defer os.Unsetenv("BASE_URL")
	cfg := Load()
	if cfg.BaseURL != "https://short.example.com" {
		t.Errorf("expected custom base URL, got %q", cfg.BaseURL)
	}
}

func TestLoad_CacheTTL(t *testing.T) {
	os.Setenv("CACHE_TTL", "7200")
	defer os.Unsetenv("CACHE_TTL")
	cfg := Load()
	if cfg.CacheTTL != 7200 {
		t.Errorf("expected TTL 7200, got %d", cfg.CacheTTL)
	}
}
