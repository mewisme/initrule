package tui

import (
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/mewisme/agentrule/pastel"
)

const bannerArt = `
 █████╗  ██████╗ ███████╗███╗   ██╗████████╗    ██████╗ ██╗   ██╗██╗     ███████╗
██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝    ██╔══██╗██║   ██║██║     ██╔════╝
███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║       ██████╔╝██║   ██║██║     █████╗
██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║       ██╔══██╗██║   ██║██║     ██╔══╝
██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║       ██║  ██║╚██████╔╝███████╗███████╗
╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝       ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝
`

// BannerView returns the styled banner block used by CLI and TUI.
func BannerView() string {
	return pastelMultiline(strings.TrimPrefix(bannerArt, "\n"))
}

// RenderBanner returns the banner; width is reserved for future wrapping.
func RenderBanner(width int) string {
	_ = width
	return BannerView()
}

// BannerHeight is the number of lines in the banner art (for layout math).
func BannerHeight() int {
	return strings.Count(strings.TrimPrefix(bannerArt, "\n"), "\n") + 1
}

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
			c := pastel.Hex(t)
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(string(r)))
			i++
		}
		if li < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
