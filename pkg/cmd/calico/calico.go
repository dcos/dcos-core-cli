package calico

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/dcos"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/networking"
	"github.com/spf13/cobra"
)

type stateResult struct {
	state *mesos.State
	err   error
}

type execContext = func(name string, arg ...string) *exec.Cmd

func runCalicoCtl(cmdContext execContext, args []string, ctx api.Context, env []string) ([]byte, error) {
	var command *exec.Cmd

	cluster, _ := ctx.Cluster()
	basePath := cluster.Dir()
	calicoCtl := path.Join(basePath, "subcommands/dcos-core-cli/env/bin", "calicoctl")
	if len(args) == 0 {
		command = cmdContext(calicoCtl, "--help")
	} else {
		command = cmdContext(calicoCtl, args...)
	}
	command.Env = append(command.Env, append(os.Environ(), env...)...)
	return command.CombinedOutput()
}

func request(url string) (*http.Response, error) {

	probeClient := httpclient.New(strings.Replace(url, "https", "http", 1) + ":12379")

	return probeClient.Get("")
}

func getMesosState(ctx api.Context) chan stateResult {
	client, _ := mesosClient(ctx)
	stateRes := make(chan stateResult)
	go func() {
		state, err := client.State()
		stateRes <- stateResult{state, err}
	}()
	return stateRes
}

func getIps(ctx api.Context) chan map[string][]string {
	cluster, _ := ctx.Cluster()
	httpClient, _ := ctx.HTTPClient(cluster, httpclient.Timeout(3*time.Second))

	ipsRes := make(chan map[string][]string)
	// Ips Start
	go func() {
		ips := make(map[string][]string)
		c := networking.NewClient(httpClient)
		nodes, err := c.Nodes()
		if err != nil {
			ctx.Logger().Debug(err)
		} else {
			for _, node := range nodes {
				ips[node.PrivateIP] = node.PublicIPs
			}
		}
		ipsRes <- ips
	}()
	return ipsRes
}

func getEnvironment(ctx api.Context) []string {
	var environmentVariables []string

	cluster, err := ctx.Cluster()
	if err != nil {
		ctx.Logger().Debug(err)
	}

	httpClient, err := ctx.HTTPClient(cluster, httpclient.Timeout(3*time.Second))
	dcosClient := dcos.NewClient(httpClient)

	//State  start
	stateRes := getMesosState(ctx)
	// State End

	ipsRes := getIps(ctx)

	url := cluster.URL()
	_, err = request(url)
	if err != nil {
		// Check out state start
		stateResult := <-stateRes
		if stateResult.err != nil {
			ctx.Logger().Debug(stateResult.err)
		}
		state := stateResult.state
		// Check out state end
		ips := <-ipsRes
		// master ip start
		masterIPRes := make(chan string)
		go func() {
			if mesosMasters, err := mesos.NewClient(httpClient).Masters(); err == nil {
				for _, master := range mesosMasters {
					if master.IP == state.Hostname {
						masterIPRes <- ips[master.IP][0]
					}
				}
			} else {
				ctx.Logger().Debug(err)

			}
		}()
		masterIP := <-masterIPRes
		url = "https://" + masterIP
		_, err = request(url)
		if err != nil {
			ctx.Logger().Debug("Calicoctl is not able to connect to the gRPC port.")
		}
	}

	if dcosVersion, err := dcosClient.Version(); err != nil {
		ctx.Logger().Debug(err)
	} else {
		if dcosVersion.DCOSVariant != "enterprise" {
			// master ip end
			environmentVariables = append(
				os.Environ(),
				fmt.Sprintf("ETCD_CUSTOM_GRPC_METADATA=authorization:token=%s", cluster.ACSToken()),
				fmt.Sprintf("ETCD_ENDPOINTS=%s:12379", url),
			)
			return environmentVariables
		}
	}

	caFilePathRes := make(chan string)
	go func() {
		caClient := NewClient(httpClient)
		cacert, caerr := caClient.getCaCertificate()

		if caerr != nil {
			ctx.Logger().Debug(caerr)
		}

		caFilePath := path.Join(cluster.Dir(), "dcos-ca.crt")
		out, fileErr := os.Create(caFilePath)
		if fileErr != nil {
			ctx.Logger().Debug(fileErr)
		}
		out.WriteString(cacert)
		out.Close()
		if err != nil {
			ctx.Logger().Debug(err)
		}
		caFilePathRes <- caFilePath
	}()

	caFilePath := <-caFilePathRes

	environmentVariables = append(
		os.Environ(),
		fmt.Sprintf("ETCD_CUSTOM_GRPC_METADATA=authorization:token=%s", cluster.ACSToken()),
		fmt.Sprintf("ETCD_ENDPOINTS=%s:12379", url),
		fmt.Sprintf("ETCD_CA_CERT_FILE=%s", caFilePath),
	)

	return environmentVariables
}

// NewCommand creates the `dcos package` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calico [command]",
		Short: "Calicoctl wrapper",
		Run: func(cmd *cobra.Command, args []string) {
			out, err := runCalicoCtl(exec.Command, args, ctx, getEnvironment(ctx))
			fmt.Print(string(out))
			if err != nil {
				ctx.Logger().Debug(err)
				fmt.Println(calicoError())
			}

		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		out, err := runCalicoCtl(exec.Command, args[1:], ctx, getEnvironment(ctx))
		fmt.Print(string(out))
		if err != nil {
			ctx.Logger().Debug(err)
			fmt.Println(calicoError())
		}
	})

	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		out, err := runCalicoCtl(exec.Command, os.Args[2:], ctx, getEnvironment(ctx))
		fmt.Print(string(out))
		if err != nil {
			ctx.Logger().Debug(err)
			fmt.Println(calicoError())
		}
		return nil
	})

	return cmd
}

func calicoError() string {
	return "CalicoCtl exited with an error"
}
