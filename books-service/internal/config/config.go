package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	AuthService AuthServiceConfig
}

type AuthServiceConfig struct {
	URL                string
	Timeout            time.Duration
	ServiceKey         string
	AuthServiceTimeout time.Duration
	AuthServiceRetries int
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8082"),
			ReadTimeout:  getDuration("READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getDuration("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL: getEnv(
				"DATABASE_URL",
				"postgres://postgres:postgres@localhost:5433/books?sslmode=disable",
			),
		},
		AuthService: AuthServiceConfig{
			URL:                getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
			Timeout:            getDuration("AUTH_SERVICE_TIMEOUT", 5*time.Second),
			ServiceKey:         getEnv("SERVICE_KEY", ""),
			AuthServiceTimeout: getDuration("MAX_REQUEST_TIMEOUT", 10*time.Second),
			AuthServiceRetries: getInt("MAX_RETRIES", 3),
		},
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}

	return def
}
