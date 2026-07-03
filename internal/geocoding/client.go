package geocoding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://nominatim.openstreetmap.org",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type Result struct {
	Latitude    float64
	Longitude   float64
	DisplayName string
}

type nominatimResponse struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

// Geocode convierte una dirección en coordenadas.
// Por privacidad, nunca enviamos el número de portal — solo calle, CP y ciudad.
func (c *Client) Geocode(street, postalCode, city, country string) (*Result, error) {
	if country == "" {
		country = "Spain"
	}

	// anonimizamos: eliminamos el número de portal de la calle
	sanitizedStreet := sanitizeStreet(street)

	query := fmt.Sprintf("%s, %s %s, %s", sanitizedStreet, postalCode, city, country)

	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("limit", "1")
	params.Set("addressdetails", "0") // no devolver detalles de dirección

	reqURL := fmt.Sprintf("%s/search?%s", c.baseURL, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// Nominatim requiere User-Agent identificativo
	req.Header.Set("User-Agent", "homeclimate-api/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocoding request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var results []nominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode geocoding response: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("address not found: %s", query)
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude %q: %w", results[0].Lat, err)
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude %q: %w", results[0].Lon, err)
	}

	return &Result{
		Latitude:    lat,
		Longitude:   lon,
		DisplayName: results[0].DisplayName,
	}, nil
}

// sanitizeStreet elimina el número de portal para proteger la privacidad
// "Calle Mayor 42" -> "Calle Mayor"
// "Gran Via, 10" -> "Gran Via"
func sanitizeStreet(street string) string {
	// eliminamos todo lo que sea número al final de la cadena
	runes := []rune(street)
	i := len(runes) - 1

	// saltamos espacios y comas del final
	for i >= 0 && (runes[i] == ' ' || runes[i] == ',') {
		i--
	}
	// saltamos dígitos del final
	for i >= 0 && runes[i] >= '0' && runes[i] <= '9' {
		i--
	}
	// saltamos espacios y comas intermedios
	for i >= 0 && (runes[i] == ' ' || runes[i] == ',') {
		i--
	}

	result := runes[:i+1]
	if len(result) == 0 {
		return street // si queda vacío devolvemos el original
	}
	return string(result)
}
