package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Location string

const (
	Local  Location = "local"
	Global Location = "global"
)

type Detection struct {
	Installed bool
	ConfigDir string
}

// Result is one WriteRule / RemoveRule outcome.
type Result struct {
	Path   string
	Action string // created | updated | unchanged | removed | not-found | skipped
}

type Target interface {
	ID() string
	DisplayName() string
	SupportsLocation(Location) bool
	Detect(loc Location, cwd string) Detection
	WriteRule(loc Location, cwd, ruleName, body string) (Result, error)
	RemoveRule(loc Location, cwd, ruleName string) (Result, error)
}

var registry = []Target{
	cursorTarget{},
	claudeTarget{},
	codexTarget{},
	opencodeTarget{},
	hermesTarget{},
	geminiTarget{},
	antigravityTarget{},
	kiroTarget{},
}

func All() []Target { return append([]Target(nil), registry...) }

func IDs() []string {
	out := make([]string, len(registry))
	for i, t := range registry {
		out[i] = t.ID()
	}
	return out
}

func Get(id string) (Target, bool) {
	for _, t := range registry {
		if t.ID() == id {
			return t, true
		}
	}
	return nil, false
}

// ParseLocation accepts "local" / "global" (default local).
func ParseLocation(s string) (Location, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "local":
		return Local, nil
	case "global":
		return Global, nil
	default:
		return "", fmt.Errorf("unknown location %q (want local|global)", s)
	}
}

// Resolve picks targets from flag: ""|"auto", "all", or csv ids.
func Resolve(loc Location, cwd, flag string) ([]Target, error) {
	flag = strings.TrimSpace(flag)
	if flag == "" || flag == "auto" {
		var out []Target
		for _, t := range registry {
			if !t.SupportsLocation(loc) {
				continue
			}
			if t.Detect(loc, cwd).Installed {
				out = append(out, t)
			}
		}
		if len(out) == 0 {
			// Least-surprise fallback: Cursor (agentrule's Cursor surface).
			if t, ok := Get("cursor"); ok && t.SupportsLocation(loc) {
				return []Target{t}, nil
			}
		}
		return out, nil
	}
	if flag == "all" {
		var out []Target
		for _, t := range registry {
			if t.SupportsLocation(loc) {
				out = append(out, t)
			}
		}
		return out, nil
	}

	var out []Target
	var unknown []string
	for _, id := range strings.Split(flag, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		t, ok := Get(id)
		if !ok {
			unknown = append(unknown, id)
			continue
		}
		if !t.SupportsLocation(loc) {
			continue
		}
		out = append(out, t)
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("unknown --target id(s): %s (known: %s, plus auto|all)",
			strings.Join(unknown, ", "), strings.Join(IDs(), ","))
	}
	return out, nil
}

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func xdgConfigHome() string {
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return v
	}
	return filepath.Join(homeDir(), ".config")
}
