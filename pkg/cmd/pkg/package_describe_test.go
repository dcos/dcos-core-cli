package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos/mocks"

	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestPkgDescribeShouldProxyCommandsToCosmos(t *testing.T) {
	client := &mocks.Client{}
	desc := cosmos.Description{
		Package: dcos.CosmosPackage{
			Description:      "Package stub with only required fields",
			Maintainer:       "John Doe",
			Name:             "helloworld",
			PackagingVersion: "0.0.1",
			ReleaseVersion:   9223372036854775807,
			Version:          "0.0.2",
			Marathon: dcos.CosmosPackageMarathon{
				V2AppMustacheTemplate: "ewogICAgImlkIjogInt7aGVsbG93b3JsZC5pZH19IiwKICAgICJjbWQiOiAiZWNobyAnSGVsbG8gV29ybGQnIgp9",
			},
			Config: map[string]interface{}{
				"foo": "bar",
			},
		},
	}
	client.On("PackageDescribe", "helloworld", "").Return(&desc, nil)
	client.On("PackageListVersions", "helloworld").Return([]string{"1", "2", "3"}, nil)
	client.On("PackageRender", "", "helloworld", "", "").
		Return(map[string]interface{}{"foo": "bar"}, nil)

	path := filepath.Join("testdata", "describe.expected.json")
	expectedDescribe, err := ioutil.ReadFile(path)
	assert.NoError(t, err)

	var testCases = []struct {
		in  describeOptions
		out string
	}{
		{describeOptions{}, string(expectedDescribe)},
		{describeOptions{cliOnly: true}, ""},
		{describeOptions{appOnly: true}, "{\n    \"id\": \"{{helloworld.id}}\",\n    \"cmd\": \"echo 'Hello World'\"\n}"},
		{describeOptions{config: true}, "{\n  \"foo\": \"bar\"\n}\n"},
		{describeOptions{appOnly: true, render: true}, "{\n  \"foo\": \"bar\"\n}\n"},
		{describeOptions{allVersions: true}, "[\n  \"1\",\n  \"2\",\n  \"3\"\n]\n"},
	}
	for _, tt := range testCases {
		t.Run(fmt.Sprintf("%#v", tt.in), func(t *testing.T) {
			ctx, out := newContext()

			err = pkgDescribe(ctx, "helloworld", tt.in, client)
			assert.NoError(t, err)
			assert.Equal(t, tt.out, string(out.Bytes()))
		})
	}
}

func TestPkgDescribeShouldDetectIfPackageSupportsCli(t *testing.T) {
	var testCases = []struct {
		in  cosmos.Description
		out string
	}{
		{cosmos.Description{Package: dcos.CosmosPackage{Command: dcos.CosmosPackageCommand{Name: "some command"}}},
			"{\n  \"name\": \"some command\"\n}\n"},
		{cosmos.Description{Package: dcos.CosmosPackage{Command: dcos.CosmosPackageCommand{Pip: []string{"deprecated"}}}},
			"{\n  \"pip\": [\n    \"deprecated\"\n  ]\n}\n"},
		{cosmos.Description{Package: dcos.CosmosPackage{Command: dcos.CosmosPackageCommand{Name: "some command", Pip: []string{"deprecated"}}}},
			"{\n  \"name\": \"some command\",\n  \"pip\": [\n    \"deprecated\"\n  ]\n}\n"},
		{cosmos.Description{Package: dcos.CosmosPackage{
			Command: dcos.CosmosPackageCommand{Name: "some command", Pip: []string{"deprecated"}},
			Resource: dcos.CosmosPackageResource{Cli: dcos.CosmosPackageResourceCli{Binaries: dcos.CosmosPackageResourceCliBinaries{
				Linux: dcos.CosmosPackageResourceCliOsBinaries{X8664: dcos.CosmosPackageResourceCliArtifact{
					Kind: "zip", Url: "http://exmaple.com"}}}}}}},
			"{\n  \"binaries\": {" +
				"\n    \"darwin\": {\n      \"x86-64\": {}\n    }," +
				"\n    \"linux\": {\n      \"x86-64\": {\n        \"kind\": \"zip\",\n        \"url\": \"http://exmaple.com\"\n      }\n    }," +
				"\n    \"windows\": {\n      \"x86-64\": {}\n    }\n  }\n}\n"},
		{cosmos.Description{Package: dcos.CosmosPackage{
			Resource: dcos.CosmosPackageResource{Cli: dcos.CosmosPackageResourceCli{Binaries: dcos.CosmosPackageResourceCliBinaries{
				Linux: dcos.CosmosPackageResourceCliOsBinaries{X8664: dcos.CosmosPackageResourceCliArtifact{
					Kind: "zip", Url: "http://exmaple.com"}}}}}}},
			"{\n  \"binaries\": {" +
				"\n    \"darwin\": {\n      \"x86-64\": {}\n    }," +
				"\n    \"linux\": {\n      \"x86-64\": {\n        \"kind\": \"zip\",\n        \"url\": \"http://exmaple.com\"\n      }\n    }," +
				"\n    \"windows\": {\n      \"x86-64\": {}\n    }\n  }\n}\n"},
	}
	for _, tt := range testCases {
		t.Run(fmt.Sprintf("%#v", tt.in), func(t *testing.T) {
			client := &mocks.Client{}
			client.On("PackageDescribe", "helloworld", "").Return(&tt.in, nil)

			ctx, out := newContext()

			err := pkgDescribe(ctx, "helloworld", describeOptions{cliOnly: true}, client)
			assert.NoError(t, err)
			assert.Equal(t, tt.out, string(out.Bytes()))
		})
	}
}

