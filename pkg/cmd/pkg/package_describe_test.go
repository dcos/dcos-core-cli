package pkg

import (
	"bytes"
	"encoding/json"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos/mocks"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pkgDescribe(t *testing.T) {
	var out bytes.Buffer
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	env.Out = &out
	ctx := mock.NewContext(env)

	client := &mocks.Client{}
	path := filepath.Join("testdata", "cosmos_chronos_describe_response.json")
	raw, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	var desc cosmos.Description
	err = json.Unmarshal(raw, &desc)
	require.NoError(t, err)
	client.On("PackageDescribe", "helloworld", "").Return(&desc, nil)
	client.On("PackageListVersions", "helloworld").Return([]string{"1", "2", "3"}, nil)
	client.On("PackageRender", "", "helloworld", "", "").
		Return(map[string]interface{}{"foo": "bar"}, nil)

	var testCases = []struct {
		in  describeOptions
		out string
	}{
		{describeOptions{}, "cosmos_chronos_describe.expected.json"},
		{describeOptions{cliOnly: true}, "cosmos_chronos_describe_cli.expected.txt"},
		{describeOptions{appOnly: true}, "cosmos_chronos_describe_app.expected.txt"},
		{describeOptions{config: true}, "cosmos_chronos_describe_config.expected.json"},
		{describeOptions{appOnly: true, render: true}, "cosmos_chronos_describe_app_render.expected.json"},
		{describeOptions{allVersions: true}, "cosmos_chronos_describe_pkg_versions.expected.json"},
	}
	for _, tt := range testCases {
		t.Run(tt.out, func(t *testing.T) {
			var out bytes.Buffer
			env.Out = &out

			path := filepath.Join("testdata", tt.out)
			expected, err := ioutil.ReadFile(path)
			assert.NoError(t, err)

			err = pkgDescribe(ctx, "helloworld", tt.in, client)
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(out.Bytes()))
		})
	}
}