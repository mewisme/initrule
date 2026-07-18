package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripFrontmatter(t *testing.T) {
	in := "---\ndescription: x\nalwaysApply: true\n---\n\n<!-- FOO_START -->\nbody\n<!-- FOO_END -->\n"
	got := StripFrontmatter(in)
	if strings.Contains(got, "description:") {
		t.Fatalf("frontmatter not stripped: %q", got)
	}
	if !strings.Contains(got, "<!-- FOO_START -->") {
		t.Fatalf("body missing: %q", got)
	}
}

func TestUpsertAndRemoveMarkedSection(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "CLAUDE.md")
	start, end := Markers("ponytail")
	block := MarkedBody("ponytail", "# hi\n")

	action, err := UpsertMarkedSection(file, block, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if action != "created" {
		t.Fatalf("action=%s", action)
	}

	action, err = UpsertMarkedSection(file, block, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if action != "unchanged" {
		t.Fatalf("action=%s", action)
	}

	if err := os.WriteFile(file, []byte("user stuff\n\n"+block), 0o644); err != nil {
		t.Fatal(err)
	}
	block2 := strings.Replace(block, "# hi", "# hello", 1)
	action, err = UpsertMarkedSection(file, block2, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if action != "updated" {
		t.Fatalf("action=%s", action)
	}
	b, _ := os.ReadFile(file)
	if !strings.Contains(string(b), "user stuff") || !strings.Contains(string(b), "# hello") {
		t.Fatalf("bad content: %s", b)
	}

	action, err = RemoveMarkedSection(file, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if action != "removed" {
		t.Fatalf("action=%s", action)
	}
	b, _ = os.ReadFile(file)
	if strings.Contains(string(b), start) {
		t.Fatal("markers still present")
	}
}

func TestCursorAndClaudeWrite(t *testing.T) {
	cwd := t.TempDir()
	body := "---\ndescription: t\nalwaysApply: true\n---\n<!-- PONYTAIL_START -->\n# P\n<!-- PONYTAIL_END -->\n"

	c := cursorTarget{}
	res, err := c.WriteRule(Local, cwd, "ponytail", body)
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "created" {
		t.Fatalf("cursor action=%s", res.Action)
	}
	got, err := os.ReadFile(filepath.Join(cwd, ".cursor", "rules", "ponytail.mdc"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body && string(got) != ensureTrailingNewline(body) {
		t.Fatalf("cursor content mismatch")
	}

	cl := claudeTarget{}
	res, err = cl.WriteRule(Local, cwd, "ponytail", body)
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "created" {
		t.Fatalf("claude action=%s", res.Action)
	}
	b, err := os.ReadFile(filepath.Join(cwd, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "description:") {
		t.Fatal("claude file should not include frontmatter")
	}
	if !strings.Contains(string(b), "<!-- PONYTAIL_START -->") {
		t.Fatal("missing markers")
	}
}

func TestCodexSkippedLocal(t *testing.T) {
	res, err := codexTarget{}.WriteRule(Local, t.TempDir(), "ponytail", "x")
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "skipped" {
		t.Fatalf("action=%s", res.Action)
	}
}

func TestResolveAllLocalSkipsGlobalOnly(t *testing.T) {
	targets, err := Resolve(Local, t.TempDir(), "all")
	if err != nil {
		t.Fatal(err)
	}
	for _, tgt := range targets {
		if tgt.ID() == "codex" || tgt.ID() == "hermes" || tgt.ID() == "antigravity" {
			t.Fatalf("unexpected %s in local all", tgt.ID())
		}
	}
}
