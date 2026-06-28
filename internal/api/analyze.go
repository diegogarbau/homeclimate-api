package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"homeclimate-api/internal/solar"
	"homeclimate-api/internal/weather"
)

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := validateRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	cacheKey := fmt.Sprintf("%.4f:%.4f", req.Latitude, req.Longitude)
	var weatherData *weather.Response

	if cached, ok := s.cache.Get(cacheKey); ok {
		slog.Info("cache hit", "key", cacheKey)
		weatherData = cached.(*weather.Response)
	} else {
		slog.Info("cache miss — fetching from Open-Meteo", "key", cacheKey)
		var err error
		weatherData, err = s.weatherClient.GetCurrent(req.Latitude, req.Longitude)
		if err != nil {
			slog.Error("weather fetch failed", "error", err)
			writeError(w, http.StatusServiceUnavailable, "weather data unavailable", err.Error())
			return
		}
		s.cache.Set(cacheKey, weatherData)
	}

	now := time.Now()
	sunPos, orientationReports := solar.Calculate(req.Latitude, req.Longitude, req.Orientations, now)

	resp := AnalyzeResponse{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
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
			IsDay:         sunPos.Altitude > 0,
			Orientations: orientationReports,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func validateRequest(req AnalyzeRequest) error {
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