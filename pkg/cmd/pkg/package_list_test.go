package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const NoPlugins = "no_plugins"
const MultipleCommands = "multiple_commands"

func TestListPackages(t *testing.T) {
	out, ctx := setupCluster("multiple_commands")

	client := &mocks.Client{}
	packages := []cosmos.Package{
		{
			Name:        "package-1",
			Apps:        []string{"a", "b", "c"},
			Description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas cursus nec diam non fringilla.",
			Version:     "0.0.0.1",
		},
		{
			Name:        "package-3",
			Description: "Lorem ipsum dolor sit amet, \nconsectetur adipiscing elit. \nMaecenas cursus nec diam non fringilla.",
			Version:     "0.1",
		},
		{
			Name:        "package-2",
			Command:     &dcos.CosmosPackageCommand{Name: "xyz"},
			Description: "XYZ",
			Version:     "0.0.1",
		},
	}
	client.On("PackageList").Return(packages, nil)

	err := listPackages(ctx, listOptions{}, client)
	assert.NoError(t, err)
	assert.Equal(t, `    NAME       VERSION     CERTIFIED    APP     COMMAND                               DESCRIPTION                               
  cli-app    alpha         true       /cli-app  cli-app  Some CLI APP                                                           
  package-1  0.0.0.1       false      a         ---      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas ...  
                                      b                                                                                         
                                      c                                                                                         
  package-2  0.0.1         false      ---       xyz      XYZ                                                                    
  package-3           0.1  false      ---       ---      Lorem ipsum dolor sit amet, ...                                        
  pkg-cli    2.4.4-1.15.4  true       /pkg-cli  pkg-cli  Some CLI Package                                                       
`, out.String())
}

