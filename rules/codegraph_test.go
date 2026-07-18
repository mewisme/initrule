package rules

import (
	"context"
	"errors"
	"testing"
)

func TestEnsureCodegraphBinary(t *testing.T) {
	oldLook, oldInst := lookPath, runCGInstaller
	t.Cleanup(func() {
		lookPath, runCGInstaller = oldLook, oldInst
	})

	t.Run("present skips install", func(t *testing.T) {
		calls := 0
		lookPath = func(string) (string, error) { return "/bin/codegraph", nil }
		runCGInstaller = func(context.Context) error {
			calls++
			return nil
		}
		if _, err := ensureCodegraphBinary(context.Background(), Options{}); err != nil {
			t.Fatal(err)
		}
		if calls != 0 {
			t.Fatalf("installer called %d times", calls)
		}
	})

	t.Run("missing runs install", func(t *testing.T) {
		calls := 0
		n := 0
		lookPath = func(string) (string, error) {
			n++
			if n == 1 {
				return "", errors.New("not found")
			}
			return "/bin/codegraph", nil
		}
		runCGInstaller = func(context.Context) error {
			calls++
			return nil
		}
		if _, err := ensureCodegraphBinary(context.Background(), Options{}); err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Fatalf("installer calls=%d", calls)
		}
	})

	t.Run("update forces install", func(t *testing.T) {
		calls := 0
		lookPath = func(string) (string, error) { return "/bin/codegraph", nil }
		runCGInstaller = func(context.Context) error {
			calls++
			return nil
		}
		if _, err := ensureCodegraphBinary(context.Background(), Options{Update: true}); err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Fatalf("installer calls=%d", calls)
		}
	})

	t.Run("still missing after install errors", func(t *testing.T) {
		lookPath = func(string) (string, error) { return "", errors.New("not found") }
		runCGInstaller = func(context.Context) error { return nil }
		if _, err := ensureCodegraphBinary(context.Background(), Options{}); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestInstallCodegraphAgentsUsesStub(t *testing.T) {
	old := runCGAgents
	t.Cleanup(func() { runCGAgents = old })
	called := false
	runCGAgents = func(ctx context.Context, opts Options) error {
		called = true
		if opts.Target != "cursor" {
			t.Fatalf("target=%s", opts.Target)
		}
		return nil
	}
	if _, err := installCodegraphAgents(context.Background(), Options{Cwd: t.TempDir(), Target: "cursor", Location: "local"}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("runCGAgents not called")
	}
}
