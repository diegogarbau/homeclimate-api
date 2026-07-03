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
	// Use the Claude-powered advisor when an API key is configured; otherwise
	// fall back to the rule-based mock so the API works with zero config.
	var adv advisor.Advisor
	if cfg.ClaudeAPIKey != "" {
		adv = advisor.NewClaudeAdvisor(cfg.ClaudeAPIKey)
	} else {
		adv = advisor.NewMock()
	}

	s := &Server{
		weatherClient: weather.NewClient(cfg.OpenMeteoURL),
		cache:         cache.New(time.Duration(cfg.CacheTTLMinutes) * time.Minute),
		advisor:       adv,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /v1/analyze", s.handleAnalyze)
	mux.HandleFunc("/docs", handleDocs)
	mux.HandleFunc("/docs/openapi.json", handleOpenAPISpec)

	return mux
}

// handleHealth godoc
// @Summary      Health check
// @Description  Returns service status and version
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "1.0.0",
	})
}

func handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>HomeClimate API Docs</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "/docs/openapi.json",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  })
</script>
</body>
</html>`))
}

func handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{
  "openapi": "3.0.0",
  "info": {
    "title": "HomeClimate API",
    "version": "1.0.0",
    "description": "Given a location and home orientations, analyzes weather and solar conditions to recommend comfort actions.",
    "contact": {
      "name": "Diego García Bautista",
      "url": "https://linkedin.com/in/diegogarbau",
      "email": "contact@diegogarbau.dev"
    },
    "license": { "name": "MIT" }
  },
  "servers": [{ "url": "http://localhost:8080" }],
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check",
        "tags": ["system"],
        "responses": {
          "200": {
            "description": "Service status",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": { "type": "string", "example": "ok" },
                    "version": { "type": "string", "example": "1.0.0" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/v1/analyze": {
      "post": {
        "summary": "Analyze climate conditions",
        "description": "Given coordinates or address and home orientations, returns weather, solar analysis and comfort recommendations.",
        "tags": ["analysis"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "latitude":  { "type": "number", "example": 40.4168 },
                  "longitude": { "type": "number", "example": -3.7038 },
                  "address": {
                    "type": "object",
                    "properties": {
                      "street":      { "type": "string", "example": "Gran Via 1" },
                      "postal_code": { "type": "string", "example": "28013" },
                      "city":        { "type": "string", "example": "Madrid" },
                      "country":     { "type": "string", "example": "Spain" }
                    }
                  },
                  "orientations": {
                    "type": "array",
                    "items": { "type": "string", "enum": ["N","S","E","W"] },
                    "example": ["S","E","N","W"]
                  },
                  "floor": { "type": "integer", "example": 1 },
                  "obstruction_height_m":   { "type": "number", "example": 12 },
                  "obstruction_distance_m": { "type": "number", "example": 6 }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Climate analysis and recommendations" },
          "400": { "description": "Invalid request" },
          "503": { "description": "Weather data unavailable" }
        }
      }
    }
  }
}`))
}
