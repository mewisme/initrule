package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCavemanMoveAndLockCleanup(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Cwd: dir}

	old := runCavemanAdd
	t.Cleanup(func() { runCavemanAdd = old })
	runCavemanAdd = func(cwd string) error {
		skill := filepath.Join(cwd, ".agents", "skills", "caveman")
		if err := os.MkdirAll(skill, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("# caveman\n"), 0o644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(cwd, "skills-lock.json"), []byte(`{}`), 0o644)
	}

	if err := RunNames([]string{"caveman"}, opts); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".cursor", "skills", "caveman", "SKILL.md")); err != nil {
		t.Fatalf("skill not moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "skills-lock.json")); !os.IsNotExist(err) {
		t.Fatal("expected skills-lock.json removed when it did not exist before")
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); !os.IsNotExist(err) {
		t.Fatal("expected empty .agents removed")
	}
	want, err := content("caveman")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".cursor", "rules", "caveman.mdc"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("caveman.mdc mismatch")
	}
}

func TestCavemanMovesOnlyNewSkills(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Cwd: dir}

	pre := filepath.Join(dir, ".agents", "skills", "already-there")
	if err := os.MkdirAll(pre, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pre, "SKILL.md"), []byte("# old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	old := runCavemanAdd
	t.Cleanup(func() { runCavemanAdd = old })
	runCavemanAdd = func(cwd string) error {
		skill := filepath.Join(cwd, ".agents", "skills", "caveman")
		if err := os.MkdirAll(skill, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("# caveman\n"), 0o644)
	}

	if err := RunNames([]string{"caveman"}, opts); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".cursor", "skills", "caveman", "SKILL.md")); err != nil {
		t.Fatalf("new skill not moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "already-there", "SKILL.md")); err != nil {
		t.Fatalf("pre-existing skill should stay in .agents/skills: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".cursor", "skills", "already-there")); !os.IsNotExist(err) {
		t.Fatal("pre-existing skill should not be moved to .cursor/skills")
	}
}

func TestCavemanKeepsExistingLock(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Cwd: dir}
	lock := filepath.Join(dir, "skills-lock.json")
	if err := os.WriteFile(lock, []byte(`{"keep":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	old := runCavemanAdd
	t.Cleanup(func() { runCavemanAdd = old })
	runCavemanAdd = func(cwd string) error {
		skill := filepath.Join(cwd, ".agents", "skills", "caveman")
		if err := os.MkdirAll(skill, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("# caveman\n"), 0o644)
	}

	if err := RunNames([]string{"caveman"}, opts); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(lock)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"keep":true}` {
		t.Fatalf("lock changed: %s", b)
	}
}
