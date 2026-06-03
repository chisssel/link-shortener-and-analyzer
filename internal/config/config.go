package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort string
	BaseURL    string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	DatabaseURL      string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	CacheTTL int
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "urlshortener"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "urlshortener_secret"),
		PostgresDB:       getEnv("POSTGRES_DB", "urlshortener"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		CacheTTL: getEnvInt("CACHE_TTL", 3600),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
