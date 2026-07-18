package tui

import (
	"context"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mewisme/agentrule/agents"
	"github.com/mewisme/agentrule/rules"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	mainStyle     = lipgloss.NewStyle().MarginLeft(2)
)

const (
	stepRules = iota
	stepLocation
	stepAgents
	stepInstall
	stepDone
	stepFailed
)

type agentInfo struct {
	id   string
	name string
}

type completedLine struct {
	rule   string
	label  string
	detail string
}

type model struct {
	opts rules.Options

	step int
	list list.Model

	ruleNames  []string
	savedRules map[string]bool // rule selection across steps

	locations      []agents.Location
	chosenLocation agents.Location
	prevAgentIDs   map[string]bool // preserved across location changes

	work      []rules.WorkItem
	workIndex int
	completed int
	total     int
	lines     []completedLine
	failedLbl string

	spinner spinner.Model

	installErr error
	canceled   bool
	quitting   bool

	ctx    context.Context
	cancel context.CancelFunc

	width  int
	height int
}

func newModel(opts rules.Options) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{
		opts:           opts,
		step:           stepRules,
		ruleNames:      rules.Names(),
		savedRules:     map[string]bool{},
		locations:      []agents.Location{agents.Local, agents.Global},
		chosenLocation: agents.Local,
		prevAgentIDs:   map[string]bool{},
		spinner:        s,
		width:          80,
		height:         24,
	}
	m.list = newSelectList(ruleItems(m.ruleNames, nil), "Step 1 of 3 — Rules", true)
	m.applyListSize()
	return m
}

func (m model) Init() tea.Cmd { return nil }

func agentsForLocation(loc agents.Location) []agentInfo {
	var out []agentInfo
	for _, t := range agents.All() {
		if !t.SupportsLocation(loc) {
			continue
		}
		out = append(out, agentInfo{id: t.ID(), name: t.DisplayName()})
	}
	return out
}

// precheckAgents marks detected installs; falls back to cursor when none detected.
func precheckAgents(items []agentInfo, loc agents.Location, cwd string) map[string]bool {
	sel := map[string]bool{}
	any := false
	for _, item := range items {
		t, ok := agents.Get(item.id)
		if !ok {
			continue
		}
		if t.Detect(loc, cwd).Installed {
			sel[item.id] = true
			any = true
		}
	}
	if !any {
		sel["cursor"] = true
	}
	return sel
}

func mergeAgentSelection(items []agentInfo, prev map[string]bool, loc agents.Location, cwd string) map[string]bool {
	out := map[string]bool{}
	valid := map[string]bool{}
	for _, a := range items {
		valid[a.id] = true
	}
	kept := false
	for id, on := range prev {
		if on && valid[id] {
			out[id] = true
			kept = true
		}
	}
	if kept {
		return out
	}
	return precheckAgents(items, loc, cwd)
}
