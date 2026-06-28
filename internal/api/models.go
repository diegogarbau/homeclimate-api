package api

import "homeclimate-api/internal/solar"

type AnalyzeRequest struct {
	Latitude     float64              `json:"latitude"`
	Longitude    float64              `json:"longitude"`
	Orientations []solar.Orientation  `json:"orientations"`
}

type WeatherData struct {
	Temperature   float64 `json:"temperature_celsius"`
	Humidity      int     `json:"humidity_percent"`
	Precipitation float64 `json:"precipitation_mm"`
	WindSpeed     float64 `json:"wind_speed_kmh"`
	WindDirection int     `json:"wind_direction_degrees"`
}

type SolarData struct {
	SunAzimuth  float64                    `json:"sun_azimuth_degrees"`
	SunAltitude float64                    `json:"sun_altitude_degrees"`
	IsDay       bool                       `json:"is_day"`
	Orientations []solar.OrientationReport `json:"orientations"`
}

type AnalyzeResponse struct {
	Latitude  float64     `json:"latitude"`
	Longitude float64     `json:"longitude"`
	Timestamp string      `json:"timestamp"`
	Weather   WeatherData `json:"weather"`
	Solar     SolarData   `json:"solar"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}