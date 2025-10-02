package utils

import (
	"fmt"
)

// PresentableDistance formats distance information for display
func PresentableDistance(stopsFromCurStop int, distToCurrentStopKM float64, distToImmedNextStopKM float64) string {
	D := 0.5
	N := 3
	E := 0.5
	P := 500.0
	T := 100.0

	distToImmedNextStopMi := distToImmedNextStopKM * MilesPerKilometer
	distToCurrentStopMi := distToCurrentStopKM * MilesPerKilometer

	showMiles := (distToImmedNextStopMi > D) || ((stopsFromCurStop > N) && (distToCurrentStopMi > E))
	if showMiles {
		return fmt.Sprintf("%g mile%s", distToCurrentStopMi, ternary(distToCurrentStopMi == 1, "", "s"))
	}
	if stopsFromCurStop == 0 {
		distFt := distToCurrentStopMi * FeetPerMile
		if distFt < T {
			return "at stop"
		}
		if distFt < P {
			return "approaching"
		}
	}
	if stopsFromCurStop == 1 {
		return "1 stop"
	}
	return fmt.Sprintf("%d stops", stopsFromCurStop)
}

func ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
