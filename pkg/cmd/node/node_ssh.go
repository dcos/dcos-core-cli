package node

import (
	"fmt"

	"github.com/dcos/dcos-cli/pkg/dcos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/dcos/dcos-core-cli/pkg/sshclient"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeSSH(ctx api.Context) *cobra.Command {
	var leader, masterProxy bool
	var mesosID, proxyIP, privateIP string

	clientOpts:= sshclient.ClientOpts{
		Input:      ctx.Input(),
		Out:        ctx.Out(),
		ErrOut:     ctx.ErrOut(),
	}

	cmd := &cobra.Command{
		Use:   "ssh <command>",
		Short: "Establish an SSH connection to the master or agent nodes of your DC/OS cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			clientOpts.Host, err = detectHost(ctx, leader, mesosID, privateIP)
			if err != nil {
				return err
			}

			clientOpts.Proxy, err = detectProxy(ctx, masterProxy, proxyIP)
			if err != nil {
				return err
			}

			sshClient, err := sshclient.NewClient(clientOpts, pluginutil.Logger())
			if err != nil {
				return err
			}

			return sshClient.Run(args)
		},
	}

	defUser := "core"
	cluster, err := ctx.Cluster()
	if err == nil {
		if dcosConfUser, ok := cluster.Config().Get("core.ssh_user").(string); ok {
			defUser = dcosConfUser
		}
	}
	cmd.Flags().BoolVar(&leader, "leader", false, "SSH into the leading master")
	cmd.Flags().BoolVar(&masterProxy, "master-proxy", false, "Proxy the SSH connection through a master node")
	cmd.Flags().StringVar(&mesosID, "mesos-id", "", "The agent ID of a node")
	cmd.Flags().StringVar(&proxyIP, "proxy-ip", "", "Proxy the SSH connection through a different IP address")
	cmd.Flags().StringVar(&privateIP, "private-ip", "", "Agent node with the provided private IP")
	cmd.Flags().StringVar(&clientOpts.User, "user", defUser, "The SSH user")
	cmd.Flags().StringVar(&clientOpts.Config, "config-file", "", "Path to SSH configuration file")
	cmd.Flags().StringArrayVar(&clientOpts.SSHOptions, "option", nil, "The SSH options")

	return cmd
}

func detectHost(ctx api.Context, leader bool, mesosID, privateIP string) (string, error) {
	if privateIP != "" {
		return privateIP, nil
	}

	if leader {
		leader, err := mesosDNSClient().Leader()
		if err != nil {
			return "", err
		}
		if leader.IP == "" {
			return "", fmt.Errorf("invalid leader response, missing field 'ip'")
		}
		return leader.IP, nil
	}
	c, err := mesosClient(ctx)
	if err != nil {
		return "", err
	}
	state, err := c.State()
	if err != nil {
		return "", err
	}
	for _, agent := range state.Slaves {
		if mesosID == agent.ID {
			return agent.IP(), nil
		}
	}
	return "", fmt.Errorf("agent '%s' not found", mesosID)
}

func detectProxy(ctx api.Context, masterProxy bool, proxyIP string) (string, error) {
	// proxyIP is set. We check for ENV and return.
	if proxyIP != "" {
		if _, ok := ctx.EnvLookup("SSH_AUTH_SOCK"); ok {
			ctx.Logger().Warn(
				`There is no SSH_AUTH_SOCK env variable, which likely means
				you aren't running 'ssh-agent'. 'dcos node ssh' --master-proxy/--proxy-ip
				depends on 'ssh-agent' to safely use your private key
				to hop between nodes in your cluster.
				Please run 'ssh-agent', then add your private key with 'ssh-add'`,
			)
		}
		return proxyIP, nil
	}

	// check if we have a proxyIP in the DC/OS config.
	cluster, err := ctx.Cluster()
	if err != nil {
		return "", err
	}
	dcosConfProxyIP := cluster.Config().Get("core.ssh_proxy_ip")
	if dcosConfProxyIP != nil {
		return dcosConfProxyIP.(string), nil
	}

	// neither proxyIP nor masterProxy are set.
	if proxyIP == "" && !masterProxy {
		ctx.Logger().Info(
			`If your are running this command from another network than DC/OS,
			consider using "--master-proxy" or "--proxy-ip"`,
		)
		return "", nil
	}

	// masterProxy is true.
	metadata, err := dcos.NewClient(pluginutil.HTTPClient("")).Metadata()
	if err != nil {
		return "", err
	}
	if metadata.PublicIPv4 == "" {
		return "", fmt.Errorf(`cannot use "--master-proxy", failed to detect public IP for the master`)
	}
	return metadata.PublicIPv4, nil
}
