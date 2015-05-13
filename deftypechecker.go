package appdeftransf

import (
	"math"
)

// defTypeChecker implementations are supposed to detect a certain app
// definition.
type defTypeChecker interface {
	// Parse receives the user given byte slice representing the app definition
	// to parse detect. It returns its own DefType in any case. The second return
	// value is a probability that indicates how likely it is to have its own
	// definition type detected. When returning 0.0 there is no chance that the
	// given app definition is detected. When 100.0 is returned there is
	// definitly the given app definition detected.
	Parse(b []byte) (DefType, float64)
}

func round(f float64) float64 {
	shift := math.Pow(10, float64(2))
	return math.Floor((f*shift)+.5) / shift
}
