package quota

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_listQuota(t *testing.T) {
	roles := `{"roles":[{
			 "name":"tonytest",
			 "weight":1.0,
			 "quota":{
				"role":"tonytest",
				"limit":{"cpus":10.0, "gpus":0.0, "mem":50000.0, "disk":200000.0001},
				"consumed":{"cpus":0.1, "mem":128.0}
			 },
			 "resources":{"cpus":0.1, "mem":128.0},
			 "allocated":{"cpus":0.1, "mem":128.0}
		  	}]}`
	smalMemRoles := `{"roles":[{
			 "name":"tonytest",
			 "weight":1.0,
			 "quota":{
				"role":"tonytest",
				"limit":{"cpus":10.0, "gpus":0.0, "mem":500.0, "disk":200.0},
				"consumed":{"cpus":0.1, "mem":128.0}
			 },
			 "resources":{"cpus":0.1, "mem":128.0},
			 "allocated":{"cpus":0.1, "mem":128.0}
		  	}]}`
	noLimitRoles := `{"roles":[{
			 "name":"tonytest",
			 "weight":1.0,
			 "quota":{
				"role":"tonytest",
				"limit":{"cpus":"NaN", "gpus":0.0, "mem":500.0, "disk":200.0},
				"consumed":{"cpus":0.1, "mem":128.0}
			 },
			 "resources":{"cpus":0.1, "mem":128.0},
			 "allocated":{"cpus":0.1, "mem":128.0}
		  	}]}`
	groups := `{"id":"/", "groups":[{
			 "id":"/tonytest",
			 "apps":[{
				   "id":"/tonytest/x",
				   "cpus":0.1,
				   "disk":0.0,
				   "instances":1
			}]}]}`

	tests := []struct {
		name           string
		marathonGroups string
		mesosRoles     string
		err            string
		jsonOutput     bool
		output         string
	}{
		{name: "JSON output", marathonGroups: groups, mesosRoles: roles, jsonOutput: true,
			output: `[{"role": "tonytest","consumed": {"cpus": 0.1,"mem": 128},"limit": {"cpus": 10, "disk":200000.0001, "gpus": 0,"mem": 50000}}]`},
		{name: "JSON output (small mem)", marathonGroups: groups, mesosRoles: smalMemRoles, jsonOutput: true,
			output: `[{"role": "tonytest","consumed": {"cpus": 0.1,"mem": 128},"limit": {"cpus": 10,"disk":200,"gpus": 0,"mem": 500}}]`},
		{name: "JSON output (no limit)", marathonGroups: groups, mesosRoles: noLimitRoles, jsonOutput: true,
			output: `[{"role": "tonytest","consumed": {"cpus": 0.1,"mem": 128},"limit": {"cpus": "NaN","disk":200,"gpus": 0,"mem": 500}}]`},
		{name: "table output", marathonGroups: groups, mesosRoles: roles,
			output: "    NAME        CPU CONSUMED            MEMORY CONSUMED             DISK CONSUMED            GPU CONSUMED      \n" +
				"  tonytest  1.00% (0 of 10 Cores)  0.26% (0.1 GiB of 50 GiB)  0.00% (0 GiB of 200.0 GiB)  NaN% (0 of 0 Cores)  \n"},
		{name: "table output (small mem)", marathonGroups: groups, mesosRoles: smalMemRoles,
			output: "    NAME        CPU CONSUMED             MEMORY CONSUMED             DISK CONSUMED           GPU CONSUMED      \n" +
				"  tonytest  1.00% (0 of 10 Cores)  25.60% (128 MiB of 500 MiB)  0.00% (0 MiB of 200 MiB)  NaN% (0 of 0 Cores)  \n"},
		{name: "table output (no limit)", marathonGroups: groups, mesosRoles: noLimitRoles,
			output: "    NAME    CPU CONSUMED        MEMORY CONSUMED             DISK CONSUMED           GPU CONSUMED      \n" +
				"  tonytest  No limit      25.60% (128 MiB of 500 MiB)  0.00% (0 MiB of 200 MiB)  NaN% (0 of 0 Cores)  \n"},
		{name: "marathon err", marathonGroups: "invalid json", mesosRoles: roles, err: "invalid character 'i' looking for beginning of value"},
		{name: "mesos err", marathonGroups: groups, mesosRoles: "invalid json", err: "invalid character 'i' looking for beginning of value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/mesos/roles" {
					w.Write([]byte(tt.mesosRoles))
					return
				}
				if r.URL.Path == "/service/marathon/v2/groups" {
					w.Write([]byte(tt.marathonGroups))
				}
			}))
			defer ts.Close()

			ctx, out := newContext(ts)
			marathonClient, err := marathon.NewClient(ctx)
			require.NoError(t, err)
			mesos, err := mesosClient(ctx)
			require.NoError(t, err)

			err = listQuota(marathonClient, mesos, tt.jsonOutput, ctx)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			}
			if tt.jsonOutput {
				assert.JSONEq(t, tt.output, out.String())
			} else {
				assert.Equal(t, tt.output, out.String())
			}
		})
	}

}

func newContext(ts *httptest.Server) (*mock.Context, *bytes.Buffer) {
	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	env.ErrOut = out
	ctx := mock.NewContext(env)
	// Use tempdir as cluster dir to check if path if properly created
	conf := config.Empty()
	conf.SetPath(path.Join(os.TempDir(), "conf.toml"))
	cluster := config.NewCluster(conf)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)
	ctx.Logger().Out = out
	ctx.Logger().SetLevel(logrus.DebugLevel)
	ctx.Logger().SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	return ctx, out
}
