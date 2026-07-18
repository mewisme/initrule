package agents

import (
	"os"
	"path/filepath"
)

type hermesTarget struct{}

func (hermesTarget) ID() string          { return "hermes" }
func (hermesTarget) DisplayName() string { return "Hermes Agent" }

// Hermes has no project-local instruction file; MCP-only via codegraph install.
func (hermesTarget) SupportsLocation(loc Location) bool { return loc == Global }

func (hermesTarget) Detect(loc Location, _ string) Detection {
	if loc != Global {
		return Detection{}
	}
	dir := hermesHome()
	return Detection{Installed: exists(dir) || exists(filepath.Join(dir, "config.yaml")), ConfigDir: dir}
}

func (hermesTarget) WriteRule(Location, string, string, string) (Result, error) {
	// No markdown instructions surface — MCP wiring is codegraph install's job.
	return Result{Action: "skipped"}, nil
}

func (hermesTarget) RemoveRule(Location, string, string) (Result, error) {
	return Result{Action: "skipped"}, nil
}

func hermesHome() string {
	if v := os.Getenv("HERMES_HOME"); v != "" {
		return v
	}
	return filepath.Join(homeDir(), ".hermes")
}
