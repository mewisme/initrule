package rules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUninstallRemovesCursorAndClaude(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Cwd: dir, Target: "cursor,claude", Location: "local"}

	oldLook, oldAgents, oldInit, oldUn := lookPath, runCGAgents, runCGInit, runCGUninstall
	t.Cleanup(func() {
		lookPath, runCGAgents, runCGInit, runCGUninstall = oldLook, oldAgents, oldInit, oldUn
	})
	lookPath = func(string) (string, error) { return "/fake/codegraph", nil }
	runCGAgents = func(context.Context, Options) error { return nil }
	runCGInit = func(context.Context, string) error { return nil }
	unCalls := 0
	runCGUninstall = func(Options) error {
		unCalls++
		return nil
	}

	if err := RunNames([]string{"ponytail", "codegraph"}, opts); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ponytail", "codegraph"} {
		if _, err := os.Stat(filepath.Join(dir, ".cursor", "rules", name+".mdc")); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "PONYTAIL_START") {
		t.Fatal("ponytail not in CLAUDE.md")
	}

	if err := UninstallNames([]string{"ponytail", "codegraph"}, opts); err != nil {
		t.Fatal(err)
	}
	if unCalls != 1 {
		t.Fatalf("codegraph uninstall calls=%d", unCalls)
	}
	for _, name := range []string{"ponytail", "codegraph"} {
		if _, err := os.Stat(filepath.Join(dir, ".cursor", "rules", name+".mdc")); !os.IsNotExist(err) {
			t.Fatalf("%s.mdc still present", name)
		}
	}
	b, err = os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if err == nil && (strings.Contains(string(b), "PONYTAIL_START") || strings.Contains(string(b), "CODEGRAPH_START")) {
		t.Fatalf("markers still in CLAUDE.md: %s", b)
	}
}

func TestUninstallIdempotent(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Cwd: dir, Target: "cursor", Location: "local"}
	oldUn, oldLook := runCGUninstall, lookPath
	t.Cleanup(func() {
		runCGUninstall, lookPath = oldUn, oldLook
	})
	runCGUninstall = func(Options) error { return nil }
	lookPath = func(string) (string, error) { return "", os.ErrNotExist }

	if err := UninstallNames([]string{"ponytail"}, opts); err != nil {
		t.Fatal(err)
	}
}

func TestUninstallAutoBecomesAll(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Cwd: dir, Target: "auto", Location: "local"}
	if err := os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0o755); err != nil {
		t.Fatal(err)
	}
	body, _ := content("ponytail")
	if err := os.WriteFile(filepath.Join(dir, ".cursor", "rules", "ponytail.mdc"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := UninstallNames([]string{"ponytail"}, opts); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".cursor", "rules", "ponytail.mdc")); !os.IsNotExist(err) {
		t.Fatal("expected removed")
	}
}

func TestUninstallUnknownRule(t *testing.T) {
	err := UninstallNames([]string{"nope"}, Options{Cwd: t.TempDir(), Target: "all", Location: "local"})
	if err == nil {
		t.Fatal("expected error")
	}
}
