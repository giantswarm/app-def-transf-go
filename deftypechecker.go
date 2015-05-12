package appdeftransf

import (
	"math"
)

type defTypeChecker interface {
	Parse(b []byte) (DefType, float64)
}

func round(f float64) float64 {
	shift := math.Pow(10, float64(2))
	return math.Floor((f*shift)+.5) / shift
}
