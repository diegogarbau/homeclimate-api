package config_test

import (
	"testing"

	"homeclimate-api/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.OpenMeteoURL != "https://api.open-meteo.com/v1" {
		t.Errorf("expected default Open-Meteo URL, got %s", cfg.OpenMeteoURL)
	}
	if cfg.CacheTTLMinutes != 30 {
		t.Errorf("expected default cache TTL 30, got %d", cfg.CacheTTLMinutes)
	}
	if cfg.ClaudeAPIKey != "" {
		t.Errorf("expected empty Claude API key by default, got %s", cfg.ClaudeAPIKey)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("OPEN_METEO_URL", "https://custom-weather.example.com")
	t.Setenv("CLAUDE_API_KEY", "sk-ant-test-key")

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090 from env, got %s", cfg.Port)
	}
	if cfg.OpenMeteoURL != "https://custom-weather.example.com" {
		t.Errorf("expected custom Open-Meteo URL from env, got %s", cfg.OpenMeteoURL)
	}
	if cfg.ClaudeAPIKey != "sk-ant-test-key" {
		t.Errorf("expected Claude API key from env, got %s", cfg.ClaudeAPIKey)
	}
}
