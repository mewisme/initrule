package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mewisme/initrule/rule"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type model struct {
	names     []string
	cursor    int
	selected  map[int]bool
	quitting  bool
	confirmed bool
}

func newModel(names []string) model {
	return model{
		names:    names,
		selected: map[int]bool{},
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.names)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "a":
			allOn := true
			for i := range m.names {
				if !m.selected[i] {
					allOn = false
					break
				}
			}
			for i := range m.names {
				m.selected[i] = !allOn
			}
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting || m.confirmed {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select rules to install"))
	b.WriteString("\n\n")
	for i, name := range m.names {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render(">")
		}
		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}
		fmt.Fprintf(&b, "%s %s %s\n", cursor, check, name)
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("space toggle · a all · enter confirm · q quit"))
	b.WriteString("\n")
	return b.String()
}

func (m model) chosen() []string {
	var out []string
	for i, name := range m.names {
		if m.selected[i] {
			out = append(out, name)
		}
	}
	return out
}

// Run shows multi-select; on confirm runs selected rules. Returns nil on quit with no selection.
func Run(opts rule.Options) error {
	p := tea.NewProgram(newModel(rule.Names()))
	final, err := p.Run()
	if err != nil {
		return err
	}
	m := final.(model)
	if m.quitting || !m.confirmed {
		return nil
	}
	names := m.chosen()
	if len(names) == 0 {
		fmt.Println("nothing selected")
		return nil
	}
	return rule.RunNames(names, opts)
}
