package calico

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/networking"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"

	"github.com/spf13/cobra"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/dcos"
	"github.com/dcos/dcos-cli/pkg/httpclient"
)

const GrpcPort = ":12379"

// NewCommand creates the `dcos package` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calico [command]",
		Short: "Calicoctl wrapper",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := getEnvironment(ctx, GrpcPort)
			if err != nil {
				return err
			}
			return runCalicoCtl(args, ctx, env).Run()
		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		err := runCalicoCtl(args[1:], ctx, nil).Run()
		if err != nil {
			fmt.Fprint(ctx.ErrOut(), err)
			return
		}
	})

	return cmd
}

func runCalicoCtl(args []string, ctx api.Context, env []string) *exec.Cmd {
	var command *exec.Cmd
	cluster, _ := ctx.Cluster()
	basePath := cluster.Dir()
	calicoCtl := path.Join(basePath, "subcommands/dcos-core-cli/env/bin", "calicoctl")
	if len(args) == 0 {
		command = exec.Command(calicoCtl, "--help")
	} else {
		args = append([]string{"-l", ctx.Logger().Level.String()}, args...)
		command = exec.Command(calicoCtl, args...)
	}
	command.Stdin = ctx.Input()
	command.Stdout = ctx.Out()
	command.Stderr = ctx.ErrOut()
	command.Env = append(os.Environ(), env...)
	ctx.Logger().Debugf("%s %s %s", strings.Join(command.Env, " "), command.Path, strings.Join(command.Args, " "))
	return command
}

func getEnvironment(ctx api.Context, grpcPort string) ([]string, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, fmt.Errorf("can't get cluster: %s", err)
	}

	httpClient, err := ctx.HTTPClient(cluster, httpclient.Timeout(cluster.Timeout()))
	if err != nil {
		return nil, fmt.Errorf("can't get cluster client: %s", err)
	}
	dcosClient := dcos.NewClient(httpClient)

	ctx.Logger().Debugln("Get leader private IP")
	mesosClient := mesos.NewClient(httpClient)
	leader, err := mesosClient.Leader()
	if err != nil {
		return nil, fmt.Errorf("could not get leader: %s", err)
	}

	ctx.Logger().Debugln("Get nodes public IPs")
	c := networking.NewClient(httpClient)
	nodes, err := c.Nodes()
	if err != nil {
		return nil, fmt.Errorf("could not get nodes: %s", err)
	}

	host := ""
	for _, n := range nodes {
		if n.PrivateIP == leader.IP && len(n.PublicIPs) > 0 {
			host = n.PublicIPs[0]
			break
		}
	}
	_, err = request(host+grpcPort, cluster.Timeout())
	if err != nil {
		return nil, fmt.Errorf("calicoctl is not able to connect to the gRPC port: %s", err)
	}

	dcosVersion, err := dcosClient.Version()
	if err != nil {
		return nil, fmt.Errorf("could not get DC/OS version: %s", err)
	}
	if dcosVersion.DCOSVariant != "enterprise" {
		return []string{
			fmt.Sprintf("ETCD_CUSTOM_GRPC_METADATA=authorization:token=%s", cluster.ACSToken()),
			fmt.Sprintf("ETCD_ENDPOINTS=%s%s", host, grpcPort),
		}, nil
	}

	caClient := newClient(httpClient)
	caCert, err := caClient.getCaCertificate()
	if err != nil {
		return nil, fmt.Errorf("could not get certificate: %s", err)
	}

	caFilePath := path.Join(cluster.Dir(), "dcos-ca.crt")
	out, err := os.Create(caFilePath)
	defer out.Close()
	if err != nil {
		return nil, fmt.Errorf("could not create CA file: %s", err)
	}
	_, err = out.WriteString(caCert)
	if err != nil {
		return nil, fmt.Errorf("could not write CA cert to file: %s", err)
	}

	return []string{
		fmt.Sprintf("ETCD_CUSTOM_GRPC_METADATA=authorization:token=%s", cluster.ACSToken()),
		fmt.Sprintf("ETCD_ENDPOINTS=%s%s", host, grpcPort),
		fmt.Sprintf("ETCD_CA_CERT_FILE=%s", caFilePath),
	}, nil
}

func request(url string, timeout time.Duration) (*http.Response, error) {
	probeClient := pluginutil.HTTPClient(
		"http://"+url,
		httpclient.Timeout(timeout),
	)
	return probeClient.Get("")
}
