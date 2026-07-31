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
	b.WriteString(dimStyle.Render("┌"))
	b.WriteString("  ")
	b.WriteString(titleStyle.Render("Installing rules"))
	b.WriteString("\n")
	b.WriteString(pipeLine())
	b.WriteString("\n")
	loc := m.opts.Location
	if loc == "" {
		loc = "local"
	}
	tgt := m.opts.Target
	if tgt == "" {
		tgt = "auto"
	}
	b.WriteString(okStyle.Render("◆"))
	b.WriteString("  Initialized in ")
	b.WriteString(m.opts.Cwd)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s  Agents: %s · location: %s\n", pipeLine(), tgt, loc))
	b.WriteString(pipeLine())
	b.WriteString("\n")
	return b.String()
}

// writeTimelineLines writes completed steps grouped by rule, matching the CLI timeline.
func writeTimelineLines(b *strings.Builder, lines []completedLine) (lastRule string) {
	for _, line := range lines {
		if line.rule != lastRule {
			if lastRule != "" {
				b.WriteString(pipeLine())
				b.WriteString("\n")
			}
			lastRule = line.rule
			b.WriteString(okStyle.Render("◆"))
			b.WriteString("  ")
			b.WriteString(line.rule)
			b.WriteString("\n")
		}
		b.WriteString(pipeLine())
		b.WriteString("  ")
		b.WriteString(okStyle.Render("*"))
		b.WriteString(" ")
		b.WriteString(line.label)
		b.WriteString("\n")
		if line.detail != "" {
			b.WriteString(pipeLine())
			b.WriteString("  ")
			b.WriteString(dimStyle.Render(line.detail))
			b.WriteString("\n")
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
				b.WriteString(pipeLine())
				b.WriteString("\n")
			}
			b.WriteString(okStyle.Render("◆"))
			b.WriteString("  ")
			b.WriteString(cur.Rule)
			b.WriteString("\n")
		}
		b.WriteString(pipeLine())
		b.WriteString("  ")
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(cur.Label)
		b.WriteString("\n")
	}
	return b.String()
}

func (m model) doneView() string {
	var b strings.Builder
	b.WriteString(m.timelineHeader())
	lastRule := writeTimelineLines(&b, m.lines)
	if lastRule != "" {
		b.WriteString(pipeLine())
		b.WriteString("\n")
	}
	b.WriteString(dimStyle.Render("└  "))
	b.WriteString("Done")
	b.WriteString("\n")
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
			b.WriteString(pipeLine())
			b.WriteString("\n")
		}
		b.WriteString(okStyle.Render("◆"))
		b.WriteString("  ")
		b.WriteString(failRule)
		b.WriteString("\n")
	}
	if m.failedLbl != "" {
		b.WriteString(pipeLine())
		b.WriteString("  ")
		b.WriteString(errStyle.Render("*"))
		b.WriteString(" ")
		b.WriteString(m.failedLbl)
		b.WriteString(" failed")
		b.WriteString("\n")
	}
	if m.installErr != nil {
		b.WriteString(pipeLine())
		b.WriteString("  ")
		b.WriteString(errStyle.Render(m.installErr.Error()))
		b.WriteString("\n")
	}
	b.WriteString(pipeLine())
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("└  "))
	b.WriteString("Failed")
	b.WriteString("\n")
	return b.String()
}
