package solar_test

import (
	"testing"
	"time"

	"homeclimate-api/internal/solar"
)

func TestCalculate_SummerNoon_Madrid(t *testing.T) {
	// 11:45 UTC = mediodía solar real en Madrid (lon=-3.7, UTC+2)
	madrid := time.Date(2026, 6, 27, 11, 45, 0, 0, time.UTC)
	lat, lon := 40.4168, -3.7038

	pos, reports := solar.Calculate(lat, lon, []solar.Orientation{
		solar.North, solar.South, solar.East, solar.West,
	}, madrid)

	if pos.Altitude <= 0 {
		t.Errorf("expected sun above horizon at noon, got altitude %.2f", pos.Altitude)
	}

	for _, r := range reports {
		switch r.Orientation {
		case solar.South:
			if !r.ReceivesSun {
				t.Errorf("South should receive sun at noon in summer, azimuth=%.1f", pos.Azimuth)
			}
			if r.SunIntensity != "high" {
				t.Errorf("South intensity should be high at noon, got %s", r.SunIntensity)
			}
		case solar.North:
			if r.ReceivesSun {
				t.Errorf("North should NOT receive direct sun at noon in summer")
			}
		}
	}
}

func TestCalculate_Night(t *testing.T) {
	// Medianoche — ninguna orientación recibe sol
	midnight := time.Date(2026, 6, 27, 22, 0, 0, 0, time.UTC)
	lat, lon := 40.4168, -3.7038

	pos, reports := solar.Calculate(lat, lon, []solar.Orientation{
		solar.North, solar.South, solar.East, solar.West,
	}, midnight)

	if pos.Altitude > 0 {
		t.Errorf("expected sun below horizon at midnight, got altitude %.2f", pos.Altitude)
	}

	for _, r := range reports {
		if r.ReceivesSun {
			t.Errorf("orientation %s should not receive sun at night", r.Orientation)
		}
		if r.SunIntensity != "none" {
			t.Errorf("intensity should be none at night, got %s for %s", r.SunIntensity, r.Orientation)
		}
	}
}

func TestCalculate_Morning_East(t *testing.T) {
	// 08:00 UTC (10:00 local) — sol al este
	morning := time.Date(2026, 6, 27, 6, 0, 0, 0, time.UTC)
	lat, lon := 40.4168, -3.7038

	_, reports := solar.Calculate(lat, lon, []solar.Orientation{
		solar.East, solar.West,
	}, morning)

	for _, r := range reports {
		if r.Orientation == solar.East && !r.ReceivesSun {
			t.Errorf("East should receive sun in the morning")
		}
		if r.Orientation == solar.West && r.ReceivesSun {
			t.Errorf("West should NOT receive sun in the morning")
		}
	}
}
