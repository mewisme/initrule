package agents

import "path/filepath"

type geminiTarget struct{}

func (geminiTarget) ID() string          { return "gemini" }
func (geminiTarget) DisplayName() string { return "Gemini CLI" }

func (geminiTarget) SupportsLocation(Location) bool { return true }

func (geminiTarget) Detect(loc Location, cwd string) Detection {
	dir := filepath.Join(homeDir(), ".gemini")
	if loc == Local {
		dir = filepath.Join(cwd, ".gemini")
	}
	installed := exists(dir)
	if loc == Global {
		installed = installed || exists(filepath.Join(homeDir(), ".gemini"))
	}
	return Detection{Installed: installed, ConfigDir: dir}
}

func (geminiTarget) WriteRule(loc Location, cwd, ruleName, body string) (Result, error) {
	return writeMarkdownRule(geminiInstructions(loc, cwd), ruleName, body)
}

func (geminiTarget) RemoveRule(loc Location, cwd, ruleName string) (Result, error) {
	return removeMarkdownRule(geminiInstructions(loc, cwd), ruleName)
}

func geminiInstructions(loc Location, cwd string) string {
	if loc == Global {
		return filepath.Join(homeDir(), ".gemini", "GEMINI.md")
	}
	return filepath.Join(cwd, "GEMINI.md")
}
