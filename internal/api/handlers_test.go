package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"homeclimate-api/internal/advisor"
	"homeclimate-api/internal/cache"
	"homeclimate-api/internal/weather"
)

const mockWeatherBody = `{
  "latitude": 40.42,
  "longitude": -3.7,
  "current": {
    "temperature_2m": 31.4,
    "relative_humidity_2m": 29,
    "precipitation": 0,
    "wind_speed_10m": 8.3,
    "wind_direction_10m": 185
  }
}`

// newTestServer wires a Server against a mock Open-Meteo endpoint.
// It returns the server and a pointer to the upstream hit counter.
func newTestServer(t *testing.T) (*Server, *int32) {
	t.Helper()
	var hits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte(mockWeatherBody))
	}))
	t.Cleanup(ts.Close)

	s := &Server{
		weatherClient: weather.NewClient(ts.URL),
		cache:         cache.New(30 * time.Minute),
		advisor:       advisor.NewMock(),
	}
	return s, &hits
}

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
	if body["version"] == "" {
		t.Error("version missing from health response")
	}
}

func TestHandleAnalyze_Success(t *testing.T) {
	s, hits := newTestServer(t)

	reqBody := `{"latitude":40.4168,"longitude":-3.7038,"orientations":["S","E"],"floor":1}`
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", strings.NewReader(reqBody))
	rr := httptest.NewRecorder()

	s.handleAnalyze(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rr.Code, rr.Body.String())
	}
	var resp AnalyzeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	// coordinates rounded to 2 decimals for privacy
	if resp.Latitude != 40.42 || resp.Longitude != -3.7 {
		t.Errorf("coords = (%v, %v), want (40.42, -3.7)", resp.Latitude, resp.Longitude)
	}
	if resp.Weather.Temperature != 31.4 {
		t.Errorf("temperature = %v, want 31.4", resp.Weather.Temperature)
	}
	if len(resp.Solar.Orientations) != 2 {
		t.Errorf("orientations = %d, want 2", len(resp.Solar.Orientations))
	}
	if resp.Recommendation == nil {
		t.Error("recommendation missing")
	}
	if *hits != 1 {
		t.Errorf("upstream hits = %d, want 1", *hits)
	}
}

func TestHandleAnalyze_UsesCache(t *testing.T) {
	s, hits := newTestServer(t)

	reqBody := `{"latitude":40.4168,"longitude":-3.7038,"orientations":["S"],"floor":0}`
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/analyze", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()
		s.handleAnalyze(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, rr.Code)
		}
	}
	// same rounded coordinates -> only the first request should hit upstream
	if *hits != 1 {
		t.Errorf("upstream hits = %d, want 1 (cache should serve the rest)", *hits)
	}
}

func TestHandleAnalyze_InvalidJSON(t *testing.T) {
	s, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", strings.NewReader(`{bad`))
	rr := httptest.NewRecorder()
	s.handleAnalyze(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestHandleAnalyze_ValidationErrors(t *testing.T) {
	s, _ := newTestServer(t)

	cases := map[string]string{
		"no location":           `{"orientations":["S"]}`,
		"no orientations":       `{"latitude":40.4,"longitude":-3.7,"orientations":[]}`,
		"invalid orientation":   `{"latitude":40.4,"longitude":-3.7,"orientations":["X"]}`,
		"latitude out of range": `{"latitude":200,"longitude":-3.7,"orientations":["S"]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/analyze", strings.NewReader(body))
			rr := httptest.NewRecorder()
			s.handleAnalyze(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400 (body: %s)", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleAnalyze_WeatherUnavailable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(ts.Close)

	s := &Server{
		weatherClient: weather.NewClient(ts.URL),
		cache:         cache.New(30 * time.Minute),
		advisor:       advisor.NewMock(),
	}

	reqBody := `{"latitude":10.12,"longitude":20.34,"orientations":["S"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", strings.NewReader(reqBody))
	rr := httptest.NewRecorder()
	s.handleAnalyze(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rr.Code)
	}
}

func TestHandleOpenAPISpec_IsValidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	rr := httptest.NewRecorder()

	handleOpenAPISpec(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var spec map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("openapi spec is not valid JSON: %v", err)
	}
	if spec["openapi"] == nil {
		t.Error("spec missing openapi version field")
	}
}