var testCases = []struct {
	cluster string
	options listOptions
	out     string
}{
	{NoPlugins, listOptions{jsonOutput: true},
		`[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"},
{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
	{NoPlugins, listOptions{query: "package-2", jsonOutput: true},
		`[{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
	{NoPlugins, listOptions{query: "-2", jsonOutput: true},
		`[{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
	{NoPlugins, listOptions{query: "package", jsonOutput: true},
		`[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"},
{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
	{NoPlugins, listOptions{query: "a", jsonOutput: true},
		`[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"},
{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
	{NoPlugins, listOptions{query: "some-app", jsonOutput: true},
		`[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
	{NoPlugins, listOptions{appID: "some-app", jsonOutput: true},
		`[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
	{NoPlugins, listOptions{appID: "a", jsonOutput: true},
		`[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
	{NoPlugins, listOptions{appID: "b", jsonOutput: true},
		`[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
	{NoPlugins, listOptions{appID: "package-2", jsonOutput: true},
		`[{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
	{NoPlugins, listOptions{cliOnly: true, jsonOutput: true}, `[]`},
	{NoPlugins, listOptions{appID: "not found", jsonOutput: true}, `[]`},
	{NoPlugins, listOptions{query: "not found", jsonOutput: true}, `[]`},
	{MultipleCommands, listOptions{cliOnly: true, jsonOutput: true},
		`[{"apps":["/cli-app"],"command":{"name":"cli-app"},"description":"Some CLI APP","framework":true,
"name":"cli-app","selected":true,"version":"alpha"},
{"apps":["/pkg-cli"],"command":{"name":"pkg-cli"},"description":"Some CLI Package","framework":true,
"name":"pkg-cli","selected":true,"version":"2.4.4-1.15.4"}]`},
	{MultipleCommands, listOptions{query: "cli", cliOnly: true, jsonOutput: true},
		`[{"apps":["/cli-app"],"command":{"name":"cli-app"},"description":"Some CLI APP","framework":true,
"name":"cli-app","selected":true,"version":"alpha"},
{"apps":["/pkg-cli"],"command":{"name":"pkg-cli"},"description":"Some CLI Package","framework":true,
"name":"pkg-cli","selected":true,"version":"2.4.4-1.15.4"}]`},
	{MultipleCommands, listOptions{query: "app", jsonOutput: true},
		`[{"apps":["/cli-app"],"command":{"name":"cli-app"},"description":"Some CLI APP","framework":true,
"name":"cli-app","selected":true,"version":"alpha"},
{"apps":["a","b","some-app"],"description":"package-2","framework":false,
"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
	{MultipleCommands, listOptions{appID: "app", jsonOutput: true}, `[]`},
	{MultipleCommands, listOptions{appID: "cli-app", jsonOutput: true},
		`[{"apps":["/cli-app"],"command":{"name":"cli-app"},"description":"Some CLI APP","framework":true,
"name":"cli-app","selected":true,"version":"alpha"}]`},
	{MultipleCommands, listOptions{query: "package", cliOnly: true, jsonOutput: true}, `[]`},
}

func TestListPackagesFilterJson(t *testing.T) {
	client := &mocks.Client{}
	packages := []cosmos.Package{
		{Name: "package-1", Apps: []string{"a", "b", "some-app"}, Description: "package-2", Version: "0.0.0.1"},
		{Name: "package-2", Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Version: "0.0.1"},
	}
	client.On("PackageList").Return(packages, nil)

	for _, tt := range testCases {
		t.Run(fmt.Sprintf("%s_%#v", tt.cluster, tt.options), func(t *testing.T) {
			out, ctx := setupCluster(tt.cluster)

			err := listPackages(ctx, tt.options, client)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.out, out.String(), out.String())
		})
	}
}

func TestListErrorWhenNoPackageFound(t *testing.T) {
	client := &mocks.Client{}
	packages := []cosmos.Package{
		{Name: "package-1", Apps: []string{"a", "b", "some-app"}, Description: "package-2", Version: "0.0.0.1"},
		{Name: "package-2", Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Version: "0.0.1"},
	}
	client.On("PackageList").Return(packages, nil)

	var testCases = []listOptions{
		{cliOnly: true},
		{appID: "not found"},
		{query: "not found"},
	}
	for _, options := range testCases {
		t.Run(fmt.Sprintf("%#v", options), func(t *testing.T) {
			out, ctx := setupCluster(NoPlugins)

			err := listPackages(ctx, options, client)
			assert.EqualError(t, err, "cannot find packages matching the provided filter")
			assert.Empty(t, out.Bytes())
		})
	}
}

func TestListErrorWhenNoPackageReturned(t *testing.T) {
	out, ctx := setupCluster(NoPlugins)

	client := &mocks.Client{}
	client.On("PackageList").Return(nil, nil)

	err := listPackages(ctx, listOptions{}, client)
	assert.EqualError(t, err, "cannot find packages matching the provided filter")
	assert.Empty(t, out.Bytes())
}

func TestListErrorWhenCosmosError(t *testing.T) {
	out, ctx := setupCluster(NoPlugins)

	client := &mocks.Client{}
	client.On("PackageList").Return(nil, errors.New("cosmos error"))

	err := listPackages(ctx, listOptions{}, client)
	assert.EqualError(t, err, "cosmos error")
	assert.Empty(t, out.Bytes())
}

func TestListErrorWhenClusterError(t *testing.T) {
	out, ctx := setupCluster("invalid")

	err := listPackages(ctx, listOptions{}, nil)
	assert.EqualError(t, err, "cannot read cluster data: invalid character 'o' in literal null (expecting 'u')")
	assert.Empty(t, out.Bytes())
}

func setupCluster(name string) (*bytes.Buffer, *mock.Context) {
	var out bytes.Buffer
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	env.Out = &out
	ctx := mock.NewContext(env)
	wd, _ := os.Getwd()
	clusterDir := filepath.Join(wd, "testdata", name)
	conf := config.New(config.Opts{
		Fs: env.Fs,
	})
	conf.SetPath(filepath.Join(clusterDir, "dcos.toml"))
	ctx.SetCluster(config.NewCluster(conf))
	return &out, ctx
}
