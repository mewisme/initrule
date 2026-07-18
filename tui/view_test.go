package tui

import (
	"strings"
	"testing"
)

func TestRenderScreenComposesBanner(t *testing.T) {
	m := newModel(testOpts(t))
	out := m.renderScreen("hello", "footer")
	if !strings.Contains(out, "hello") {
		t.Fatal("missing content")
	}
	if !strings.Contains(out, "footer") {
		t.Fatal("missing footer")
	}
	if !strings.Contains(out, "█") {
		t.Fatal("missing banner glyphs")
	}
}

func TestInstallViewNoStepFour(t *testing.T) {
	m := newModel(testOpts(t))
	m.step = stepInstall
	m.total = 2
	v := m.View().Content
	if strings.Contains(v, "Step 4") || strings.Contains(v, "4 of 3") {
		t.Fatal("install must not show Step 4 of 3")
	}
	if !strings.Contains(v, "Installing rules") {
		t.Fatal("expected Installing rules title")
	}
	if !strings.Contains(v, "┌") || !strings.Contains(v, "│") {
		t.Fatal("expected timeline lines")
	}
}
