package rules

import (
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
		runCGInstaller = func() error {
			calls++
			return nil
		}
		if err := ensureCodegraphBinary(Options{}); err != nil {
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
		runCGInstaller = func() error {
			calls++
			return nil
		}
		if err := ensureCodegraphBinary(Options{}); err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Fatalf("installer calls=%d", calls)
		}
	})

	t.Run("update forces install", func(t *testing.T) {
		calls := 0
		lookPath = func(string) (string, error) { return "/bin/codegraph", nil }
		runCGInstaller = func() error {
			calls++
			return nil
		}
		if err := ensureCodegraphBinary(Options{Update: true}); err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Fatalf("installer calls=%d", calls)
		}
	})

	t.Run("still missing after install errors", func(t *testing.T) {
		lookPath = func(string) (string, error) { return "", errors.New("not found") }
		runCGInstaller = func() error { return nil }
		if err := ensureCodegraphBinary(Options{}); err == nil {
			t.Fatal("expected error")
		}
	})
}
