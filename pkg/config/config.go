package config

import (
	"os"
	"strings"
)

type Config struct {
	// Server
	Port           string
	AllowedOrigins string

	// Database
	DatabaseURL string

	// OpenAI
	OpenAIAPIKey string

	// Environment
	Environment string
	Debug       bool
}

func New() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		OpenAIAPIKey:   getEnv("OPENAI_API_KEY", ""),
		Environment:    getEnv("ENVIRONMENT", "development"),
		Debug:          getEnv("DEBUG", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}
