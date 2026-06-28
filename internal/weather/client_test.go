package weather_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"homeclimate-api/internal/weather"
)

func TestGetCurrent_Success(t *testing.T) {
	// servidor mock que devuelve una respuesta fija
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"latitude":  40.4168,
			"longitude": -3.7038,
			"elevation": 666.0,
			"timezone":  "Europe/Madrid",
			"current": map[string]any{
				"temperature_2m":       33.6,
				"relative_humidity_2m": 21,
				"precipitation":        0.0,
				"wind_speed_10m":       9.2,
				"wind_direction_10m":   249,
				"time":                 "2026-06-27T20:15",
			},
		})
	}))
	defer mock.Close()

	client := weather.NewClient(mock.URL)
	result, err := client.GetCurrent(40.4168, -3.7038)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Current.Temperature != 33.6 {
		t.Errorf("expected temperature 33.6, got %.1f", result.Current.Temperature)
	}
	if result.Current.Humidity != 21 {
		t.Errorf("expected humidity 21, got %d", result.Current.Humidity)
	}
}

func TestGetCurrent_ServerError(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mock.Close()

	client := weather.NewClient(mock.URL)
	_, err := client.GetCurrent(40.4168, -3.7038)

	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}

func TestGetCurrent_Timeout(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer mock.Close()

	// cliente con timeout muy corto para el test
	c := &http.Client{Timeout: 100 * time.Millisecond}
	_ = c

	client := weather.NewClientWithTimeout(mock.URL, 100*time.Millisecond)
	_, err := client.GetCurrent(40.4168, -3.7038)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}