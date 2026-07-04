package advisor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// claudeModel is the model used for the real advisor: fast and cost-efficient,
// and it supports structured outputs so we can guarantee valid JSON.
const claudeModel = "claude-haiku-4-5"

const anthropicVersion = "2023-06-01"

const defaultAnthropicBaseURL = "https://api.anthropic.com"

// ClaudeAdvisor calls the Anthropic Messages API to produce recommendations.
// It talks to the API over raw net/http to keep the module dependency-free.
//
// Privacy: the advisor receives only aggregated climate and solar data via
// advisor.Input — never coordinates or addresses. That boundary is enforced by
// the Input type itself, which has no location fields.
type ClaudeAdvisor struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	// fallback produces a recommendation when the API call fails, so the
	// endpoint always returns actionable advice.
	fallback Advisor
}

// NewClaudeAdvisor returns an Advisor backed by the Anthropic Messages API.
// If apiKey is empty, it transparently falls back to the rule-based Mock.
func NewClaudeAdvisor(apiKey string) Advisor {
	return &ClaudeAdvisor{
		apiKey:     apiKey,
		model:      claudeModel,
		baseURL:    defaultAnthropicBaseURL,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		fallback:   NewMock(),
	}
}

// NewClaudeAdvisorWithBaseURL is like NewClaudeAdvisor but allows overriding
// the Anthropic base URL — used in tests to point at a mock server.
func NewClaudeAdvisorWithBaseURL(apiKey, baseURL string) Advisor {
	return &ClaudeAdvisor{
		apiKey:     apiKey,
		model:      claudeModel,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		fallback:   NewMock(),
	}
}

// ---- Anthropic Messages API wire types (subset we use) ----

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicOutputFormat struct {
	Type   string `json:"type"` // "json_schema"
	Schema any    `json:"schema"`
}

type anthropicOutputConfig struct {
	Format anthropicOutputFormat `json:"format"`
}

type anthropicRequest struct {
	Model        string                `json:"model"`
	MaxTokens    int                   `json:"max_tokens"`
	System       string                `json:"system,omitempty"`
	Messages     []anthropicMessage    `json:"messages"`
	OutputConfig anthropicOutputConfig `json:"output_config"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content    []anthropicContentBlock `json:"content"`
	StopReason string                  `json:"stop_reason"`
}

func (c *ClaudeAdvisor) Recommend(input Input) (*Recommendation, error) {
	if c.apiKey == "" {
		return c.fallback.Recommend(input)
	}

	rec, err := c.recommendViaAPI(input)
	if err != nil {
		// Never fail the request over an advisor hiccup — degrade to rules.
		return c.fallback.Recommend(input)
	}
	return rec, nil
}

func (c *ClaudeAdvisor) recommendViaAPI(input Input) (*Recommendation, error) {
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: 2048,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: buildUserPrompt(input)},
		},
		OutputConfig: anthropicOutputConfig{
			Format: anthropicOutputFormat{Type: "json_schema", Schema: recommendationSchema},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic returned status %d", resp.StatusCode)
	}

	var apiResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if apiResp.StopReason == "refusal" {
		return nil, fmt.Errorf("anthropic refused the request")
	}

	text := firstText(apiResp.Content)
	if text == "" {
		return nil, fmt.Errorf("empty response content")
	}

	var rec Recommendation
	if err := json.Unmarshal([]byte(text), &rec); err != nil {
		return nil, fmt.Errorf("parse recommendation JSON: %w", err)
	}
	if rec.Actions == nil {
		rec.Actions = []Action{}
	}
	return &rec, nil
}

func firstText(blocks []anthropicContentBlock) string {
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			return b.Text
		}
	}
	return ""
}

const systemPrompt = `You are HomeClimate, an assistant that helps people keep their home comfortable.

Given current weather and per-facade solar exposure, recommend concrete, actionable steps the resident can take right now: open or close windows, raise or lower blinds, deploy or retract awnings, and collect laundry from the terrace when rain or strong wind is coming.

Rules:
- Base every recommendation only on the data provided. Do not invent readings.
- Prioritize safety: strong wind or rain means close windows and retract awnings.
- Reference specific orientations (N/S/E/W) when advising on blinds or awnings.
- comfort_level must be one of: "good", "acceptable", "poor".
- Each action's priority must be one of: "high", "medium", "low".
- Keep the summary to one or two sentences. Keep reasons short and practical.
- Respond only with the structured JSON requested.`

func buildUserPrompt(input Input) string {
	var b strings.Builder
	timeOfDay := "night"
	if input.IsDay {
		timeOfDay = "day"
	}
	fmt.Fprintf(&b, "Current conditions (%s):\n", timeOfDay)
	fmt.Fprintf(&b, "- Temperature: %.1f°C\n", input.Temperature)
	fmt.Fprintf(&b, "- Humidity: %d%%\n", input.Humidity)
	fmt.Fprintf(&b, "- Precipitation: %.1f mm\n", input.Precipitation)
	fmt.Fprintf(&b, "- Wind speed: %.1f km/h\n", input.WindSpeed)
	fmt.Fprintf(&b, "- Floor: %d (0 = ground floor)\n", input.Floor)

	if len(input.Orientations) == 0 {
		b.WriteString("\nNo facade orientations provided.\n")
	} else {
		b.WriteString("\nSolar exposure per facade:\n")
		for _, o := range input.Orientations {
			obstruction := ""
			if o.ObstructedBy != "" {
				obstruction = fmt.Sprintf(", obstructed by %s", o.ObstructedBy)
			}
			fmt.Fprintf(&b, "- %s: receives_sun=%t, intensity=%s%s\n",
				o.Orientation, o.ReceivesSun, o.SunIntensity, obstruction)
		}
	}

	b.WriteString("\nRecommend what the resident should do now.")
	return b.String()
}

// recommendationSchema constrains the model output to the Recommendation shape.
// Structured outputs require additionalProperties:false and explicit required lists.
var recommendationSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"properties": map[string]any{
		"summary": map[string]any{"type": "string"},
		"comfort_level": map[string]any{
			"type": "string",
			"enum": []string{"good", "acceptable", "poor"},
		},
		"actions": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"action": map[string]any{"type": "string"},
					"reason": map[string]any{"type": "string"},
					"priority": map[string]any{
						"type": "string",
						"enum": []string{"high", "medium", "low"},
					},
				},
				"required": []string{"action", "reason", "priority"},
			},
		},
	},
	"required": []string{"summary", "comfort_level", "actions"},
}
