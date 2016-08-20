package util

import (
	"math"
)

// degrees to radians
func DtoR(d float32) float32 {
	return math.Pi * (d / 180.0)
}

// radians to degrees
func RtoD(r float32) float32 {
	return 180.0 * (r / math.Pi)
}
