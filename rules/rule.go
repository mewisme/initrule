package rules

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

type Options struct {
	Cwd    string
	Update bool // --update / -u
}

type Step struct {
	Key   string // preinstall | install | postinstall
	Label string
	Run   func(opts Options) error
}

type Rule struct {
	Name  string
	Steps []Step
}

var (
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	titleStyle = lipgloss.NewStyle().Bold(true)
)

func writeMDC(cwd, name string) error {
	b, err := content(name)
	if err != nil {
		return err
	}
	dir := filepath.Join(cwd, ".cursor", "rules")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".mdc"), b, 0o644)
}

// All is the registered rule list (append new rules here).
func All() []Rule {
	return []Rule{
		{
			Name: "codegraph",
			Steps: []Step{
				{Key: "preinstall", Label: "Ensure codegraph binary", Run: ensureCodegraphBinary},
				{Key: "install", Label: "Run: codegraph init -i", Run: initCodegraph},
				{Key: "postinstall", Label: "Wrote: .cursor/rules/codegraph.mdc", Run: func(opts Options) error {
					return writeMDC(opts.Cwd, "codegraph")
				}},
			},
		},
		{
			Name: "ponytail",
			Steps: []Step{
				{Key: "install", Label: "Wrote: .cursor/rules/ponytail.mdc", Run: func(opts Options) error {
					return writeMDC(opts.Cwd, "ponytail")
				}},
			},
		},
	}
}

func ByName(name string) (Rule, bool) {
	for _, r := range All() {
		if r.Name == name {
			return r, true
		}
	}
	return Rule{}, false
}

func Names() []string {
	all := All()
	out := make([]string, len(all))
	for i, r := range all {
		out[i] = r.Name
	}
	return out
}

func pipe() { fmt.Println(dimStyle.Render("│")) }

func runRules(rules []Rule, opts Options) error {
	fmt.Printf("\n%s  %s\n", dimStyle.Render("┌"), titleStyle.Render("Installing rules"))
	pipe()

	fmt.Printf("%s  Initialized in %s\n", okStyle.Render("◆"), opts.Cwd)
	pipe()

	for _, r := range rules {
		fmt.Printf("%s  %s\n", okStyle.Render("◆"), r.Name)
		for _, s := range r.Steps {
			if s.Run == nil {
				continue
			}
			if err := s.Run(opts); err != nil {
				fmt.Printf("%s  %s %s\n", dimStyle.Render("│"), errStyle.Render("*"), s.Label+" failed")
				pipe()
				fmt.Println(dimStyle.Render("└  ") + "Failed")
				return fmt.Errorf("%s: %s: %w", r.Name, s.Key, err)
			}
			fmt.Printf("%s  %s %s\n", dimStyle.Render("│"), okStyle.Render("*"), s.Label)
		}
		pipe()
	}

	fmt.Println(dimStyle.Render("└  ") + "Done")
	return nil
}

// Run executes one rule inside the shared install timeline.
func Run(r Rule, opts Options) error {
	return runRules([]Rule{r}, opts)
}

// RunNames resolves and runs rules in registry order (not arg order).
func RunNames(names []string, opts Options) error {
	want := map[string]bool{}
	for _, n := range names {
		if _, ok := ByName(n); !ok {
			return fmt.Errorf("unknown rule %q", n)
		}
		want[n] = true
	}
	var rules []Rule
	for _, r := range All() {
		if want[r.Name] {
			rules = append(rules, r)
		}
	}
	return runRules(rules, opts)
}

func RunAll(opts Options) error {
	return runRules(All(), opts)
}
