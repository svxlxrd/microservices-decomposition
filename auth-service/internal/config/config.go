package config

import (
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	DatabaseURL DatabaseConfig
	JWTSecret      JWTConfig
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

type JWTConfig struct {
	Secret string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8081"),
			ReadTimeout:  getDuration("READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getDuration("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		DatabaseURL: DatabaseConfig{
			URL: getEnv(
				"DATABASE_URL",
				"postgres://postgres:postgres@localhost:5432/bookshelf?sslmode=disable",
			),
		},
		JWTSecret: JWTConfig{
			Secret: getEnv("JWT_SECRET", "mock"),
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
