package config

import (
	"os"
)

type Config struct {
	Port            string
	OpenMeteoURL    string
	ClaudeAPIKey    string
	CacheTTLMinutes int
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		OpenMeteoURL:    getEnv("OPEN_METEO_URL", "https://api.open-meteo.com/v1"),
		ClaudeAPIKey:    getEnv("CLAUDE_API_KEY", ""),
		CacheTTLMinutes: 30,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
