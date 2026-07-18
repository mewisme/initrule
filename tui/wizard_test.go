package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mewisme/agentrule/agents"
	"github.com/mewisme/agentrule/rules"
)

func testOpts(t *testing.T) rules.Options {
	t.Helper()
	return rules.Options{Cwd: t.TempDir(), Target: "auto", Location: "local"}
}

func TestRulesCannotAdvanceEmpty(t *testing.T) {
	m := newModel(testOpts(t))
	next, _ := m.onEnter()
	nm := next.(model)
	if nm.step != stepRules {
		t.Fatalf("step=%d", nm.step)
	}
}

func TestWizardAdvancesAndBack(t *testing.T) {
	m := newModel(testOpts(t))
	// select first rule
	cmd := m.toggleCurrent()
	if cmd != nil {
		m.list, _ = m.list.Update(nil)
	}
	// force check via SetItem
	it := m.list.Items()[0].(selectItem)
	it.checked = true
	_ = m.list.SetItem(0, it)

	next, _ := m.onEnter()
	m = next.(model)
	if m.step != stepLocation {
		t.Fatalf("step=%d want location", m.step)
	}

	next, _ = m.onEnter()
	m = next.(model)
	if m.step != stepAgents {
		t.Fatalf("step=%d want agents", m.step)
	}
	if !hasChecked(m.list.Items()) {
		t.Fatal("expected prechecked agents")
	}

	next, _ = m.onBack()
	m = next.(model)
	if m.step != stepLocation {
		t.Fatalf("back to location, got %d", m.step)
	}
	next, _ = m.onBack()
	m = next.(model)
	if m.step != stepRules {
		t.Fatalf("back to rules, got %d", m.step)
	}
	if !hasChecked(m.list.Items()) {
		t.Fatal("rule selection should be preserved")
	}
}

func TestLocationChangeRebuildsAgents(t *testing.T) {
	m := newModel(testOpts(t))
	it := m.list.Items()[0].(selectItem)
	it.checked = true
	_ = m.list.SetItem(0, it)
	next, _ := m.onEnter()
	m = next.(model) // location

	// pick global
	for i, loc := range m.locations {
		if loc == agents.Global {
			m.list.Select(i)
			break
		}
	}
	next, _ = m.onEnter()
	m = next.(model)
	if m.step != stepAgents {
		t.Fatalf("step=%d", m.step)
	}
	found := false
	for _, item := range m.list.Items() {
		si := item.(selectItem)
		if si.id == "codex" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("global agents should include codex")
	}

	// Mark codex+cursor checked, then back → local → agents; codex dropped, cursor kept.
	for i, item := range m.list.Items() {
		si := item.(selectItem)
		si.checked = si.id == "codex" || si.id == "cursor"
		_ = m.list.SetItem(i, si)
	}
	next, _ = m.onBack()
	m = next.(model)
	for i, loc := range m.locations {
		if loc == agents.Local {
			m.list.Select(i)
			break
		}
	}
	next, _ = m.onEnter()
	m = next.(model)
	for _, item := range m.list.Items() {
		if item.(selectItem).id == "codex" {
			t.Fatal("codex should not appear for local")
		}
	}
	ids := m.selectedIDSet()
	if ids["codex"] {
		t.Fatal("codex selection should be dropped")
	}
	if !ids["cursor"] {
		t.Fatal("cursor should remain")
	}
}

func TestToggleAll(t *testing.T) {
	m := newModel(testOpts(t))
	_ = m.toggleAll()
	if !hasChecked(m.list.Items()) {
		t.Fatal("toggle all should check")
	}
	_ = m.toggleAll()
	if hasChecked(m.list.Items()) {
		t.Fatal("toggle all again should clear")
	}
}

func TestAgentsCannotAdvanceEmpty(t *testing.T) {
	m := newModel(testOpts(t))
	it := m.list.Items()[0].(selectItem)
	it.checked = true
	_ = m.list.SetItem(0, it)
	next, _ := m.onEnter()
	m = next.(model)
	next, _ = m.onEnter()
	m = next.(model)
	// clear all agents
	items := m.list.Items()
	for i, item := range items {
		si := item.(selectItem)
		si.checked = false
		_ = m.list.SetItem(i, si)
	}
	next, _ = m.onEnter()
	m = next.(model)
	if m.step != stepAgents {
		t.Fatalf("should stay on agents, got %d", m.step)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := newModel(testOpts(t))
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = next.(model)
	if m.width != 100 || m.height != 40 {
		t.Fatalf("size=%dx%d", m.width, m.height)
	}
	if m.list.Width() < 1 || m.list.Height() < 1 {
		t.Fatal("list size invalid")
	}
}

func TestEscQuitsNotBack(t *testing.T) {
	m := newModel(testOpts(t))
	it := m.list.Items()[0].(selectItem)
	it.checked = true
	_ = m.list.SetItem(0, it)
	next, _ := m.onEnter()
	m = next.(model)
	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = next.(model)
	if !m.quitting {
		t.Fatal("esc should quit")
	}
	if cmd == nil {
		t.Fatal("expected Quit")
	}
	if m.step != stepLocation {
		t.Fatal("esc must not navigate back")
	}
}

func TestViewsIncludeBanner(t *testing.T) {
	m := newModel(testOpts(t))
	for _, step := range []int{stepRules, stepLocation, stepAgents, stepInstall, stepDone, stepFailed} {
		m.step = step
		v := m.View().Content
		if !strings.Contains(v, "█") {
			t.Fatalf("step %d: no banner glyphs in view", step)
		}
	}
}
