package agents

import "path/filepath"

type codexTarget struct{}

func (codexTarget) ID() string          { return "codex" }
func (codexTarget) DisplayName() string { return "Codex CLI" }

func (codexTarget) SupportsLocation(loc Location) bool { return loc == Global }

func (codexTarget) Detect(loc Location, _ string) Detection {
	if loc != Global {
		return Detection{}
	}
	dir := filepath.Join(homeDir(), ".codex")
	return Detection{Installed: exists(dir), ConfigDir: dir}
}

func (codexTarget) WriteRule(loc Location, _, ruleName, body string) (Result, error) {
	if loc != Global {
		return Result{Action: "skipped"}, nil
	}
	return writeMarkdownRule(filepath.Join(homeDir(), ".codex", "AGENTS.md"), ruleName, body)
}

func (codexTarget) RemoveRule(loc Location, _, ruleName string) (Result, error) {
	if loc != Global {
		return Result{Action: "skipped"}, nil
	}
	return removeMarkdownRule(filepath.Join(homeDir(), ".codex", "AGENTS.md"), ruleName)
}
