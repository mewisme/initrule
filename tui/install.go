package tui

import (
	"context"
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/mewisme/agentrule/rules"
)

type workDoneMsg struct {
	rule   string
	key    string
	label  string
	result rules.WorkResult
	err    error
}

func runWorkCmd(ctx context.Context, item rules.WorkItem, opts rules.Options) tea.Cmd {
	return func() tea.Msg {
		res, err := item.Run(ctx, opts)
		return workDoneMsg{
			rule:   item.Rule,
			key:    item.Key,
			label:  item.Label,
			result: res,
			err:    err,
		}
	}
}

func (m model) cancelInstall() (tea.Model, tea.Cmd) {
	m.canceled = true
	if m.cancel != nil {
		m.cancel()
	}
	m.quitting = true
	return m, tea.Quit
}

func (m model) startInstall(names []string) (tea.Model, tea.Cmd) {
	m.opts.Location = string(m.chosenLocation)
	m.opts.Target = strings.Join(m.selectedIDs(), ",")

	items, err := rules.WorkItems(names, m.opts)
	if err != nil {
		m.step = stepFailed
		m.installErr = err
		m.failedLbl = "plan"
		return m, tea.Quit
	}

	m.work = items
	m.workIndex = 0
	m.completed = 0
	m.total = len(items)
	m.lines = nil
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.step = stepInstall
	m.canceled = false

	if m.total == 0 {
		m.step = stepDone
		return m, tea.Quit
	}
	return m, tea.Batch(m.spinner.Tick, runWorkCmd(m.ctx, m.work[0], m.opts))
}

func (m model) onWorkDone(msg workDoneMsg) (tea.Model, tea.Cmd) {
	if m.canceled || m.step != stepInstall {
		return m, nil
	}
	if m.ctx != nil && m.ctx.Err() != nil {
		m.canceled = true
		return m, nil
	}

	if msg.err != nil {
		if errors.Is(msg.err, context.Canceled) {
			m.canceled = true
			m.quitting = true
			return m, tea.Quit
		}
		m.installErr = msg.err
		m.failedLbl = msg.label
		m.step = stepFailed
		return m, tea.Quit
	}

	label := msg.result.Label
	if label == "" {
		label = msg.label
	}
	m.lines = append(m.lines, completedLine{rule: msg.rule, label: label, detail: msg.result.Detail})
	m.completed++

	if m.completed >= m.total {
		m.step = stepDone
		return m, tea.Quit
	}

	m.workIndex++
	if m.workIndex >= len(m.work) {
		m.step = stepDone
		return m, tea.Quit
	}
	return m, tea.Batch(m.spinner.Tick, runWorkCmd(m.ctx, m.work[m.workIndex], m.opts))
}
