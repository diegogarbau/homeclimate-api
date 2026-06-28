package solar

import (
	"math"
	"time"
)

type Orientation string

const (
	North Orientation = "N"
	South Orientation = "S"
	East  Orientation = "E"
	West  Orientation = "W"
)

type SunPosition struct {
	Azimuth  float64 // grados desde el Norte, sentido horario
	Altitude float64 // grados sobre el horizonte
}

type OrientationReport struct {
	Orientation  Orientation `json:"orientation"`
	ReceivesSun  bool        `json:"receives_sun"`
	SunIntensity string      `json:"sun_intensity"`
	ObstructedBy string      `json:"obstructed_by,omitempty"`
}

// Calculate devuelve el estado solar para cada orientación dada
func Calculate(lat, lon float64, orientations []Orientation, t time.Time) (SunPosition, []OrientationReport) {
	pos := sunPosition(lat, lon, t)
	reports := make([]OrientationReport, 0, len(orientations))

	for _, o := range orientations {
		report := OrientationReport{
			Orientation: o,
		}

		if pos.Altitude <= 0 {
			// el sol está bajo el horizonte — noche
			report.ReceivesSun = false
			report.SunIntensity = "none"
		} else {
			report.ReceivesSun = orientationFacesSun(o, pos.Azimuth)
			report.SunIntensity = intensity(report.ReceivesSun, pos.Altitude)
		}

		reports = append(reports, report)
	}

	return pos, reports
}

// sunPosition calcula azimut y altitud del sol para lat/lon y momento dados
func sunPosition(lat, lon float64, t time.Time) SunPosition {
	// día del año
	dayOfYear := float64(t.YearDay())

	// declinación solar (grados)
	declination := 23.45 * math.Sin(toRad((360.0/365.0)*(dayOfYear-81)))

	// hora solar verdadera
	utcHour := float64(t.UTC().Hour()) + float64(t.UTC().Minute())/60.0
	solarNoon := 12.0 - (lon / 15.0)
	hourAngle := 15.0 * (utcHour - solarNoon)

	latRad := toRad(lat)
	declRad := toRad(declination)
	hourRad := toRad(hourAngle)

	// altitud solar
	sinAlt := math.Sin(latRad)*math.Sin(declRad) +
		math.Cos(latRad)*math.Cos(declRad)*math.Cos(hourRad)
	altitude := toDeg(math.Asin(sinAlt))

	// azimut solar
	cosAz := (math.Sin(declRad) - math.Sin(latRad)*sinAlt) /
		(math.Cos(latRad) * math.Cos(math.Asin(sinAlt)))
	cosAz = math.Max(-1, math.Min(1, cosAz)) // clamp para evitar NaN
	azimuth := toDeg(math.Acos(cosAz))
	if hourAngle > 0 {
		azimuth = 360 - azimuth
	}

	return SunPosition{
		Azimuth:  azimuth,
		Altitude: altitude,
	}
}

// orientationFacesSun determina si una orientación recibe sol directo
// según el azimut solar (grados desde el Norte, sentido horario)
func orientationFacesSun(o Orientation, azimuth float64) bool {
	switch o {
	case North:
		return azimuth >= 330 || azimuth <= 30
	case East:
		return azimuth > 30 && azimuth <= 120
	case South:
		return azimuth > 120 && azimuth <= 240
	case West:
		return azimuth > 240 && azimuth < 330
	}
	return false
}

// intensity clasifica la intensidad solar según la altitud
func intensity(receivesSun bool, altitude float64) string {
	if !receivesSun {
		return "none"
	}
	switch {
	case altitude < 15:
		return "low"
	case altitude < 45:
		return "medium"
	default:
		return "high"
	}
}

func toRad(deg float64) float64 { return deg * math.Pi / 180 }
func toDeg(rad float64) float64 { return rad * 180 / math.Pi }