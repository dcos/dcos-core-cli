package python

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dcos/dcos-cli/api"
)

func InvokePythonCLI(ctx api.Context) error {
	pythonBinaryPath, _ := ctx.EnvLookup("DCOS_CLI_EXPERIMENTAL_PATH_DCOS_PY")
	if len(strings.TrimSpace(pythonBinaryPath)) == 0 {
		executablePath, err := os.Executable()
		if err != nil {
			return err
		}

		pythonBinaryPath = filepath.Join(filepath.Dir(executablePath), "dcos_py")
		if runtime.GOOS == "windows" {
			pythonBinaryPath += ".exe"
		}
	}

	if _, err := os.Stat(pythonBinaryPath); os.IsNotExist(err) {
		return errors.New(pythonBinaryPath + " does not exist")
	}

	execCmd := exec.Command(pythonBinaryPath, ctx.Args()[1:]...)
	execCmd.Stdout = ctx.Out()
	execCmd.Stderr = ctx.ErrOut()
	execCmd.Stdin = ctx.Input()

	err := execCmd.Run()

	if execCmd.ProcessState.Exited() {
		exitCode := execCmd.ProcessState.ExitCode()
		os.Exit(exitCode)
	}

	return err
}
