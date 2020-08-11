package pkg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos/mocks"
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
			Name:        "package-1",
			Description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas cursus nec diam non fringilla.",
			Version:     "0.0.0.2",
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
	expected := []string{
		"NAME       VERSION     CERTIFIED    APP     COMMAND                               DESCRIPTION",
		"cli-app    alpha         true       /cli-app  cli-app  Some CLI APP",
		"package-1  0.0.0.1       false      a         ---      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas ...",
		"b",
		"c",
		"package-1  0.0.0.2       false      ---       ---      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas ...",
		"package-2  0.0.1         false      ---       xyz      XYZ",
		"package-3           0.1  false      ---       ---      Lorem ipsum dolor sit amet, ...",
		"pkg-cli    2.4.4-1.15.4  true       /pkg-cli  pkg-cli  Some CLI Package",
	}

	line := 0
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		assert.Equal(t, expected[line], strings.TrimSpace(scanner.Text()))
		line++
	}
	assert.Len(t, expected, line)
}

var abPackage = cosmos.Package{
	Apps: []string{"a", "b", "some-app"}, Description: "package-2", Name: "package-1", Version: "0.0.0.1",
}
var xyzPackage = cosmos.Package{
	Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Name: "package-2", Version: "0.0.1",
}
var cliApp = cosmos.Package{
	Apps: []string{"/cli-app"}, Command: &dcos.CosmosPackageCommand{Name: "cli-app"}, Description: "Some CLI APP",
	Framework: true, Name: "cli-app", Selected: true, Version: "alpha",
}
var pkgCli = cosmos.Package{
	Apps: []string{"/pkg-cli"}, Command: &dcos.CosmosPackageCommand{Name: "pkg-cli"}, Description: "Some CLI Package",
	Framework: true, Name: "pkg-cli", Selected: true, Version: "2.4.4-1.15.4",
}
var mockPackages = []cosmos.Package{abPackage, xyzPackage}
var testCases = []struct {
	cluster        string
	options        listOptions
	cosmosResposne []cosmos.Package
	out            []cosmos.Package
}{
	{NoPlugins, listOptions{jsonOutput: true}, mockPackages, mockPackages},
	{NoPlugins, listOptions{query: "package-2", jsonOutput: true}, mockPackages, []cosmos.Package{xyzPackage}},
	{NoPlugins, listOptions{query: "-2", jsonOutput: true}, mockPackages, []cosmos.Package{xyzPackage}},
	{NoPlugins, listOptions{query: "package", jsonOutput: true}, mockPackages, mockPackages},
	{NoPlugins, listOptions{query: "a", jsonOutput: true}, mockPackages, mockPackages},
	{NoPlugins, listOptions{query: "some-app", jsonOutput: true}, mockPackages, []cosmos.Package{abPackage}},
	{NoPlugins, listOptions{appID: "some-app", jsonOutput: true}, mockPackages, []cosmos.Package{abPackage}},
	{NoPlugins, listOptions{appID: "a", jsonOutput: true}, mockPackages, []cosmos.Package{abPackage}},
	{NoPlugins, listOptions{appID: "b", jsonOutput: true}, mockPackages, []cosmos.Package{abPackage}},
	{NoPlugins, listOptions{appID: "package-2", jsonOutput: true}, mockPackages,
		[]cosmos.Package{xyzPackage}},
	{NoPlugins, listOptions{cliOnly: true, jsonOutput: true}, mockPackages, []cosmos.Package{}},
	{NoPlugins, listOptions{appID: "not found", jsonOutput: true}, mockPackages, []cosmos.Package{}},
	{NoPlugins, listOptions{query: "not found", jsonOutput: true}, mockPackages, []cosmos.Package{}},
	{MultipleCommands, listOptions{cliOnly: true, jsonOutput: true}, mockPackages, []cosmos.Package{cliApp, pkgCli}},
	{MultipleCommands, listOptions{query: "cli", cliOnly: true, jsonOutput: true}, mockPackages, []cosmos.Package{cliApp, pkgCli}},
	{MultipleCommands, listOptions{query: "app", jsonOutput: true}, mockPackages, []cosmos.Package{cliApp, abPackage}},
	{MultipleCommands, listOptions{appID: "app", jsonOutput: true}, mockPackages, []cosmos.Package{}},
	{MultipleCommands, listOptions{query: "package", cliOnly: true, jsonOutput: true}, mockPackages, []cosmos.Package{}},
	{MultipleCommands, listOptions{appID: "cli-app", jsonOutput: true}, mockPackages, []cosmos.Package{cliApp}},
	{MultipleCommands, listOptions{jsonOutput: true}, []cosmos.Package{
		{Name: "cli-app", Apps: []string{"not", "a", "cliApp"}, Description: "package-2", Version: "0.0.0.1"},
		{Name: "pkg-cli", Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Version: "0.0.1"},
	}, []cosmos.Package{
		{Name: "cli-app", Apps: []string{"not", "a", "cliApp"}, Description: "package-2", Version: "0.0.0.1"}, cliApp,
		{Name: "pkg-cli", Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Version: "0.0.1"}, pkgCli,
	}},
	{MultipleCommands, listOptions{jsonOutput: true}, []cosmos.Package{
		{Name: "cli-app", Apps: []string{"not", "a", "cliApp"}, Description: "package-2", Version: "alpha"},
		{Name: "pkg-cli", Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Version: "2.4.4-1.15.4"},
	}, []cosmos.Package{
		cliApp, {Name: "cli-app", Apps: []string{"not", "a", "cliApp"}, Description: "package-2", Version: "alpha"},
		pkgCli, {Name: "pkg-cli", Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Version: "2.4.4-1.15.4"},
	}},
	{MultipleCommands, listOptions{jsonOutput: true}, []cosmos.Package{cliApp, pkgCli}, []cosmos.Package{cliApp, pkgCli}},
	{MultipleCommands, listOptions{jsonOutput: true, cliOnly: true}, []cosmos.Package{
		{Name: "cli-app", Apps: []string{"not", "a", "cliApp"}, Description: "package-2", Version: "0.0.0.1"},
		{Name: "pkg-cli", Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Version: "0.0.1"},
	}, []cosmos.Package{cliApp, pkgCli}},
}

func TestListPackagesFilterJson(t *testing.T) {
	for _, tt := range testCases {
		t.Run(fmt.Sprintf("%s_%#v", tt.cluster, tt.options), func(t *testing.T) {
			out, ctx := setupCluster(tt.cluster)
			client := &mocks.Client{}
			client.On("PackageList").Return(tt.cosmosResposne, nil)

			err := listPackages(ctx, tt.options, client)
			assert.NoError(t, err)

			expectedJSON, err := json.MarshalIndent(tt.out, "", "    ")
			require.NoError(t, err)

			assert.Equal(t, string(expectedJSON)+"\n", out.String())
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
