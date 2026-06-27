package api

import (
	"encoding/json"
	"net/http"

	"homeclimate-api/config"
)

func NewRouter(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /v1/analyze", handleAnalyze)

	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "1.0.0",
	})
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// placeholder — lo implementamos en el siguiente paso
	json.NewEncoder(w).Encode(map[string]string{
		"status": "not implemented yet",
	})
}