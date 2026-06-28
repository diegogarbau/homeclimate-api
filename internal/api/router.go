package api

import (
	"encoding/json"
	"net/http"
	"time"

	"homeclimate-api/config"
	"homeclimate-api/internal/advisor"
	"homeclimate-api/internal/cache"
	"homeclimate-api/internal/weather"
)

type Server struct {
	weatherClient *weather.Client
	cache         *cache.Cache
	advisor       advisor.Advisor
}

func NewRouter(cfg *config.Config) http.Handler {
	s := &Server{
		weatherClient: weather.NewClient(cfg.OpenMeteoURL),
		cache:         cache.New(time.Duration(cfg.CacheTTLMinutes) * time.Minute),
		advisor:       advisor.NewMock(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /v1/analyze", s.handleAnalyze)

	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "1.0.0",
	})
}