package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mewisme/agentrule/agents"
	"github.com/mewisme/agentrule/rules"
	"github.com/mewisme/agentrule/tui"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage:
  agentrule                              interactive install wizard
  agentrule install|i <name>...          install named rules
  agentrule install|i -a|--all           install all rules
  agentrule uninstall|un <name>...       uninstall named rules
  agentrule uninstall|un -a|--all        uninstall all rules

flags (install|i and uninstall|un):
  -a, --all              all registered rules
  --target <ids>         auto|all|csv agent ids
  --location <loc>       local|global (default: local)

install-only:
  -u, --update           force codegraph binary reinstall

defaults:
  install:   --target=auto
  uninstall: --target=all  (sweeps all agents at location)

agents: %s
rules:  %v
`, strings.Join(agents.IDs(), ", "), rules.Names())
}

func main() {
	printBanner()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	opts := rules.Options{Cwd: cwd, Target: "auto", Location: "local"}

	if len(os.Args) == 1 {
		if err := tui.Run(opts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "install", "i":
		err = runInstall(os.Args[2:], opts)
	case "uninstall", "un":
		err = runUninstall(os.Args[2:], opts)
	default:
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runInstall(args []string, opts rules.Options) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var all, update bool
	var target, location string
	fs.BoolVar(&all, "a", false, "install all rules")
	fs.BoolVar(&all, "all", false, "install all rules")
	fs.BoolVar(&update, "u", false, "force codegraph binary reinstall")
	fs.BoolVar(&update, "update", false, "force codegraph binary reinstall")
	fs.StringVar(&target, "target", "auto", "agent targets: auto|all|csv")
	fs.StringVar(&location, "location", "local", "local|global")
	if err := fs.Parse(args); err != nil {
		return err
	}
	opts.Update = update
	opts.Target = target
	opts.Location = location
	if _, err := agents.ParseLocation(location); err != nil {
		return err
	}
	names := fs.Args()
	if all && len(names) > 0 {
		return fmt.Errorf("use either -a/--all or rule names, not both")
	}
	if !all && len(names) == 0 {
		usage()
		return fmt.Errorf("no rules specified")
	}
	if all {
		return rules.RunAll(opts)
	}
	return rules.RunNames(names, opts)
}

func runUninstall(args []string, opts rules.Options) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var all bool
	var target, location string
	fs.BoolVar(&all, "a", false, "uninstall all rules")
	fs.BoolVar(&all, "all", false, "uninstall all rules")
	fs.StringVar(&target, "target", "all", "agent targets: auto|all|csv")
	fs.StringVar(&location, "location", "local", "local|global")
	if err := fs.Parse(args); err != nil {
		return err
	}
	opts.Target = target
	opts.Location = location
	if _, err := agents.ParseLocation(location); err != nil {
		return err
	}
	names := fs.Args()
	if all && len(names) > 0 {
		return fmt.Errorf("use either -a/--all or rule names, not both")
	}
	if !all && len(names) == 0 {
		usage()
		return fmt.Errorf("no rules specified")
	}
	if all {
		return rules.UninstallAll(opts)
	}
	return rules.UninstallNames(names, opts)
}
