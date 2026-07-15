package rule

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteMDCAndRunNames(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Cwd: dir}

	// Stub codegraph hooks so we don't touch the real binary.
	oldLook, oldInst, oldInit := lookPath, runCGInstaller, runCGInit
	t.Cleanup(func() {
		lookPath, runCGInstaller, runCGInit = oldLook, oldInst, oldInit
	})
	lookPath = func(string) (string, error) { return "/fake/codegraph", nil }
	runCGInstaller = func() error { t.Fatal("installer should not run"); return nil }
	// Simulate codegraph init -i writing its own mdc before postinstall restores embed.
	runCGInit = func(cwd string) error {
		dir := filepath.Join(cwd, ".cursor", "rules")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "codegraph.mdc"), []byte("rewritten-by-codegraph-init"), 0o644)
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
	if _, ok := ByName("missing"); ok {
		t.Fatal("expected miss")
	}
}
