package service

import (
	"fmt"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/logs"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/dcos/dcos-core-cli/pkg/networking"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

type ipsResult struct {
	ips map[string][]string
	err error
}

func newCmdServiceLog(ctx api.Context) *cobra.Command {
	var follow bool
	var lines int
	var file, output, sshConfig string

	cmd := &cobra.Command{
		Use:   "log <service-name> [file]",
		Short: "Print logs for DC/OS services",
		Args:  cobra.RangeArgs(1, 2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
			if !ok && sshConfig != "" {
				ctx.Deprecated("The --ssh-config-file flag is deprecated.")
			}

			switch len(args) {
			case 1:
				file = "stdout"
			case 2:
				if args[0] == "marathon" {
					return fmt.Errorf("the <file> argument is invalid for marathon")
				}
				file = args[1]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			marathonClient, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			if args[0] == "marathon" {
				url, urlErr := marathonMasterURL(marathonClient, ctx)
				if urlErr != nil {
					return urlErr
				}

				logClient := logs.NewClient(pluginutil.HTTPClient(url), ctx.Out())
				opts := logs.Options{
					Follow: follow,
					Format: output,
					Skip:   -1 * lines,
				}
				return logClient.PrintComponent("/leader/mesos", "/dcos-marathon.service", opts)
			}

			appID, err := serviceAppID(marathonClient, args[0])
			if err != nil {
				return err
			}
			taskID, err := serviceTaskID(marathonClient, appID)
			if err != nil {
				return err
			}

			logClient := logs.NewClient(pluginutil.HTTPClient(""), ctx.Out())
			if output != "short" {
				output = "short"
				ctx.Logger().Info(`Task logs don't support output options. Defaulting to "short"...`)
			}
			opts := logs.Options{
				Follow: follow,
				Format: output,
				Skip:   -1 * lines,
			}
			return logClient.PrintTask(taskID, file, opts)
		},
	}

	cmd.Flags().BoolVar(&follow, "follow", false, "Dynamically update the log")
	cmd.Flags().IntVar(&lines, "lines", 10, "Print the N last lines. 10 is the default")
	cmd.Flags().StringVarP(&output, "output", "o", "short", "Format log message output")
	cmd.Flags().StringVar(&sshConfig, "ssh-config-file", "", "Path to SSH configuration file. This is deprecated")
	cmd.Flags().String("user", "", "The SSH user for Marathon")
	return cmd
}

func marathonMasterURL(c *marathon.Client, ctx api.Context) (string, error) {
	ipsRes := make(chan ipsResult)
	go publicIPs(ipsRes)

	leader, err := c.API.Leader()
	if err != nil {
		return "", err
	}

	nodeIPs := <-ipsRes
	if nodeIPs.err != nil {
		return "", nil
	}

	// from <ip>:<port> get the <ip> field
	marathonIP := strings.Split(leader, ":")[0]

	cluster, err := ctx.Cluster()
	if err != nil {
		return "", err
	}
	// from <scheme>://<url> get the <scheme> field
	scheme := strings.Split(cluster.URL(), "://")[0]

	// check if we have public IP addresses for that private Marathon leader IP
	// return url with first available public IP
	if ips, ok := nodeIPs.ips[marathonIP]; ok {
		return fmt.Sprintf("%s://%s", scheme, ips[0]), nil
	}

	return "", fmt.Errorf("could not map private IP: [%s] to a public IP", marathonIP)
}

func publicIPs(ch chan<- ipsResult) {
	c := networking.NewClient(pluginutil.HTTPClient(""))
	nodes, err := c.Nodes()

	ips := make(map[string][]string)
	if err == nil {
		for _, node := range nodes {
			ips[node.PrivateIP] = node.PublicIPs
		}
	}
	ch <- ipsResult{ips, err}
}

func serviceAppID(c *marathon.Client, service string) (string, error) {
	apps, err := c.API.Applications(nil)
	if err != nil {
		return "", err
	}

	var appIDs []string
	for _, a := range apps.Apps {
		if (*a.Labels)["DCOS_PACKAGE_FRAMEWORK_NAME"] == service {
			appIDs = append(appIDs, a.ID)
		}
	}

	switch len(appIDs) {
	case 0:
		return "", fmt.Errorf("no marathon apps found for service name [%s]", service)
	case 1:
		return appIDs[0], nil
	default:
		return "", fmt.Errorf("multiple marathon apps found for service name [%s]: [%s]", service, strings.Join(appIDs, ", "))
	}
}

func serviceTaskID(c *marathon.Client, appID string) (string, error) {
	app, err := c.API.Application(appID)
	if err != nil {
		return "", err
	}

	switch len(app.Tasks) {
	case 1:
		return app.Tasks[0].ID, nil
	default:
		return "", fmt.Errorf("expected marathon app [%s] to be running 1 task, but instead found %d tasks", appID, len(app.Tasks))
	}
}
