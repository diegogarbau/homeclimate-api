# HomeClimate API

> A Go REST API that turns a home's location and window orientations into actionable comfort advice — when to open windows, lower blinds, deploy awnings, or bring the laundry in.

[![Go](https://img.shields.io/badge/Go-1.22.4-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenAPI 3.0](https://img.shields.io/badge/OpenAPI-3.0-6BA539?logo=openapiinitiative&logoColor=white)](https://spec.openapis.org/oas/v3.0.3)

HomeClimate API takes a home's location (coordinates **or** a street address) and the orientations of its windows/facades (N/S/E/W). It fetches real-time weather, computes the sun's position and per-orientation solar exposure, and asks an advisor to recommend concrete comfort actions.

---

## Features

- **Location flexibility** — accepts raw coordinates or a street address (geocoded via Nominatim/OpenStreetMap).
- **Solar modeling** — computes sun azimuth/altitude from first principles and determines which facades receive direct sun, at what intensity, accounting for neighboring-building obstruction and floor height.
- **Real-time weather** — temperature, humidity, precipitation, wind speed and direction from Open-Meteo.
- **Actionable advice** — a rule-based advisor (Claude-powered advisor planned) returns prioritized actions with reasons and an overall comfort level.
- **Privacy by design** — coordinates are rounded to ~1 km before any external call, street numbers are stripped before geocoding, and nothing is persisted.
- **Zero external dependencies** — the module is standard-library only.

---

## Architecture

```
                    ┌─────────────────────────────────────────────┐
                    │              HomeClimate API                 │
   POST /v1/analyze │                                              │
   ────────────────▶│  ┌────────┐   ┌─────────┐   ┌────────────┐   │
                    │  │  api   │──▶│geocoding │──▶│ Nominatim  │───┼──▶ OSM
                    │  │ router │   └─────────┘   └────────────┘   │
                    │  │ handler│   ┌─────────┐   ┌────────────┐   │
                    │  │        │──▶│ weather │──▶│ Open-Meteo │───┼──▶ api.open-meteo.com
                    │  │        │   └─────────┘   └────────────┘   │
                    │  │        │   ┌─────────┐                    │
                    │  │        │──▶│  solar  │  (pure calculation)│
                    │  │        │   └─────────┘                    │
                    │  │        │   ┌─────────┐   ┌────────────┐   │
                    │  │        │──▶│ advisor │   │ cache (TTL)│   │
                    │  └────────┘   └─────────┘   └────────────┘   │
                    └─────────────────────────────────────────────┘
```

| Package | Responsibility |
|---|---|
| `internal/api` | HTTP router, request/response models, `/v1/analyze` and `/health` handlers, Swagger UI |
| `internal/weather` | Open-Meteo client (30s timeout) |
| `internal/solar` | Sun position, per-orientation exposure, obstruction/floor adjustment |
| `internal/geocoding` | Nominatim client; strips street number before sending |
| `internal/advisor` | `Advisor` interface + rule-based mock implementation |
| `internal/cache` | Concurrent-safe in-memory cache with TTL and background cleanup |
| `config` | Environment-based configuration |

---

## API

| Method | Path | Description |
|---|---|---|
| `GET`  | `/health` | Liveness check → `{"status":"ok","version":"1.0.0"}` |
| `POST` | `/v1/analyze` | Full climate + solar analysis with recommendations |
| `GET`  | `/docs` | Swagger UI |
| `GET`  | `/docs/openapi.json` | OpenAPI 3.0 specification |

### `POST /v1/analyze`

**Request** — provide either `latitude`/`longitude` or `address`:

```json
{
  "latitude": 40.4168,
  "longitude": -3.7038,
  "orientations": ["S", "E", "N", "W"],
  "floor": 1,
  "obstruction_height_m": 12,
  "obstruction_distance_m": 6
}
```

Using an address instead of coordinates:

```json
{
  "address": {
    "street": "Gran Via 1",
    "postal_code": "28013",
    "city": "Madrid",
    "country": "Spain"
  },
  "orientations": ["S", "E"]
}
```

**Response:**

```json
{
  "latitude": 40.42,
  "longitude": -3.7,
  "timestamp": "2026-06-28T14:01:06+02:00",
  "weather": {
    "temperature_celsius": 31.4,
    "humidity_percent": 29,
    "precipitation_mm": 0,
    "wind_speed_kmh": 8.3,
    "wind_direction_degrees": 185
  },
  "solar": {
    "sun_azimuth_degrees": 169.3,
    "sun_altitude_degrees": 72.6,
    "is_day": true,
    "orientations": [
      { "orientation": "S", "receives_sun": true, "sun_intensity": "high", "obstructed_by": "" },
      { "orientation": "E", "receives_sun": false, "sun_intensity": "none", "obstructed_by": "" }
    ]
  },
  "recommendation": {
    "summary": "Hot conditions: 31.4°C. Manage sunlight and ventilation carefully.",
    "comfort_level": "acceptable",
    "actions": [
      { "action": "Close blinds on S-facing windows", "reason": "Direct high-intensity sun", "priority": "high" },
      { "action": "Deploy awning on S-facing terrace", "reason": "Reduce heat gain", "priority": "medium" }
    ]
  }
}
```

### Try it

```bash
curl -s -X POST http://localhost:8080/v1/analyze \
  -H 'Content-Type: application/json' \
  -d '{
    "latitude": 40.4168,
    "longitude": -3.7038,
    "orientations": ["S","E","N","W"],
    "floor": 1
  }' | jq
```

---

## Getting started

### Prerequisites

- Go 1.22.4+ (to run from source), or Docker.

### Run locally

```bash
git clone https://github.com/diegogarbau/homeclimate-api.git
cd homeclimate-api

cp .env.example .env   # optional — sensible defaults are built in

go run ./cmd/server
# server starting on :8080
```

### Run with Docker

```bash
docker compose up --build
```

---

## Configuration

Configuration is read from environment variables (see [`.env.example`](.env.example)):

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `OPEN_METEO_URL` | `https://api.open-meteo.com/v1` | Open-Meteo base URL |
| `CLAUDE_API_KEY` | _(empty)_ | Enables the Claude-powered advisor (planned). Empty → rule-based mock advisor |

The in-memory cache uses a fixed 30-minute TTL.

---

## Solar & privacy notes

- **Sun position** is derived from latitude, longitude and UTC time using standard solar geometry (declination, hour angle, altitude, azimuth) with mean solar time — accurate enough for comfort estimates.
- **Orientation sectors:** N = 330°–30°, E = 30°–120°, S = 120°–240°, W = 240°–330°.
- **Sun intensity:** altitude < 15° = low, < 45° = medium, ≥ 45° = high.
- **Obstruction:** given a neighboring building's height and distance, the obstruction angle is `arctan(effective_height / distance)`; floor height is estimated as `floor × 3.0 + 1.0 m`. Sun is obstructed when its altitude is below the obstruction angle.
- **Privacy:** coordinates are rounded to 2 decimals (~1 km) before external calls; the street number is stripped before geocoding; no request data is persisted.

---

## Testing

```bash
go test ./...
```

The suite covers the weather client, solar calculations, cache behavior, and advisor rules.

---

## Roadmap

- **V1 (current)** — public, stateless API with a single analysis endpoint.
- **V2** — user registration (JWT) and PostgreSQL-backed home profiles; analyze by saved profile.
- **V3** — weather-alert subscriptions with a background poller (storm/high-wind/temperature-drop notifications).

---

## License

MIT © Diego García Bautista
