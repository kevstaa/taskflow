package config

import "os"

type Config struct {
	Port        string
	JWTSecret   string
	DatabaseURL string
	RedisURL    string
	Environment string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080:"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", ""),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
