package solar_test

import (
	"testing"

	"homeclimate-api/internal/solar"
)

func TestObstructionAngle(t *testing.T) {
	tests := []struct {
		name            string
		buildingHeightM float64
		distanceM       float64
		floorHeightM    float64
		wantAngleApprox float64
	}{
		{"typical obstruction", 15, 8, 1, 60.3},
		{"floor above building", 10, 5, 15, 0},
		{"zero distance", 15, 0, 1, 0},
		{"floor equals building height", 10, 5, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := solar.ObstructionAngle(tt.buildingHeightM, tt.distanceM, tt.floorHeightM)
			diff := got - tt.wantAngleApprox
			if diff < -1 || diff > 1 {
				t.Errorf("ObstructionAngle(%v, %v, %v) = %.1f, want ~%.1f",
					tt.buildingHeightM, tt.distanceM, tt.floorHeightM, got, tt.wantAngleApprox)
			}
		})
	}
}

func TestFloorHeightM(t *testing.T) {
	tests := []struct {
		floor int
		want  float64
	}{
		{0, 0.5},
		{-1, 0.5},
		{1, 4.0},
		{3, 10.0},
	}

	for _, tt := range tests {
		got := solar.FloorHeightM(tt.floor)
		if got != tt.want {
			t.Errorf("FloorHeightM(%d) = %.1f, want %.1f", tt.floor, got, tt.want)
		}
	}
}

func TestAdjustForObstruction_NoObstruction(t *testing.T) {
	reports := []solar.OrientationReport{
		{Orientation: solar.South, ReceivesSun: true, SunIntensity: "high"},
	}

	got := solar.AdjustForObstruction(reports, 45, 0)

	if !got[0].ReceivesSun {
		t.Error("expected ReceivesSun to remain true when obstructionAngle is 0")
	}
}

func TestAdjustForObstruction_SunBlocked(t *testing.T) {
	reports := []solar.OrientationReport{
		{Orientation: solar.South, ReceivesSun: true, SunIntensity: "high"},
		{Orientation: solar.North, ReceivesSun: false, SunIntensity: "none"},
	}

	// sun altitude (20) is below the obstruction angle (40) — South should be blocked
	got := solar.AdjustForObstruction(reports, 20, 40)

	if got[0].ReceivesSun {
		t.Error("expected South to be obstructed when sun altitude < obstruction angle")
	}
	if got[0].SunIntensity != "none" {
		t.Errorf("expected intensity 'none' after obstruction, got %s", got[0].SunIntensity)
	}
	if got[0].ObstructedBy == "" {
		t.Error("expected ObstructedBy to be set after obstruction")
	}
	// North was never receiving sun — should remain unaffected
	if got[1].ObstructedBy != "" {
		t.Error("expected North to remain unaffected since it never received sun")
	}
}

func TestAdjustForObstruction_SunAboveObstruction(t *testing.T) {
	reports := []solar.OrientationReport{
		{Orientation: solar.South, ReceivesSun: true, SunIntensity: "high"},
	}

	// sun altitude (60) is above the obstruction angle (40) — South remains sunny
	got := solar.AdjustForObstruction(reports, 60, 40)

	if !got[0].ReceivesSun {
		t.Error("expected South to still receive sun when altitude > obstruction angle")
	}
}
