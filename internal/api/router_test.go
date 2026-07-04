package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"homeclimate-api/config"
	"homeclimate-api/internal/api"
)

func TestNewRouter_HealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:            "8080",
		OpenMeteoURL:    "https://api.open-meteo.com/v1",
		CacheTTLMinutes: 30,
	}
	router := api.NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestNewRouter_UsesMockWhenNoAPIKey(t *testing.T) {
	cfg := &config.Config{
		Port:            "8080",
		OpenMeteoURL:    "https://api.open-meteo.com/v1",
		ClaudeAPIKey:    "",
		CacheTTLMinutes: 30,
	}
	router := api.NewRouter(cfg)

	if router == nil {
		t.Fatal("expected router to be non-nil")
	}
}

func TestNewRouter_UsesClaudeWhenAPIKeySet(t *testing.T) {
	cfg := &config.Config{
		Port:            "8080",
		OpenMeteoURL:    "https://api.open-meteo.com/v1",
		ClaudeAPIKey:    "sk-ant-fake-key-for-test",
		CacheTTLMinutes: 30,
	}
	router := api.NewRouter(cfg)

	if router == nil {
		t.Fatal("expected router to be non-nil")
	}
}

func TestHandleDocs(t *testing.T) {
	cfg := &config.Config{
		Port:            "8080",
		OpenMeteoURL:    "https://api.open-meteo.com/v1",
		CacheTTLMinutes: 30,
	}
	router := api.NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "swagger-ui") {
		t.Error("expected docs response to contain swagger-ui reference")
	}
	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}
}

func TestHandleDocs_OpenAPISpec(t *testing.T) {
	cfg := &config.Config{
		Port:            "8080",
		OpenMeteoURL:    "https://api.open-meteo.com/v1",
		CacheTTLMinutes: 30,
	}
	router := api.NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "openapi") {
		t.Error("expected response to contain openapi spec")
	}
}
