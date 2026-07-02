// @title           HomeClimate API
// @version         1.0
// @description     Given a location and home orientations, analyzes weather and solar conditions to recommend comfort actions (open/close windows, blinds, awnings).
// @termsOfService  http://swagger.io/terms/

// @contact.name   Diego García Bautista
// @contact.url    https://linkedin.com/in/diegogarbau
// @contact.email  diego.garbau@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

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
