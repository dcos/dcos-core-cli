package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdPackageRepoImport(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "import <repo-file>",
		Short: "Import a file containing a package repository",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(args[0]); os.IsNotExist(err) {
				return fmt.Errorf("path '%s' does not exist", args[0])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			repoFile, err := os.Open(args[0])
			if err != nil {
				return err
			}

			var repo *dcos.CosmosPackageAddRepoV1Response
			err = json.NewDecoder(repoFile).Decode(&repo)
			if err != nil {
				return err
			}

			if len(repo.Repositories) == 0 {
				return errors.New("no repositories found to import")
			}

			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			for index, repo := range repo.Repositories {
				if repo.Name == "" || repo.Uri == "" {
					fmt.Fprintf(ctx.Out(), "Repo missing name or uri. Skipping.")
					continue
				}

				_, err = c.PackageAddRepo(repo.Name, repo.Uri, index)
				if err != nil {
					fmt.Fprintf(ctx.ErrOut(), "Error (%s) while adding repo '%s' (%s). Skipping.\n", err, repo.Name, repo.Uri)
					continue
				}
				fmt.Fprintf(ctx.Out(), "Added repo '%s' (%s) at index %d", repo.Name, repo.Uri, index)
			}

			return nil
		},
	}
}
