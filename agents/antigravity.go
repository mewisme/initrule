package agents

import "path/filepath"

type antigravityTarget struct{}

func (antigravityTarget) ID() string          { return "antigravity" }
func (antigravityTarget) DisplayName() string { return "Antigravity IDE" }

// Antigravity shares ~/.gemini/GEMINI.md with Gemini CLI — gemini target writes it.
func (antigravityTarget) SupportsLocation(loc Location) bool { return loc == Global }

func (antigravityTarget) Detect(loc Location, _ string) Detection {
	if loc != Global {
		return Detection{}
	}
	unified := filepath.Join(homeDir(), ".gemini", "config")
	legacy := filepath.Join(homeDir(), ".gemini", "antigravity")
	installed := exists(unified) || exists(legacy) || exists(filepath.Join(homeDir(), ".gemini"))
	return Detection{Installed: installed, ConfigDir: unified}
}

func (antigravityTarget) WriteRule(Location, string, string, string) (Result, error) {
	return Result{Action: "skipped"}, nil
}

func (antigravityTarget) RemoveRule(Location, string, string) (Result, error) {
	return Result{Action: "skipped"}, nil
}
