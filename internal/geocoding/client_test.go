package geocoding

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient points the client at a mock Nominatim server.
func newTestClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func TestGeocode_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"lat":"40.4168","lon":"-3.7038","display_name":"Madrid, Spain"}]`))
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)
	res, err := c.Geocode("Gran Via 1", "28013", "Madrid", "Spain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Latitude != 40.4168 || res.Longitude != -3.7038 {
		t.Errorf("coordinates = (%v, %v), want (40.4168, -3.7038)", res.Latitude, res.Longitude)
	}
	if res.DisplayName != "Madrid, Spain" {
		t.Errorf("display name = %q, want %q", res.DisplayName, "Madrid, Spain")
	}
}

func TestGeocode_StripsHouseNumberAndDefaultsCountry(t *testing.T) {
	var gotQuery string
	var gotUserAgent string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		gotUserAgent = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`[{"lat":"40.41","lon":"-3.70","display_name":"x"}]`))
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)
	// country empty -> should default to Spain; house number must be stripped
	if _, err := c.Geocode("Calle Mayor 42", "28013", "Madrid", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(gotQuery, "42") {
		t.Errorf("query %q leaked the house number", gotQuery)
	}
	if !strings.Contains(gotQuery, "Calle Mayor") {
		t.Errorf("query %q missing street name", gotQuery)
	}
	if !strings.Contains(gotQuery, "Spain") {
		t.Errorf("query %q should default country to Spain", gotQuery)
	}
	if gotUserAgent == "" {
		t.Error("Nominatim requires an identifying User-Agent, none was sent")
	}
}

func TestGeocode_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)
	if _, err := c.Geocode("Nowhere", "00000", "Atlantis", "Spain"); err == nil {
		t.Fatal("expected error for empty result set, got nil")
	}
}

func TestGeocode_MalformedResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)
	if _, err := c.Geocode("Gran Via", "28013", "Madrid", "Spain"); err == nil {
		t.Fatal("expected decode error for malformed response, got nil")
	}
}

func TestSanitizeStreet(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Calle Mayor 42", "Calle Mayor"},
		{"Gran Via, 10", "Gran Via"},
		{"Calle Mayor", "Calle Mayor"}, // no number -> unchanged
		{"Avenida de America 123 ", "Avenida de America"},
		{"123", "123"}, // only digits -> original returned
		{"", ""},
	}
	for _, tc := range cases {
		if got := sanitizeStreet(tc.in); got != tc.want {
			t.Errorf("sanitizeStreet(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Fatal("expected NewClient to return a non-nil client")
	}
}
