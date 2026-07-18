package rules

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/mewisme/agentrule/agents"
)

// Injectable for tests.
var (
	lookPath       = exec.LookPath
	runCGInstaller = installCodegraphBinary
	runCGAgents    = func(ctx context.Context, opts Options) error {
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
			ids = []string{"cursor"}
		}
		targetFlag := strings.Join(ids, ",")
		cmd := exec.CommandContext(ctx, "codegraph", "install",
			"--target="+targetFlag,
			"--location="+string(loc),
			"--yes",
		)
		cmd.Dir = opts.Cwd
		return runQuiet(cmd)
	}
	runCGInit = func(ctx context.Context, cwd string) error {
		cmd := exec.CommandContext(ctx, "codegraph", "init")
		cmd.Dir = cwd
		return runQuiet(cmd)
	}
)

func ensureCodegraphBinary(ctx context.Context, opts Options) (WorkResult, error) {
	_, err := lookPath("codegraph")
	have := err == nil
	if have && !opts.Update {
		return WorkResult{Label: "Ensure codegraph binary"}, nil
	}
	if err := runCGInstaller(ctx); err != nil {
		return WorkResult{}, fmt.Errorf("install codegraph binary: %w", err)
	}
	if _, err := lookPath("codegraph"); err != nil {
		return WorkResult{}, fmt.Errorf("codegraph still not on PATH after install")
	}
	return WorkResult{Label: "Ensure codegraph binary"}, nil
}

func installCodegraphBinary(ctx context.Context) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
			"irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex")
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c",
			"curl -fsSL https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.sh | sh")
	}
	return runQuiet(cmd)
}

func installCodegraphAgents(ctx context.Context, opts Options) (WorkResult, error) {
	if err := runCGAgents(ctx, opts); err != nil {
		return WorkResult{}, err
	}
	return WorkResult{Label: "Run: codegraph install (MCP)"}, nil
}

func initCodegraph(ctx context.Context, opts Options) (WorkResult, error) {
	if err := runCGInit(ctx, opts.Cwd); err != nil {
		return WorkResult{}, err
	}
	return WorkResult{Label: "Run: codegraph init"}, nil
}

func writeRuleWork(name string) func(context.Context, Options) (WorkResult, error) {
	return func(_ context.Context, opts Options) (WorkResult, error) {
		if err := writeRule(opts, name); err != nil {
			return WorkResult{}, err
		}
		label := ruleWriteSummary(opts, name)
		if label == "" {
			label = "Wrote rule to agent targets"
		}
		return WorkResult{Label: label}, nil
	}
}

// runQuiet runs cmd without inheriting the parent terminal I/O.
// On failure, captured output is attached to the error.
// Callers must build cmd with exec.CommandContext so cancellation reaches the process.
func runQuiet(cmd *exec.Cmd) error {
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(string(out))
	if msg == "" {
		return err
	}
	return fmt.Errorf("%w\n%s", err, msg)
}
