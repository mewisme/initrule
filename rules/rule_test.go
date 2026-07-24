package rules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteRuleAndRunNames(t *testing.T) {
	dir := t.TempDir()
	// Seed .cursor so auto-detect picks cursor.
	if err := os.MkdirAll(filepath.Join(dir, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}
	opts := Options{Cwd: dir, Target: "cursor", Location: "local"}

	oldLook, oldInst, oldAgents, oldInit := lookPath, runCGInstaller, runCGAgents, runCGInit
	t.Cleanup(func() {
		lookPath, runCGInstaller, runCGAgents, runCGInit = oldLook, oldInst, oldAgents, oldInit
	})
	lookPath = func(string) (string, error) { return "/fake/codegraph", nil }
	runCGInstaller = func(context.Context) error { t.Fatal("installer should not run"); return nil }
	runCGAgents = func(context.Context, Options) error { return nil }
	runCGInit = func(ctx context.Context, cwd string) error {
		d := filepath.Join(cwd, ".cursor", "rules")
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(d, "codegraph.mdc"), []byte("rewritten-by-codegraph-init"), 0o644)
	}

	if err := RunNames([]string{"ponytail", "codegraph"}, opts); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"ponytail", "codegraph"} {
		p := filepath.Join(dir, ".cursor", "rules", name+".mdc")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		want, err := content(name)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(want) {
			t.Fatalf("%s: content mismatch", name)
		}
	}
}

func TestWriteRuleClaude(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Cwd: dir, Target: "claude,cursor", Location: "local"}
	if err := writeRule(opts, "ponytail"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "<!-- PONYTAIL_START -->") {
		t.Fatal("missing ponytail block")
	}
	if strings.Contains(string(b), "alwaysApply:") {
		t.Fatal("frontmatter leaked into CLAUDE.md")
	}
}

func TestEmbedFromRewrite(t *testing.T) {
	b, err := content("codegraph")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "codegraph_explore") {
		t.Fatal("expected explore-only rewrite")
	}
	if strings.Contains(string(b), "codegraph_search") {
		t.Fatal("stale multi-tool docs still embedded")
	}

	adhd, err := content("i-have-adhd")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adhd), "Lead with the next action") {
		t.Fatal("expected i-have-adhd rewrite")
	}

	ps, err := content("powershell")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(ps), "Never `&&` or `||`") {
		t.Fatal("expected powershell rewrite")
	}
}

func TestUnknownRule(t *testing.T) {
	err := RunNames([]string{"nope"}, Options{Cwd: t.TempDir()})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestByName(t *testing.T) {
	if _, ok := ByName("ponytail"); !ok {
		t.Fatal("ponytail missing")
	}
	if _, ok := ByName("i-have-adhd"); !ok {
		t.Fatal("i-have-adhd missing")
	}

	old := goos
	goos = "windows"
	t.Cleanup(func() { goos = old })
	if _, ok := ByName("powershell"); !ok {
		t.Fatal("powershell missing on windows")
	}

	if _, ok := ByName("missing"); ok {
		t.Fatal("expected miss")
	}
}
