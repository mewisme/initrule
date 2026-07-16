package rules

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Injectable for tests.
var (
	lookPath       = exec.LookPath
	runCGInstaller = installCodegraphBinary
	runCGInit      = func(cwd string) error {
		cmd := exec.Command("codegraph", "init", "-i")
		cmd.Dir = cwd
		return runQuiet(cmd)
	}
)

func ensureCodegraphBinary(opts Options) error {
	_, err := lookPath("codegraph")
	have := err == nil
	if have && !opts.Update {
		return nil
	}
	if err := runCGInstaller(); err != nil {
		return fmt.Errorf("install codegraph binary: %w", err)
	}
	if _, err := lookPath("codegraph"); err != nil {
		return fmt.Errorf("codegraph still not on PATH after install")
	}
	return nil
}

func installCodegraphBinary() error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
			"irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex")
	} else {
		cmd = exec.Command("sh", "-c",
			"curl -fsSL https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.sh | sh")
	}
	return runQuiet(cmd)
}

func initCodegraph(opts Options) error {
	return runCGInit(opts.Cwd)
}

// runQuiet runs cmd without inheriting the parent terminal I/O.
// On failure, captured output is attached to the error.
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
