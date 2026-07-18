package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mewisme/agentrule/agents"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.width < 1 {
			m.width = 1
		}
		if m.height < 1 {
			m.height = 1
		}
		m.applyListSize()
		return m, nil

	case workDoneMsg:
		return m.onWorkDone(msg)

	case tea.KeyPressMsg:
		key := msg.String()

		switch key {
		case "ctrl+c", "q", "esc":
			if m.step == stepInstall {
				return m.cancelInstall()
			}
			if m.step == stepDone || m.step == stepFailed {
				m.quitting = true
				return m, tea.Quit
			}
			m.quitting = true
			return m, tea.Quit
		}

		if m.step == stepDone || m.step == stepFailed {
			if key == "enter" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

		if m.step == stepInstall {
			return m, nil
		}

		if key == "left" || key == "backspace" {
			return m.onBack()
		}

		switch m.step {
		case stepRules, stepAgents:
			switch key {
			case "space":
				return m, m.toggleCurrent()
			case "a":
				return m, m.toggleAll()
			case "enter":
				return m.onEnter()
			}
		case stepLocation:
			if key == "enter" {
				return m.onEnter()
			}
		}

		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	if m.step == stepInstall {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) onBack() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepAgents:
		m.prevAgentIDs = m.selectedIDSet()
		idx := 0
		for i, loc := range m.locations {
			if loc == m.chosenLocation {
				idx = i
				break
			}
		}
		m.step = stepLocation
		m.list = newSelectList(locationItems(m.locations), "Step 2 of 3 — Location", false)
		m.list.Select(clampIndex(idx, len(m.locations)))
		m.applyListSize()
		return m, nil
	case stepLocation:
		m.step = stepRules
		m.list = newSelectList(ruleItems(m.ruleNames, m.savedRules), "Step 1 of 3 — Rules", true)
		m.applyListSize()
		return m, nil
	default:
		return m, nil
	}
}

func (m model) onEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepRules:
		if !hasChecked(m.list.Items()) {
			return m, nil
		}
		m.savedRules = m.selectedIDSet()
		m.step = stepLocation
		m.list = newSelectList(locationItems(m.locations), "Step 2 of 3 — Location", false)
		idx := 0
		for i, loc := range m.locations {
			if loc == m.chosenLocation {
				idx = i
				break
			}
		}
		m.list.Select(clampIndex(idx, len(m.locations)))
		m.applyListSize()
		return m, nil

	case stepLocation:
		item, ok := m.list.SelectedItem().(selectItem)
		if !ok {
			return m, nil
		}
		newLoc := agents.Location(item.id)
		changed := newLoc != m.chosenLocation
		m.chosenLocation = newLoc

		infos := agentsForLocation(m.chosenLocation)
		var sel map[string]bool
		if changed {
			sel = mergeAgentSelection(infos, m.prevAgentIDs, m.chosenLocation, m.opts.Cwd)
		} else if len(m.prevAgentIDs) > 0 {
			sel = mergeAgentSelection(infos, m.prevAgentIDs, m.chosenLocation, m.opts.Cwd)
		} else {
			sel = precheckAgents(infos, m.chosenLocation, m.opts.Cwd)
		}

		m.step = stepAgents
		m.list = newSelectList(agentListItems(infos, sel), "Step 3 of 3 — Agents", true)
		m.applyListSize()
		return m, nil

	case stepAgents:
		if !hasChecked(m.list.Items()) {
			return m, nil
		}
		m.prevAgentIDs = m.selectedIDSet()
		names := make([]string, 0, len(m.savedRules))
		for _, n := range m.ruleNames {
			if m.savedRules[n] {
				names = append(names, n)
			}
		}
		return m.startInstall(names)
	}
	return m, nil
}
