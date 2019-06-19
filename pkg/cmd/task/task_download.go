package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

const description = `Downloads files from the sandbox of a given task.
The <path> argument can specify a pattern for a single or multiple files.
If no <path> is provided, the entire sandbox will be downloaded.`

// newCmdTaskDownload downloads files from a tasks sandbox.
func newCmdTaskDownload(ctx api.Context) *cobra.Command {
	var targetDir string

	cmd := &cobra.Command{
		Use:   "download <task-id> [<path>]",
		Short: description,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := findTask(ctx, args[0])
			if err != nil {
				return err
			}

			status := task.Statuses[0]
			containerID := status.ContainerStatus.GetContainerID().Value

			executorID := task.TaskID.Value
			if task.ExecutorID != nil {
				executorID = task.ExecutorID.Value
			}

			c := mesos.NewClient(pluginutil.HTTPClient(""))
			paths, err := c.Debug(task.AgentID.Value)
			if err != nil {
				return err
			}

			path, err := getPath(paths, task.FrameworkID.Value, executorID, containerID)
			if err != nil {
				return err
			}

			pattern := "/"
			if len(args) == 2 {
				pattern = args[1]
			}

			return download(c, ctx, task.AgentID.Value, path, targetDir, pattern)
		},
	}

	dir, err := os.Getwd()
	if err == nil {
		targetDir = dir
	}
	cmd.Flags().StringVar(&targetDir, "target-dir", targetDir, "Target directory of the download. Defaults to $PWD")
	return cmd
}

func getPath(paths map[string]string, framework, executor, container string) (string, error) {
	for k, v := range paths {
		if strings.Contains(k, "/frameworks/"+framework+"/executors/"+executor+"/runs/"+container) {
			return v, nil
		}
	}
	return "", fmt.Errorf("unable to find task")
}

func download(c *mesos.Client, ctx api.Context, agentID, agentPath, targetDir, pattern string) error {
	if err := ctx.Fs().MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	parent := parentPath(agentPath, pattern)
	base := basePath(pattern)
	files, err := c.Browse(agentID, parent)
	if err != nil {
		return err
	}

	errors := make(chan error, len(files))
	for _, file := range files {
		matched, err := matchFile(base, filepath.Base(file.Path))
		if err != nil {
			return fmt.Errorf("'%s' as pattern not supported: %v", pattern, err)
		}

		if !matched {
			errors <- nil
			continue
		}

		// TODO(rgo3): source out closure to separate func with config struct
		go func(f mesos.File) {
			if strings.HasPrefix(f.Mode, "d") {
				err = download(c, ctx, agentID, f.Path, filepath.Join(targetDir, filepath.Base(f.Path)), "/")
				if err != nil {
					errors <- err
					return
				}
			} else {
				b, err := c.Download(agentID, f.Path)
				if err != nil {
					errors <- fmt.Errorf("could not fetch file '%s': %v", filepath.Base(f.Path), err)
					return
				}

				out, err := ctx.Fs().Create(filepath.Join(targetDir, filepath.Base(f.Path)))
				if err != nil {
					errors <- fmt.Errorf("unable to create file '%s'", filepath.Base(f.Path))
					return
				}
				defer out.Close()
				_, err = out.Write(b)
				if err != nil {
					errors <- fmt.Errorf("unable to write to file '%s'", filepath.Base(f.Path))
					return
				}
			}
			errors <- nil
		}(file)
	}

	var failed bool
	for range files {
		err := <-errors
		if err != nil {
			failed = true
			fmt.Fprintf(ctx.ErrOut(), "Error: %v\n", err)
		}
	}
	if failed {
		return fmt.Errorf("could not download all matched files")
	}

	return nil
}

func parentPath(agentPath, pattern string) string {
	parent := filepath.Dir(pattern)
	if parent == "/" || parent == "." {
		return agentPath
	}
	return filepath.Join(agentPath, parent)
}

func basePath(pattern string) string {
	base := filepath.Base(pattern)
	if base == "/" || base == "." {
		return ""
	}
	return base
}

func matchFile(pattern, file string) (bool, error) {
	if pattern == "" {
		return true, nil
	}
	return filepath.Match(pattern, file)
}
