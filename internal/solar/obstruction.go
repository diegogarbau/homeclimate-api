package solar

import "math"

// ObstructionAngle calcula el ángulo de obstrucción en grados
// que produce un edificio colindante dado su altura y distancia.
// Un edificio de 10m a 5m de distancia produce un ángulo de ~63°
// Por debajo de ese ángulo de altitud solar, el edificio tapa el sol.
func ObstructionAngle(buildingHeightM, distanceM, floorHeightM float64) float64 {
	if distanceM <= 0 {
		return 0
	}
	effectiveHeight := buildingHeightM - floorHeightM
	if effectiveHeight <= 0 {
		return 0 // la planta está por encima del edificio colindante
	}
	return math.Atan(effectiveHeight/distanceM) * 180 / math.Pi
}

// FloorHeightM estima la altura en metros de una planta
// usando una altura estándar de 3m por planta
func FloorHeightM(floor int) float64 {
	if floor <= 0 {
		return 0.5 // bajo: ventana a ~0.5m del suelo
	}
	return float64(floor)*3.0 + 1.0 // plantas superiores
}

// AdjustForObstruction modifica los reports de orientación
// teniendo en cuenta el ángulo de obstrucción del edificio colindante
func AdjustForObstruction(reports []OrientationReport, sunAltitude, obstructionAngle float64) []OrientationReport {
	if obstructionAngle <= 0 {
		return reports
	}
	adjusted := make([]OrientationReport, len(reports))
	copy(adjusted, reports)
	for i, r := range adjusted {
		if r.ReceivesSun && sunAltitude < obstructionAngle {
			adjusted[i].ReceivesSun = false
			adjusted[i].SunIntensity = "none"
			adjusted[i].ObstructedBy = "adjacent building"
		}
	}
	return adjusted
}