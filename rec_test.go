package rec

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestScale(t *testing.T) {
	assertNear(t, scale(0, 5, 0), float32(-1))
	assertNear(t, scale(10, 15, 15), float32(1))
	assertNear(t, scale(-10, -9, -9.75), float32(-0.5))
	assertNear(t, scale(-5, 5, 1), float32(0.2))
}

// Inexact float value assertion. Needed because of float rounding error.
func assertNear(t *testing.T, expected, actual float32) {
	var delta float32 = 0.0001
	assert.T(t, actual > expected-delta, actual, "<<", expected)
	assert.T(t, actual < expected+delta, actual, ">>", expected)
}
