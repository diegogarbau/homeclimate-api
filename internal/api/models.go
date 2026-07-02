package api

import (
	"homeclimate-api/internal/advisor"
	"homeclimate-api/internal/solar"
)

type AnalyzeRequest struct {
	// Opción A — coordenadas directas
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`

	// Opción B — dirección nominal (alternativa a coordenadas)
	Address *AddressInput `json:"address,omitempty"`

	// Orientaciones de la vivienda
	Orientations []solar.Orientation `json:"orientations"`

	// Altura de la vivienda
	Floor                int     `json:"floor"`                            // 0=bajo, 1=primero...
	ObstructionHeightM   float64 `json:"obstruction_height_m,omitempty"`   // altura edificios colindantes
	ObstructionDistanceM float64 `json:"obstruction_distance_m,omitempty"` // distancia a esos edificios
}

type AddressInput struct {
	Street     string `json:"street"` // calle y número
	PostalCode string `json:"postal_code"`
	City       string `json:"city"`
	Country    string `json:"country,omitempty"` // default "Spain"
}

type WeatherData struct {
	Temperature   float64 `json:"temperature_celsius"`
	Humidity      int     `json:"humidity_percent"`
	Precipitation float64 `json:"precipitation_mm"`
	WindSpeed     float64 `json:"wind_speed_kmh"`
	WindDirection int     `json:"wind_direction_degrees"`
}

type SolarData struct {
	SunAzimuth   float64                   `json:"sun_azimuth_degrees"`
	SunAltitude  float64                   `json:"sun_altitude_degrees"`
	IsDay        bool                      `json:"is_day"`
	Orientations []solar.OrientationReport `json:"orientations"`
}

type AnalyzeResponse struct {
	Latitude       float64                 `json:"latitude"`
	Longitude      float64                 `json:"longitude"`
	Timestamp      string                  `json:"timestamp"`
	Weather        WeatherData             `json:"weather"`
	Solar          SolarData               `json:"solar"`
	Recommendation *advisor.Recommendation `json:"recommendation"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
