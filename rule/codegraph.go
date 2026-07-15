package rule

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Injectable for tests.
var (
	lookPath       = exec.LookPath
	runCGInstaller = installCodegraphBinary
	runCGInit      = func(cwd string) error {
		cmd := exec.Command("codegraph", "init", "-i")
		cmd.Dir = cwd
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
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
		cmd = exec.Command("powershell", "-NoProfile", "-Command",
			"irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex")
	} else {
		cmd = exec.Command("sh", "-c",
			"curl -fsSL https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.sh | sh")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func initCodegraph(opts Options) error {
	return runCGInit(opts.Cwd)
}
