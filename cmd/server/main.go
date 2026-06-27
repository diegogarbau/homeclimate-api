package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"homeclimate-api/config"
	"homeclimate-api/internal/api"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	router := api.NewRouter(cfg)

	slog.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}