package rules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const caveGitignore = ".cursor/skills/cave*"

// Injectable for tests.
var runCavemanAdd = func(cwd string) error {
	cmd := exec.Command("npx", "-y", "skills", "add", "JuliusBrussee/caveman",
		"--skill", "*", "--agent", "cursor", "-y")
	cmd.Dir = cwd
	return runQuiet(cmd)
}

func caveman() Rule {
	var (
		hadLock   bool
		before    map[string]bool
		installed []string
	)
	return Rule{
		Name: "caveman",
		Steps: []Step{
			{
				Key:   "preinstall",
				Label: "Checked skills-lock.json",
				Run: func(opts Options) error {
					_, err := os.Stat(filepath.Join(opts.Cwd, "skills-lock.json"))
					hadLock = err == nil
					before = listSkillDirs(filepath.Join(opts.Cwd, ".agents", "skills"))
					return nil
				},
			},
			{
				Key:   "install",
				Label: "Run: npx skills add JuliusBrussee/caveman",
				Run: func(opts Options) error {
					if err := runCavemanAdd(opts.Cwd); err != nil {
						return err
					}
					after := listSkillDirs(filepath.Join(opts.Cwd, ".agents", "skills"))
					installed = nil
					for name := range after {
						if !before[name] {
							installed = append(installed, name)
						}
					}
					if len(installed) == 0 {
						return fmt.Errorf("no new skills installed under .agents/skills")
					}
					return nil
				},
			},
			{
				Key:   "postinstall",
				Label: "Moved new skills → .cursor/skills",
				Run: func(opts Options) error {
					if err := moveNamedSkills(opts.Cwd, installed); err != nil {
						return err
					}
					if err := pruneEmptyAgents(opts.Cwd); err != nil {
						return err
					}
					if !hadLock {
						_ = os.Remove(filepath.Join(opts.Cwd, "skills-lock.json"))
					}
					return ensureGitignoreCave(opts.Cwd)
				},
			},
			{
				Key:   "rules",
				Label: "Wrote: .cursor/rules/caveman.mdc",
				Run: func(opts Options) error {
					return writeMDC(opts.Cwd, "caveman")
				},
			},
		},
	}
}

func listSkillDirs(dir string) map[string]bool {
	out := map[string]bool{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() {
			out[e.Name()] = true
		}
	}
	return out
}

func moveNamedSkills(cwd string, names []string) error {
	src := filepath.Join(cwd, ".agents", "skills")
	dst := filepath.Join(cwd, ".cursor", "skills")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, name := range names {
		from := filepath.Join(src, name)
		to := filepath.Join(dst, name)
		if err := os.RemoveAll(to); err != nil {
			return err
		}
		if err := os.Rename(from, to); err != nil {
			return fmt.Errorf("move %s: %w", name, err)
		}
	}
	return nil
}

func pruneEmptyAgents(cwd string) error {
	skills := filepath.Join(cwd, ".agents", "skills")
	agents := filepath.Join(cwd, ".agents")
	if emptyDir(skills) {
		if err := os.Remove(skills); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if emptyDir(agents) {
		if err := os.Remove(agents); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func emptyDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) == 0
}

// ensureGitignoreCave appends .cursor/skills/cave* when .gitignore exists and lacks it.
func ensureGitignoreCave(cwd string) error {
	p := filepath.Join(cwd, ".gitignore")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == caveGitignore {
			return nil
		}
	}
	out := string(b)
	if len(out) > 0 && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	out += caveGitignore + "\n"
	return os.WriteFile(p, []byte(out), 0o644)
}
