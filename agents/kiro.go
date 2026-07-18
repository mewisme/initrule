package agents

import (
	"os"
	"path/filepath"
)

type kiroTarget struct{}

func (kiroTarget) ID() string          { return "kiro" }
func (kiroTarget) DisplayName() string { return "Kiro" }

func (kiroTarget) SupportsLocation(Location) bool { return true }

func (kiroTarget) Detect(loc Location, cwd string) Detection {
	dir := filepath.Join(homeDir(), ".kiro")
	if loc == Local {
		dir = filepath.Join(cwd, ".kiro")
	}
	return Detection{Installed: exists(dir), ConfigDir: dir}
}

func (t kiroTarget) WriteRule(loc Location, cwd, ruleName, body string) (Result, error) {
	file := t.steeringPath(loc, cwd, ruleName)
	block := MarkedBody(ruleName, body)
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return Result{}, err
	}
	if b, err := os.ReadFile(file); err == nil && string(b) == block {
		return Result{Path: file, Action: "unchanged"}, nil
	}
	created := !exists(file)
	if err := atomicWrite(file, block); err != nil {
		return Result{}, err
	}
	if created {
		return Result{Path: file, Action: "created"}, nil
	}
	return Result{Path: file, Action: "updated"}, nil
}

func (t kiroTarget) RemoveRule(loc Location, cwd, ruleName string) (Result, error) {
	file := t.steeringPath(loc, cwd, ruleName)
	if !exists(file) {
		return Result{Path: file, Action: "not-found"}, nil
	}
	if err := os.Remove(file); err != nil {
		return Result{}, err
	}
	return Result{Path: file, Action: "removed"}, nil
}

func (kiroTarget) steeringPath(loc Location, cwd, ruleName string) string {
	root := filepath.Join(homeDir(), ".kiro")
	if loc == Local {
		root = filepath.Join(cwd, ".kiro")
	}
	return filepath.Join(root, "steering", ruleName+".md")
}