func TestPkgDescribeShouldReturnCosmosErrorOnRender(t *testing.T) {
	ctx, out := newContext()

	client := &mocks.Client{}
	desc := cosmos.Description{}
	client.On("PackageDescribe", "helloworld", "").Return(&desc, nil)
	client.On("PackageRender", "", "helloworld", "", "").Return(nil, fmt.Errorf("render error"))

	err := pkgDescribe(ctx, "helloworld", describeOptions{appOnly: true, render: true}, client)
	assert.EqualError(t, err, "render error")
	assert.Empty(t, string(out.Bytes()))
}

func TestPkgDescribeShouldReturnCosmosErrorWhenCosmosFails(t *testing.T) {
	client := &mocks.Client{}
	client.On("PackageDescribe", "helloworld", "").Return(nil, fmt.Errorf("could not describe"))
	client.On("PackageListVersions", "helloworld").Return(nil, fmt.Errorf("could not get version"))

	var testCases = []struct {
		in  describeOptions
		out string
	}{
		{describeOptions{}, "could not describe"},
		{describeOptions{cliOnly: true}, "could not describe"},
		{describeOptions{appOnly: true}, "could not describe"},
		{describeOptions{config: true}, "could not describe"},
		{describeOptions{appOnly: true, render: true}, "could not describe"},
		{describeOptions{allVersions: true}, "could not get version"},
	}
	for _, tt := range testCases {
		t.Run(fmt.Sprintf("%#v", tt.in), func(t *testing.T) {
			ctx, out := newContext()

			err := pkgDescribe(ctx, "helloworld", tt.in, client)
			assert.EqualError(t, err, tt.out)
			assert.Empty(t, string(out.Bytes()))
		})
	}
}

func TestPkgDescribeShouldReturnErrorWhenUnableToDecodeMarathonConfiguration(t *testing.T) {
	ctx, out := newContext()

	client := &mocks.Client{}
	desc := cosmos.Description{
		Package: dcos.CosmosPackage{
			Marathon: dcos.CosmosPackageMarathon{
				V2AppMustacheTemplate: "not base 64",
			},
		},
	}
	client.On("PackageDescribe", "helloworld", "").Return(&desc, nil)

	err := pkgDescribe(ctx, "helloworld", describeOptions{appOnly: true, config: true}, client)
	assert.EqualError(t, err, "illegal base64 data at input byte 3")
	assert.Empty(t, string(out.Bytes()))
}

func newContext() (*mock.Context, *bytes.Buffer) {
	var out bytes.Buffer
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	env.Out = &out
	ctx := mock.NewContext(env)
	return ctx, &out
}
