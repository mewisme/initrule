package agents

import "path/filepath"

type opencodeTarget struct{}

func (opencodeTarget) ID() string          { return "opencode" }
func (opencodeTarget) DisplayName() string { return "opencode" }

func (opencodeTarget) SupportsLocation(Location) bool { return true }

func (opencodeTarget) Detect(loc Location, cwd string) Detection {
	dir := filepath.Join(xdgConfigHome(), "opencode")
	if loc == Local {
		dir = cwd
	}
	installed := loc == Global && exists(filepath.Join(xdgConfigHome(), "opencode"))
	if loc == Local {
		installed = exists(filepath.Join(cwd, "opencode.jsonc")) ||
			exists(filepath.Join(cwd, "opencode.json")) ||
			exists(filepath.Join(cwd, "AGENTS.md"))
	}
	return Detection{Installed: installed, ConfigDir: dir}
}

func (opencodeTarget) WriteRule(loc Location, cwd, ruleName, body string) (Result, error) {
	return writeMarkdownRule(opencodeInstructions(loc, cwd), ruleName, body)
}

func (opencodeTarget) RemoveRule(loc Location, cwd, ruleName string) (Result, error) {
	return removeMarkdownRule(opencodeInstructions(loc, cwd), ruleName)
}

func opencodeInstructions(loc Location, cwd string) string {
	if loc == Global {
		return filepath.Join(xdgConfigHome(), "opencode", "AGENTS.md")
	}
	return filepath.Join(cwd, "AGENTS.md")
}
