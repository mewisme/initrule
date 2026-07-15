package rule

import (
	"fmt"
	"os"
	"path/filepath"
)

type Options struct {
	Cwd    string
	Update bool // --update / -u
}

type Rule struct {
	Name        string
	PreInstall  func(opts Options) error // nil = skip
	Install     func(opts Options) error
	PostInstall func(opts Options) error // nil = skip
}

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
			Name:        "codegraph",
			PreInstall:  ensureCodegraphBinary,
			Install:     initCodegraph, // codegraph init -i
			PostInstall: func(opts Options) error { return writeMDC(opts.Cwd, "codegraph") },
		},
		{
			Name:    "ponytail",
			Install: func(opts Options) error { return writeMDC(opts.Cwd, "ponytail") },
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

// Run executes preinstall → install → postinstall for one rule.
func Run(r Rule, opts Options) error {
	steps := []struct {
		name string
		fn   func(Options) error
	}{
		{"preinstall", r.PreInstall},
		{"install", r.Install},
		{"postinstall", r.PostInstall},
	}
	for _, s := range steps {
		if s.fn == nil {
			continue
		}
		if err := s.fn(opts); err != nil {
			return fmt.Errorf("%s: %s: %w", r.Name, s.name, err)
		}
		fmt.Printf("%s: %s ok\n", r.Name, s.name)
	}
	return nil
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
	for _, r := range All() {
		if !want[r.Name] {
			continue
		}
		if err := Run(r, opts); err != nil {
			return err
		}
	}
	return nil
}

func RunAll(opts Options) error {
	for _, r := range All() {
		if err := Run(r, opts); err != nil {
			return err
		}
	}
	return nil
}
