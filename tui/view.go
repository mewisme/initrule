package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m model) renderScreen(content string, footer string) string {
	var b strings.Builder
	b.WriteString(RenderBanner(m.width))
	b.WriteString("\n\n")
	b.WriteString(mainStyle.Render(content))
	if footer != "" {
		b.WriteString("\n")
		b.WriteString(mainStyle.Render(helpStyle.Render(footer)))
	}
	b.WriteString("\n")
	return b.String()
}

func (m model) viewContent() string {
	if m.quitting && m.step < stepInstall {
		return ""
	}
	switch m.step {
	case stepRules:
		return m.renderScreen(m.list.View(), "space toggle · a all · enter next · q quit")
	case stepLocation:
		var body strings.Builder
		body.WriteString(m.list.View())
		body.WriteByte('\n')
		body.WriteString(helpStyle.Render("codex, hermes, antigravity are global-only"))
		return m.renderScreen(body.String(), "↑/↓ · enter next · ← back · q quit")
	case stepAgents:
		return m.renderScreen(m.list.View(), "space toggle · a all · enter install · ← back · q quit")
	case stepInstall:
		return m.renderScreen(m.installView(), "q/esc cancel")
	case stepDone:
		return m.renderScreen(m.doneView(), "")
	case stepFailed:
		return m.renderScreen(m.failedView(), "")
	default:
		return m.renderScreen("", "q quit")
	}
}

func (m model) View() tea.View {
	v := tea.NewView(m.viewContent())
	v.AltScreen = true
	return v
}

func pipeLine() string { return dimStyle.Render("│") }

func (m model) timelineHeader() string {
	var b strings.Builder
	b.WriteString(dimStyle.Render("┌") + "  " + titleStyle.Render("Installing rules") + "\n")
	b.WriteString(pipeLine() + "\n")
	loc := m.opts.Location
	if loc == "" {
		loc = "local"
	}
	tgt := m.opts.Target
	if tgt == "" {
		tgt = "auto"
	}
	b.WriteString(okStyle.Render("◆") + "  Initialized in " + m.opts.Cwd + "\n")
	b.WriteString(fmt.Sprintf("%s  Agents: %s · location: %s\n", pipeLine(), tgt, loc))
	b.WriteString(pipeLine() + "\n")
	return b.String()
}

// writeTimelineLines writes completed steps grouped by rule, matching the CLI timeline.
func writeTimelineLines(b *strings.Builder, lines []completedLine) (lastRule string) {
	for _, line := range lines {
		if line.rule != lastRule {
			if lastRule != "" {
				b.WriteString(pipeLine() + "\n")
			}
			lastRule = line.rule
			b.WriteString(okStyle.Render("◆") + "  " + line.rule + "\n")
		}
		b.WriteString(pipeLine() + "  " + okStyle.Render("*") + " " + line.label + "\n")
		if line.detail != "" {
			b.WriteString(pipeLine() + "  " + dimStyle.Render(line.detail) + "\n")
		}
	}
	return lastRule
}

func (m model) installView() string {
	var b strings.Builder
	b.WriteString(m.timelineHeader())
	lastRule := writeTimelineLines(&b, m.lines)

	if m.workIndex < len(m.work) && !m.canceled {
		cur := m.work[m.workIndex]
		if cur.Rule != lastRule {
			if lastRule != "" {
				b.WriteString(pipeLine() + "\n")
			}
			b.WriteString(okStyle.Render("◆") + "  " + cur.Rule + "\n")
		}
		b.WriteString(fmt.Sprintf("%s  %s %s\n", pipeLine(), m.spinner.View(), cur.Label))
	}
	return b.String()
}

func (m model) doneView() string {
	var b strings.Builder
	b.WriteString(m.timelineHeader())
	lastRule := writeTimelineLines(&b, m.lines)
	if lastRule != "" {
		b.WriteString(pipeLine() + "\n")
	}
	b.WriteString(dimStyle.Render("└  ") + "Done\n")
	return b.String()
}

func (m model) failedView() string {
	var b strings.Builder
	b.WriteString(m.timelineHeader())
	lastRule := writeTimelineLines(&b, m.lines)

	failRule := ""
	if m.workIndex < len(m.work) {
		failRule = m.work[m.workIndex].Rule
	}
	if failRule != "" && failRule != lastRule {
		if lastRule != "" {
			b.WriteString(pipeLine() + "\n")
		}
		b.WriteString(okStyle.Render("◆") + "  " + failRule + "\n")
	}
	if m.failedLbl != "" {
		b.WriteString(pipeLine() + "  " + errStyle.Render("*") + " " + m.failedLbl + " failed\n")
	}
	if m.installErr != nil {
		b.WriteString(pipeLine() + "  " + errStyle.Render(m.installErr.Error()) + "\n")
	}
	b.WriteString(pipeLine() + "\n")
	b.WriteString(dimStyle.Render("└  ") + "Failed\n")
	return b.String()
}
