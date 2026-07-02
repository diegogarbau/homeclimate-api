package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"homeclimate-api/internal/advisor"
	"homeclimate-api/internal/geocoding"
	"homeclimate-api/internal/solar"
	"homeclimate-api/internal/weather"
)

// handleAnalyze godoc
// @Summary      Analyze climate conditions for a location
// @Description  Given coordinates or an address and home orientations, returns weather data, solar analysis per orientation, and comfort recommendations
// @Tags         analysis
// @Accept       json
// @Produce      json
// @Param        request  body      AnalyzeRequest   true  "Location and home configuration"
// @Success      200      {object}  AnalyzeResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      503      {object}  ErrorResponse
// @Router       /v1/analyze [post]
func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// resolver coordenadas — desde address o directas
	if req.Address != nil {
		geo := geocoding.NewClient()
		result, err := geo.Geocode(
			req.Address.Street,
			req.Address.PostalCode,
			req.Address.City,
			req.Address.Country,
		)
		if err != nil {
			writeError(w, http.StatusBadRequest, "address not found", err.Error())
			return
		}
		req.Latitude = result.Latitude
		req.Longitude = result.Longitude
		slog.Info("geocoded address", "display_name", result.DisplayName)
	}

	if err := validateRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	// redondear coordenadas a 2 decimales antes de enviar a Open-Meteo (~1km)
	lat := math.Round(req.Latitude*100) / 100
	lon := math.Round(req.Longitude*100) / 100

	cacheKey := fmt.Sprintf("%.2f:%.2f", lat, lon)
	var weatherData *weather.Response

	if cached, ok := s.cache.Get(cacheKey); ok {
		slog.Info("cache hit", "key", cacheKey)
		weatherData = cached.(*weather.Response)
	} else {
		slog.Info("cache miss — fetching from Open-Meteo", "key", cacheKey)
		var err error
		weatherData, err = s.weatherClient.GetCurrent(lat, lon)
		if err != nil {
			slog.Error("weather fetch failed", "error", err)
			writeError(w, http.StatusServiceUnavailable, "weather data unavailable", err.Error())
			return
		}
		s.cache.Set(cacheKey, weatherData)
	}

	now := time.Now()
	sunPos, orientationReports := solar.Calculate(req.Latitude, req.Longitude, req.Orientations, now)

	// ajustar por obstrucción de edificios colindantes
	if req.ObstructionHeightM > 0 && req.ObstructionDistanceM > 0 {
		floorH := solar.FloorHeightM(req.Floor)
		obstructionAngle := solar.ObstructionAngle(req.ObstructionHeightM, req.ObstructionDistanceM, floorH)
		orientationReports = solar.AdjustForObstruction(orientationReports, sunPos.Altitude, obstructionAngle)
		slog.Info("obstruction applied",
			"floor", req.Floor,
			"floor_height_m", floorH,
			"obstruction_angle_deg", obstructionAngle,
		)
	}

	rec, err := s.advisor.Recommend(advisor.Input{
		Temperature:   weatherData.Current.Temperature,
		Humidity:      weatherData.Current.Humidity,
		Precipitation: weatherData.Current.Precipitation,
		WindSpeed:     weatherData.Current.WindSpeed,
		IsDay:         sunPos.Altitude > 0,
		Orientations:  orientationReports,
		Floor:         req.Floor,
	})
	if err != nil {
		slog.Warn("advisor failed, continuing without recommendation", "error", err)
	}

	resp := AnalyzeResponse{
		Latitude:  lat,
		Longitude: lon,
		Timestamp: now.Format(time.RFC3339),
		Weather: WeatherData{
			Temperature:   weatherData.Current.Temperature,
			Humidity:      weatherData.Current.Humidity,
			Precipitation: weatherData.Current.Precipitation,
			WindSpeed:     weatherData.Current.WindSpeed,
			WindDirection: weatherData.Current.WindDirection,
		},
		Solar: SolarData{
			SunAzimuth:   sunPos.Azimuth,
			SunAltitude:  sunPos.Altitude,
			IsDay:        sunPos.Altitude > 0,
			Orientations: orientationReports,
		},
		Recommendation: rec,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func validateRequest(req AnalyzeRequest) error {
	if req.Latitude == 0 && req.Longitude == 0 && req.Address == nil {
		return fmt.Errorf("provide either coordinates or an address")
	}
	if req.Latitude < -90 || req.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	if len(req.Orientations) == 0 {
		return fmt.Errorf("at least one orientation is required")
	}
	for _, o := range req.Orientations {
		switch o {
		case solar.North, solar.South, solar.East, solar.West:
		default:
			return fmt.Errorf("invalid orientation: %s, valid values are N, S, E, W", o)
		}
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, msg, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg, Details: details})
}
