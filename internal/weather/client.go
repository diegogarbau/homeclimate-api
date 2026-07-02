package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewClientWithTimeout(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type CurrentWeather struct {
	Temperature   float64 `json:"temperature_2m"`
	Humidity      int     `json:"relative_humidity_2m"`
	Precipitation float64 `json:"precipitation"`
	WindSpeed     float64 `json:"wind_speed_10m"`
	WindDirection int     `json:"wind_direction_10m"`
	Time          string  `json:"time"`
}

type Response struct {
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Elevation float64        `json:"elevation"`
	Timezone  string         `json:"timezone"`
	Current   CurrentWeather `json:"current"`
}

func (c *Client) GetCurrent(lat, lon float64) (*Response, error) {
	url := fmt.Sprintf(
		"%s/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,precipitation,wind_speed_10m,wind_direction_10m&timezone=auto",
		c.baseURL, lat, lon,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("open-meteo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open-meteo returned status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
