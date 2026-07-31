// Package pastel is bokub/gradient-string's core pastel gradient (HSV long spin).
// Import from any app: github.com/mewisme/agentrule/pastel
package pastel

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

// Core stops from bokub/gradient-string pastel.
const (
	Start = "#74ebd5"
	End   = "#acb6e5"
)

var (
	start, _ = colorful.Hex(Start)
	end, _   = colorful.Hex(End)
)

// At returns the pastel color at t ∈ [0,1] (HSV, long hue arc).
func At(t float64) colorful.Color {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	h1, s1, v1 := start.Hsv()
	h2, s2, v2 := end.Hsv()
	return colorful.Hsv(hueLerpLong(h1, h2, t), s1+(s2-s1)*t, v1+(v2-v1)*t)
}

// Hex is At(t).Hex().
func Hex(t float64) string { return At(t).Hex() }

func hueLerpLong(h1, h2, t float64) float64 {
	d := math.Mod(h2-h1, 360)
	if d < 0 {
		d += 360
	}
	if d < 180 {
		d -= 360
	}
	return math.Mod(h1+d*t+360, 360)
}
