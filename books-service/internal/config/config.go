package config

import (
	"os"
	"time"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	AuthService AuthServiceConfig
	App         AppConfig
}

type AppConfig struct {
	Name    string
	Version string
}

type AuthServiceConfig struct {
	URL                string
	Timeout            time.Duration
	ServiceKey         string
	AuthServiceTimeout time.Duration
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
		},
		App: AppConfig{
			Name:    getEnv("SERVICE_NAME", "auth-service"),
			Version: getEnv("SERVICE_VERSION", "1.0.0"),
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
