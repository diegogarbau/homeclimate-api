package advisor

import (
	"fmt"
	"homeclimate-api/internal/solar"
)

// Advisor es la interfaz que tanto el mock como la implementación real cumplen
type Advisor interface {
	Recommend(input Input) (*Recommendation, error)
}

type Input struct {
	Temperature   float64
	Humidity      int
	Precipitation float64
	WindSpeed     float64
	IsDay         bool
	Orientations  []solar.OrientationReport
	Floor         int
}

type Action struct {
	Action   string `json:"action"`
	Reason   string `json:"reason"`
	Priority string `json:"priority"` // "high", "medium", "low"
}

type Recommendation struct {
	Summary string   `json:"summary"`
	Actions []Action `json:"actions"`
	Comfort string   `json:"comfort_level"` // "good", "acceptable", "poor"
}

// Mock implementa Advisor con respuestas basadas en reglas simples
type Mock struct{}

func NewMock() Advisor {
	return &Mock{}
}

func (m *Mock) Recommend(input Input) (*Recommendation, error) {
	rec := &Recommendation{
		Actions: []Action{},
	}

	// lógica de confort
	comfort, summary := assessComfort(input)
	rec.Comfort = comfort
	rec.Summary = summary

	// reglas para acciones
	rec.Actions = append(rec.Actions, windowActions(input)...)
	rec.Actions = append(rec.Actions, blindActions(input)...)
	rec.Actions = append(rec.Actions, awningActions(input)...)

	return rec, nil
}

func assessComfort(input Input) (string, string) {
	switch {
	case input.Temperature > 30 && input.Humidity > 60:
		return "poor", fmt.Sprintf("Very uncomfortable conditions: %.1f°C with %d%% humidity. Take immediate action to cool your home.", input.Temperature, input.Humidity)
	case input.Temperature > 30:
		return "acceptable", fmt.Sprintf("Hot conditions: %.1f°C. Manage sunlight and ventilation carefully.", input.Temperature)
	case input.Temperature < 10:
		return "poor", fmt.Sprintf("Cold conditions: %.1f°C. Keep heat inside.", input.Temperature)
	case input.Temperature >= 18 && input.Temperature <= 26 && input.Humidity < 60:
		return "good", fmt.Sprintf("Comfortable conditions: %.1f°C and %d%% humidity. Minimal action needed.", input.Temperature, input.Humidity)
	default:
		return "acceptable", fmt.Sprintf("Moderate conditions: %.1f°C. Some adjustments recommended.", input.Temperature)
	}
}

func windowActions(input Input) []Action {
	actions := []Action{}

	// ventilación nocturna o con temperatura fresca
	if !input.IsDay && input.Temperature < 24 {
		actions = append(actions, Action{
			Action:   "Open windows",
			Reason:   fmt.Sprintf("Night air at %.1f°C will cool your home naturally. Take advantage of cross-ventilation.", input.Temperature),
			Priority: "high",
		})
		return actions
	}

	// viento fuerte — cerrar ventanas
	if input.WindSpeed > 40 {
		actions = append(actions, Action{
			Action:   "Close all windows",
			Reason:   fmt.Sprintf("Wind speed %.1f km/h may cause damage or bring dust inside.", input.WindSpeed),
			Priority: "high",
		})
		return actions
	}

	// lluvia — cerrar ventanas
	if input.Precipitation > 0 {
		actions = append(actions, Action{
			Action:   "Close all windows",
			Reason:   fmt.Sprintf("%.1f mm precipitation detected. Close windows to prevent water entry.", input.Precipitation),
			Priority: "high",
		})
		return actions
	}

	// calor — cerrar ventanas si exterior más caliente
	if input.Temperature > 28 && input.IsDay {
		actions = append(actions, Action{
			Action:   "Keep windows closed",
			Reason:   "Outside temperature is high. Keeping windows closed will maintain cooler indoor air.",
			Priority: "medium",
		})
	} else if input.Temperature < 26 {
		actions = append(actions, Action{
			Action:   "Open windows for ventilation",
			Reason:   fmt.Sprintf("Pleasant temperature of %.1f°C outside. Natural ventilation will improve indoor comfort.", input.Temperature),
			Priority: "low",
		})
	}

	return actions
}

func blindActions(input Input) []Action {
	actions := []Action{}

	if !input.IsDay {
		return actions
	}

	for _, o := range input.Orientations {
		if o.ReceivesSun && o.SunIntensity == "high" {
			actions = append(actions, Action{
				Action:   fmt.Sprintf("Close blinds on %s-facing windows", o.Orientation),
				Reason:   fmt.Sprintf("Direct high-intensity sunlight from the %s will significantly heat the room.", o.Orientation),
				Priority: "high",
			})
		} else if o.ReceivesSun && o.SunIntensity == "medium" {
			actions = append(actions, Action{
				Action:   fmt.Sprintf("Partially close blinds on %s-facing windows", o.Orientation),
				Reason:   fmt.Sprintf("Moderate sunlight from the %s. Partial shading will reduce heat gain.", o.Orientation),
				Priority: "medium",
			})
		}
	}

	return actions
}

func awningActions(input Input) []Action {
	actions := []Action{}

	// recoger toldos con viento fuerte
	if input.WindSpeed > 30 {
		actions = append(actions, Action{
			Action:   "Retract all awnings",
			Reason:   fmt.Sprintf("Wind speed of %.1f km/h may damage extended awnings.", input.WindSpeed),
			Priority: "high",
		})
		return actions
	}

	// recoger toldos con lluvia
	if input.Precipitation > 0 {
		actions = append(actions, Action{
			Action:   "Retract awnings",
			Reason:   "Rain detected. Retract awnings to prevent water accumulation and damage.",
			Priority: "medium",
		})
		return actions
	}

	// desplegar toldos con sol intenso
	if input.IsDay {
		for _, o := range input.Orientations {
			if o.ReceivesSun && o.SunIntensity == "high" {
				actions = append(actions, Action{
					Action:   fmt.Sprintf("Deploy awning on %s-facing terrace", o.Orientation),
					Reason:   fmt.Sprintf("High solar intensity from the %s. Awning will provide effective shade.", o.Orientation),
					Priority: "medium",
				})
			}
		}
	}

	return actions
}
