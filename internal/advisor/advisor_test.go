package advisor_test

import (
	"testing"

	"homeclimate-api/internal/advisor"
	"homeclimate-api/internal/solar"
)

func TestMock_HotSunnyDay(t *testing.T) {
	a := advisor.NewMock()
	rec, err := a.Recommend(advisor.Input{
		Temperature:   33.0,
		Humidity:      30,
		Precipitation: 0,
		WindSpeed:     5,
		IsDay:         true,
		Orientations: []solar.OrientationReport{
			{Orientation: solar.South, ReceivesSun: true, SunIntensity: "high"},
			{Orientation: solar.East, ReceivesSun: false, SunIntensity: "none"},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Comfort == "good" {
		t.Error("comfort should not be good at 33°C")
	}
	if len(rec.Actions) == 0 {
		t.Error("expected at least one action for hot sunny conditions")
	}
}

func TestMock_CoolNight(t *testing.T) {
	a := advisor.NewMock()
	rec, err := a.Recommend(advisor.Input{
		Temperature:   20.0,
		Humidity:      45,
		Precipitation: 0,
		WindSpeed:     8,
		IsDay:         false,
		Orientations: []solar.OrientationReport{
			{Orientation: solar.South, ReceivesSun: false, SunIntensity: "none"},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasOpenWindows := false
	for _, a := range rec.Actions {
		if a.Action == "Open windows" {
			hasOpenWindows = true
		}
	}
	if !hasOpenWindows {
		t.Error("expected open windows recommendation for cool night")
	}
}

func TestMock_StrongWind(t *testing.T) {
	a := advisor.NewMock()
	rec, err := a.Recommend(advisor.Input{
		Temperature:   22.0,
		Humidity:      50,
		Precipitation: 0,
		WindSpeed:     50,
		IsDay:         true,
		Orientations:  []solar.OrientationReport{},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasCloseWindows := false
	hasRetractAwning := false
	for _, a := range rec.Actions {
		if a.Action == "Close all windows" {
			hasCloseWindows = true
		}
		if a.Action == "Retract all awnings" {
			hasRetractAwning = true
		}
	}
	if !hasCloseWindows {
		t.Error("expected close windows for strong wind")
	}
	if !hasRetractAwning {
		t.Error("expected retract awnings for strong wind")
	}
}

func TestMock_Rain(t *testing.T) {
	a := advisor.NewMock()
	rec, err := a.Recommend(advisor.Input{
		Temperature:   18.0,
		Humidity:      80,
		Precipitation: 5.0,
		WindSpeed:     15,
		IsDay:         true,
		Orientations:  []solar.OrientationReport{},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasCloseWindows := false
	for _, a := range rec.Actions {
		if a.Action == "Close all windows" {
			hasCloseWindows = true
		}
	}
	if !hasCloseWindows {
		t.Error("expected close windows recommendation for rain")
	}
}