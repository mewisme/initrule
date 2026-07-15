package main

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

const banner = `
██╗███╗   ██╗██╗████████╗██████╗ ██╗   ██╗██╗     ███████╗
██║████╗  ██║██║╚══██╔══╝██╔══██╗██║   ██║██║     ██╔════╝
██║██╔██╗ ██║██║   ██║   ██████╔╝██║   ██║██║     █████╗
██║██║╚██╗██║██║   ██║   ██╔══██╗██║   ██║██║     ██╔══╝
██║██║ ╚████║██║   ██║   ██║  ██║╚██████╔╝███████╗███████╗
╚═╝╚═╝  ╚═══╝╚═╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝
`

// pastel stops from bokub/gradient-string (HSV long spin).
var (
	pastelStart, _ = colorful.Hex("#74ebd5")
	pastelEnd, _   = colorful.Hex("#74ecd5")
)

func printBanner() {
	println(pastelMultiline(strings.TrimPrefix(banner, "\n")))
}

// pastelMultiline mirrors gradient-string pastel.multiline: same horizontal
// HSV-long gradient on every line so colors stay vertically aligned.
func pastelMultiline(s string) string {
	lines := strings.Split(s, "\n")
	maxLen := 0
	for _, line := range lines {
		if n := utf8.RuneCountInString(line); n > maxLen {
			maxLen = n
		}
	}
	if maxLen < 2 {
		maxLen = 2
	}

	var b strings.Builder
	for li, line := range lines {
		i := 0
		for _, r := range line {
			t := float64(i) / float64(maxLen-1)
			c := pastelAt(t)
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex())).Render(string(r)))
			i++
		}
		if li < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func pastelAt(t float64) colorful.Color {
	h1, s1, v1 := pastelStart.Hsv()
	h2, s2, v2 := pastelEnd.Hsv()
	return colorful.Hsv(hueLerpLong(h1, h2, t), s1+(s2-s1)*t, v1+(v2-v1)*t)
}

func hueLerpLong(h1, h2, t float64) float64 {
	d := math.Mod(h2-h1, 360)
	if d < 0 {
		d += 360
	}
	// long arc (gradient-string hsvSpin: "long")
	if d < 180 {
		d -= 360
	}
	return math.Mod(h1+d*t+360, 360)
}
