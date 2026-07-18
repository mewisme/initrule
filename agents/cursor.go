package agents

import (
	"os"
	"path/filepath"
	"strings"
)

type cursorTarget struct{}

func (cursorTarget) ID() string          { return "cursor" }
func (cursorTarget) DisplayName() string { return "Cursor" }

func (cursorTarget) SupportsLocation(loc Location) bool {
	// Cursor rules are project-scoped only.
	return loc == Local
}

func (cursorTarget) Detect(loc Location, cwd string) Detection {
	if loc != Local {
		return Detection{}
	}
	dir := filepath.Join(cwd, ".cursor")
	return Detection{Installed: exists(dir) || exists(filepath.Join(homeDir(), ".cursor")), ConfigDir: dir}
}

func (cursorTarget) WriteRule(loc Location, cwd, ruleName, body string) (Result, error) {
	if loc != Local {
		return Result{Action: "skipped"}, nil
	}
	file := filepath.Join(cwd, ".cursor", "rules", ruleName+".mdc")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return Result{}, err
	}
	want := ensureTrailingNewline(body)
	if b, err := os.ReadFile(file); err == nil && string(b) == want {
		return Result{Path: file, Action: "unchanged"}, nil
	}
	created := !exists(file)
	if err := atomicWrite(file, want); err != nil {
		return Result{}, err
	}
	if created {
		return Result{Path: file, Action: "created"}, nil
	}
	return Result{Path: file, Action: "updated"}, nil
}

func (cursorTarget) RemoveRule(loc Location, cwd, ruleName string) (Result, error) {
	if loc != Local {
		return Result{Action: "skipped"}, nil
	}
	file := filepath.Join(cwd, ".cursor", "rules", ruleName+".mdc")
	if !exists(file) {
		return Result{Path: file, Action: "not-found"}, nil
	}
	if err := os.Remove(file); err != nil {
		return Result{}, err
	}
	return Result{Path: file, Action: "removed"}, nil
}

func ensureTrailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}
