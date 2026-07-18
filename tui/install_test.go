package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/mewisme/agentrule/rules"
)

func TestInstallSuccessToDone(t *testing.T) {
	m := newModel(testOpts(t))
	m.step = stepInstall
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.total = 1
	m.work = []rules.WorkItem{{
		Rule: "ponytail", Key: "install", Label: "Wrote rule",
		Run: func(context.Context, rules.Options) (rules.WorkResult, error) {
			return rules.WorkResult{Label: "Wrote ponytail → cursor"}, nil
		},
	}}
	m.workIndex = 0

	next, cmd := m.onWorkDone(workDoneMsg{
		label:  "Wrote rule",
		result: rules.WorkResult{Label: "Wrote ponytail → cursor"},
	})
	m = next.(model)
	if m.step != stepDone {
		t.Fatalf("step=%d want done", m.step)
	}
	if cmd == nil {
		t.Fatal("expected auto-quit on success")
	}
	if m.completed != 1 {
		t.Fatalf("completed=%d", m.completed)
	}
}

func TestInstallFailureToFailed(t *testing.T) {
	m := newModel(testOpts(t))
	m.step = stepInstall
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.total = 2
	m.work = []rules.WorkItem{
		{Rule: "r", Key: "1", Label: "one"},
		{Rule: "r", Key: "2", Label: "two"},
	}
	m.workIndex = 0

	next, cmd := m.onWorkDone(workDoneMsg{label: "one", err: errors.New("boom")})
	m = next.(model)
	if m.step != stepFailed {
		t.Fatalf("step=%d want failed", m.step)
	}
	if cmd == nil {
		t.Fatal("expected auto-quit on failure")
	}
	if m.installErr == nil {
		t.Fatal("expected installErr")
	}
}

func TestCancelPreventsAdvance(t *testing.T) {
	m := newModel(testOpts(t))
	m.step = stepInstall
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.canceled = true
	m.total = 2
	m.workIndex = 0
	m.work = []rules.WorkItem{{}, {}}

	next, cmd := m.onWorkDone(workDoneMsg{
		result: rules.WorkResult{Label: "ok"},
	})
	m = next.(model)
	if m.completed != 0 {
		t.Fatal("late success must not advance")
	}
	if m.step != stepInstall {
		t.Fatalf("step=%d", m.step)
	}
	if cmd != nil {
		t.Fatal("must not schedule next work")
	}
}

func TestCancelInstallSetsFlag(t *testing.T) {
	m := newModel(testOpts(t))
	m.step = stepInstall
	m.ctx, m.cancel = context.WithCancel(context.Background())
	next, cmd := m.cancelInstall()
	m = next.(model)
	if !m.canceled || !m.quitting {
		t.Fatal("expected canceled+quitting")
	}
	if cmd == nil {
		t.Fatal("expected Quit")
	}
	if m.ctx.Err() == nil {
		t.Fatal("context should be canceled")
	}
}

func TestSequentialScheduling(t *testing.T) {
	calls := 0
	m := newModel(testOpts(t))
	m.step = stepInstall
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.total = 2
	m.work = []rules.WorkItem{
		{Rule: "r", Key: "1", Label: "one", Run: func(context.Context, rules.Options) (rules.WorkResult, error) {
			calls++
			return rules.WorkResult{Label: "one"}, nil
		}},
		{Rule: "r", Key: "2", Label: "two", Run: func(context.Context, rules.Options) (rules.WorkResult, error) {
			calls++
			return rules.WorkResult{Label: "two"}, nil
		}},
	}
	m.workIndex = 0

	next, cmd := m.onWorkDone(workDoneMsg{label: "one", result: rules.WorkResult{Label: "one"}})
	m = next.(model)
	if m.step != stepInstall {
		t.Fatalf("should still install, got %d", m.step)
	}
	if m.workIndex != 1 {
		t.Fatalf("workIndex=%d", m.workIndex)
	}
	if cmd == nil {
		t.Fatal("expected next work cmd")
	}
	// run the cmd once
	msg := cmd()
	if _, ok := msg.(workDoneMsg); !ok {
		// Batch returns a batch msg — just ensure we got a cmd
		_ = msg
	}
}
