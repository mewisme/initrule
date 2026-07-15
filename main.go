package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mewisme/initrule/rule"
	"github.com/mewisme/initrule/tui"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage:
  initrule                         interactive multi-select
  initrule install|i <name>...     install named rules
  initrule install|i -a|--all      install all rules
  initrule install|i ... -u|--update  force reinstall codegraph binary

flags (with install|i):
  -a, --all      all registered rules
  -u, --update   force codegraph binary reinstall

rules: %v
`, rule.Names())
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	opts := rule.Options{Cwd: cwd}

	if len(os.Args) == 1 {
		if err := tui.Run(opts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	cmd := os.Args[1]
	if cmd != "install" && cmd != "i" {
		usage()
		os.Exit(1)
	}

	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var all, update bool
	fs.BoolVar(&all, "a", false, "install all rules")
	fs.BoolVar(&all, "all", false, "install all rules")
	fs.BoolVar(&update, "u", false, "force codegraph binary reinstall")
	fs.BoolVar(&update, "update", false, "force codegraph binary reinstall")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}
	opts.Update = update
	names := fs.Args()

	if all && len(names) > 0 {
		fmt.Fprintln(os.Stderr, "use either -a/--all or rule names, not both")
		os.Exit(1)
	}
	if !all && len(names) == 0 {
		usage()
		os.Exit(1)
	}

	if all {
		err = rule.RunAll(opts)
	} else {
		err = rule.RunNames(names, opts)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
