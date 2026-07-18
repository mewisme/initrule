package rules

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mewisme/agentrule/agents"
)

// Injectable for tests.
var runCGUninstall = func(opts Options) error {
	loc, err := agents.ParseLocation(opts.Location)
	if err != nil {
		return err
	}
	targets, err := agents.Resolve(loc, opts.Cwd, opts.Target)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(targets))
	for _, t := range targets {
		ids = append(ids, t.ID())
	}
	if len(ids) == 0 {
		ids = agents.IDs()
	}
	cmd := exec.CommandContext(context.Background(), "codegraph", "uninstall",
		"--target="+strings.Join(ids, ","),
		"--location="+string(loc),
		"--yes",
		"--keep-cli",
	)
	cmd.Dir = opts.Cwd
	return runQuiet(cmd)
}

func removeRule(opts Options, name string) error {
	loc, err := agents.ParseLocation(opts.Location)
	if err != nil {
		return err
	}
	targets, err := agents.Resolve(loc, opts.Cwd, opts.Target)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("no agent targets selected")
	}
	var touched int
	for _, t := range targets {
		res, err := t.RemoveRule(loc, opts.Cwd, name)
		if err != nil {
			return fmt.Errorf("%s: %w", t.ID(), err)
		}
		switch res.Action {
		case "skipped", "not-found", "kept":
			continue
		default:
			touched++
		}
	}
	// Not an error if nothing was present — uninstall is idempotent.
	_ = touched
	return nil
}

func uninstallCodegraphAgents(opts Options) error {
	if _, err := lookPath("codegraph"); err != nil {
		// No binary → nothing to reverse for MCP; rule files still removed above.
		return nil
	}
	return runCGUninstall(opts)
}

func runUninstall(names []string, opts Options) error {
	fmt.Printf("\n%s  %s\n", dimStyle.Render("┌"), titleStyle.Render("Uninstalling rules"))
	pipe()

	loc := opts.Location
	if loc == "" {
		loc = "local"
	}
	tgt := opts.Target
	if tgt == "" {
		tgt = "all"
	}
	fmt.Printf("%s  Working in %s\n", okStyle.Render("◆"), opts.Cwd)
	fmt.Printf("%s  Agents: %s · location: %s\n", dimStyle.Render("│"), tgt, loc)
	pipe()

	for _, name := range names {
		if _, ok := ByName(name); !ok {
			return fmt.Errorf("unknown rule %q", name)
		}
		fmt.Printf("%s  %s\n", okStyle.Render("◆"), name)

		if err := removeRule(opts, name); err != nil {
			fmt.Printf("%s  %s remove rule failed\n", dimStyle.Render("│"), errStyle.Render("*"))
			pipe()
			fmt.Println(dimStyle.Render("└  ") + "Failed")
			return fmt.Errorf("%s: remove: %w", name, err)
		}
		fmt.Printf("%s  %s Removed rule from agent targets\n", dimStyle.Render("│"), okStyle.Render("*"))

		if name == "codegraph" {
			if err := uninstallCodegraphAgents(opts); err != nil {
				fmt.Printf("%s  %s codegraph uninstall failed\n", dimStyle.Render("│"), errStyle.Render("*"))
				pipe()
				fmt.Println(dimStyle.Render("└  ") + "Failed")
				return fmt.Errorf("codegraph: uninstall: %w", err)
			}
			fmt.Printf("%s  %s Ran: codegraph uninstall --keep-cli\n", dimStyle.Render("│"), okStyle.Render("*"))
		}
		pipe()
	}

	fmt.Println(dimStyle.Render("└  ") + "Done")
	return nil
}

// UninstallNames removes named rules from selected agents (registry order).
func UninstallNames(names []string, opts Options) error {
	want := map[string]bool{}
	for _, n := range names {
		if _, ok := ByName(n); !ok {
			return fmt.Errorf("unknown rule %q", n)
		}
		want[n] = true
	}
	var ordered []string
	for _, r := range All() {
		if want[r.Name] {
			ordered = append(ordered, r.Name)
		}
	}
	if opts.Target == "" || opts.Target == "auto" {
		// Uninstall sweeps every agent at this location by default.
		opts.Target = "all"
	}
	return runUninstall(ordered, opts)
}

func UninstallAll(opts Options) error {
	return UninstallNames(Names(), opts)
}
