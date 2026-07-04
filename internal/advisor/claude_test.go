package advisor_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"homeclimate-api/internal/advisor"
	"homeclimate-api/internal/solar"
)

func TestClaudeAdvisor_FallsBackWhenNoAPIKey(t *testing.T) {
	a := advisor.NewClaudeAdvisor("")

	rec, err := a.Recommend(advisor.Input{
		Temperature: 25,
		Humidity:    50,
		IsDay:       true,
	})

	if err != nil {
		t.Fatalf("expected no error with empty API key (should fallback), got: %v", err)
	}
	if rec == nil {
		t.Fatal("expected a recommendation from fallback mock")
	}
}

func TestClaudeAdvisor_FallsBackOnAPIError(t *testing.T) {
	// servidor mock que simula un error de la API de Anthropic
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mock.Close()

	a := advisor.NewClaudeAdvisorWithBaseURL("sk-ant-fake-key", mock.URL)

	rec, err := a.Recommend(advisor.Input{
		Temperature: 30,
		Humidity:    40,
		IsDay:       true,
	})

	if err != nil {
		t.Fatalf("expected fallback to succeed without error, got: %v", err)
	}
	if rec == nil {
		t.Fatal("expected fallback recommendation when API errors")
	}
}

func TestClaudeAdvisor_SuccessfulResponse(t *testing.T) {
	mockResponse := map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": `{"summary":"Hot day, take precautions.","comfort_level":"acceptable","actions":[{"action":"Close blinds","reason":"Direct sun","priority":"high"}]}`,
			},
		},
		"stop_reason": "end_turn",
	}

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mock.Close()

	a := advisor.NewClaudeAdvisorWithBaseURL("sk-ant-fake-key", mock.URL)

	rec, err := a.Recommend(advisor.Input{
		Temperature: 32,
		Humidity:    30,
		IsDay:       true,
		Orientations: []solar.OrientationReport{
			{Orientation: solar.South, ReceivesSun: true, SunIntensity: "high"},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Summary != "Hot day, take precautions." {
		t.Errorf("expected summary from mocked API, got: %s", rec.Summary)
	}
	if rec.Comfort != "acceptable" {
		t.Errorf("expected comfort 'acceptable', got: %s", rec.Comfort)
	}
	if len(rec.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(rec.Actions))
	}
	if rec.Actions[0].Action != "Close blinds" {
		t.Errorf("expected action 'Close blinds', got: %s", rec.Actions[0].Action)
	}
}

func TestClaudeAdvisor_RefusalFallsBack(t *testing.T) {
	mockResponse := map[string]any{
		"content":     []map[string]any{},
		"stop_reason": "refusal",
	}

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mock.Close()

	a := advisor.NewClaudeAdvisorWithBaseURL("sk-ant-fake-key", mock.URL)

	rec, err := a.Recommend(advisor.Input{Temperature: 20, IsDay: true})

	if err != nil {
		t.Fatalf("expected fallback to succeed without error, got: %v", err)
	}
	if rec == nil {
		t.Fatal("expected fallback recommendation on refusal")
	}
}

func TestClaudeAdvisor_MalformedJSONFallsBack(t *testing.T) {
	mockResponse := map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": "not valid json{{{"},
		},
		"stop_reason": "end_turn",
	}

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mock.Close()

	a := advisor.NewClaudeAdvisorWithBaseURL("sk-ant-fake-key", mock.URL)

	rec, err := a.Recommend(advisor.Input{Temperature: 20, IsDay: true})

	if err != nil {
		t.Fatalf("expected fallback to succeed without error, got: %v", err)
	}
	if rec == nil {
		t.Fatal("expected fallback recommendation on malformed JSON")
	}
}
