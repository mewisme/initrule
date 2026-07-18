package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mewisme/agentrule/agents"
)

func TestAgentsForLocationDropsGlobalOnly(t *testing.T) {
	items := agentsForLocation(agents.Local)
	for _, a := range items {
		switch a.id {
		case "codex", "hermes", "antigravity":
			t.Fatalf("local list should not include %s", a.id)
		}
	}
	global := agentsForLocation(agents.Global)
	found := map[string]bool{}
	for _, a := range global {
		found[a.id] = true
	}
	for _, id := range []string{"codex", "hermes", "antigravity", "claude"} {
		if !found[id] {
			t.Fatalf("global list missing %s", id)
		}
	}
}

func TestPrecheckAgentsDetectAndFallback(t *testing.T) {
	cwd := t.TempDir()
	items := agentsForLocation(agents.Local)

	sel := precheckAgents(items, agents.Local, cwd)
	if !sel["cursor"] {
		t.Fatal("expected cursor pre-checked as fallback")
	}

	if err := os.MkdirAll(filepath.Join(cwd, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	sel = precheckAgents(items, agents.Local, cwd)
	if !sel["claude"] {
		t.Fatal("expected claude pre-checked when .claude exists")
	}
}

func TestMergeAgentSelection(t *testing.T) {
	local := agentsForLocation(agents.Local)
	prev := map[string]bool{"claude": true, "codex": true} // codex invalid locally
	sel := mergeAgentSelection(local, prev, agents.Local, t.TempDir())
	if !sel["claude"] {
		t.Fatal("claude should be kept")
	}
	if sel["codex"] {
		t.Fatal("codex should be dropped for local")
	}
}

func TestListSizeNonNegative(t *testing.T) {
	m := newModel(testOpts(t))
	m.width, m.height = 10, 5
	w, h := m.listSize()
	if w < 1 || h < 1 {
		t.Fatalf("w=%d h=%d", w, h)
	}
	m.width, m.height = 0, 0
	w, h = m.listSize()
	if w < 1 || h < 1 {
		t.Fatalf("narrow w=%d h=%d", w, h)
	}
}
