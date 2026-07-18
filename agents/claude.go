package agents

import "path/filepath"

type claudeTarget struct{}

func (claudeTarget) ID() string          { return "claude" }
func (claudeTarget) DisplayName() string { return "Claude Code" }

func (claudeTarget) SupportsLocation(Location) bool { return true }

func (t claudeTarget) Detect(loc Location, cwd string) Detection {
	dir := t.configDir(loc, cwd)
	installed := exists(dir)
	if loc == Global {
		installed = installed || exists(filepath.Join(homeDir(), ".claude.json"))
	} else {
		installed = installed || exists(filepath.Join(cwd, ".mcp.json"))
	}
	return Detection{Installed: installed, ConfigDir: dir}
}

func (t claudeTarget) WriteRule(loc Location, cwd, ruleName, body string) (Result, error) {
	return writeMarkdownRule(t.instructionsPath(loc, cwd), ruleName, body)
}

func (t claudeTarget) RemoveRule(loc Location, cwd, ruleName string) (Result, error) {
	return removeMarkdownRule(t.instructionsPath(loc, cwd), ruleName)
}

func (claudeTarget) configDir(loc Location, cwd string) string {
	if loc == Global {
		return filepath.Join(homeDir(), ".claude")
	}
	return filepath.Join(cwd, ".claude")
}

func (t claudeTarget) instructionsPath(loc Location, cwd string) string {
	return filepath.Join(t.configDir(loc, cwd), "CLAUDE.md")
}
