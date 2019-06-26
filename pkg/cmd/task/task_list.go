package task

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdTaskList(ctx api.Context) *cobra.Command {
	var all, jsonOutput, completed, quietOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the Mesos tasks in the cluster",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if all && completed {
				return fmt.Errorf("cannot accept both options --all and --completed")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, cobraArgs []string) error {
			exePath, err := os.Executable()
			if err != nil {
				return err
			}

			pyExePath := filepath.Join(filepath.Dir(exePath), "dcos_py")
			if runtime.GOOS == "windows" {
				pyExePath += ".exe"
			}

			args := []string{"task"}
			if all {
				args = append(args, "--all")
			}
			if jsonOutput {
				args = append(args, "--json")
			}
			if completed {
				args = append(args, "--completed")
			}
			if quietOutput {
				args = append(args, "--quiet")
			}
			if len(cobraArgs) == 1 {
				args = append(args, cobraArgs[0])
			}

			execCmd := exec.Command(pyExePath, args...)
			execCmd.Stdout = ctx.Out()
			execCmd.Stderr = ctx.ErrOut()
			execCmd.Stdin = ctx.Input()

			return execCmd.Run()
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Print completed and in-progress tasks")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	cmd.Flags().BoolVar(&completed, "completed", false, "Print completed tasks")
	cmd.Flags().BoolVarP(&quietOutput, "quiet", "q", false, "Print only IDs of listed services")
	return cmd
}
