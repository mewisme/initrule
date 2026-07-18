package tui

import (
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/mewisme/agentrule/rules"
)

// ErrCanceled is returned from Run when the user cancels during installation.
var ErrCanceled = errors.New("installation canceled")

// Run shows the wizard and runs installation inside Bubble Tea.
// On success or failure, exits automatically and reprints the timeline to the
// normal console so it remains after the alt-screen closes.
// Returns nil on success or quit-before-install; ErrCanceled on cancel;
// installErr on failure.
func Run(opts rules.Options) error {
	p := tea.NewProgram(newModel(opts))
	final, err := p.Run()
	if err != nil {
		return err
	}
	m, ok := final.(model)
	if !ok {
		return nil
	}

	if m.canceled {
		return ErrCanceled
	}
	if m.step == stepFailed {
		fmt.Print(m.failedView())
		return m.installErr
	}
	if m.quitting && m.step < stepInstall {
		return nil
	}
	if m.step == stepDone {
		fmt.Print(m.doneView())
		return nil
	}
	if m.installErr != nil {
		return m.installErr
	}
	return nil
}
