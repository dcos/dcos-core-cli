package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestListPackages(t *testing.T) {
	var out bytes.Buffer
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	env.Out = &out
	ctx := mock.NewContext(env)

	client := &mocks.Client{}
	packages := []cosmos.Package{
		{
			Name:        "package-1",
			Apps:        []string{"a", "b", "c"},
			Description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas cursus nec diam non fringilla. Duis consectetur sem vitae mi congue, et ultrices mauris mattis. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Suspendisse maximus bibendum neque, eget congue augue tristique sit amet. Duis porta molestie eros, et pellentesque tellus condimentum vel. Curabitur maximus velit condimentum justo bibendum, vel sollicitudin nibh varius. Duis euismod iaculis sem, ut lobortis ex venenatis nec. Proin at semper eros, et dignissim lorem. Vivamus id scelerisque risus. Nullam auctor et est ut aliquet. Donec sit amet sem velit. ",
			Version:     "0.0.0.1",
		},
		{
			Name:        "package-3",
			Description: "Lorem ipsum dolor sit amet, \nconsectetur adipiscing elit. \nMaecenas cursus nec diam non fringilla. Duis consectetur sem vitae mi congue, et ultrices mauris mattis. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Suspendisse maximus bibendum neque, eget congue augue tristique sit amet. Duis porta molestie eros, et pellentesque tellus condimentum vel. Curabitur maximus velit condimentum justo bibendum, vel sollicitudin nibh varius. Duis euismod iaculis sem, ut lobortis ex venenatis nec. Proin at semper eros, et dignissim lorem. Vivamus id scelerisque risus. Nullam auctor et est ut aliquet. Donec sit amet sem velit. ",
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
	assert.Equal(t, `    NAME     VERSION  CERTIFIED  APP  COMMAND                               DESCRIPTION                               
  package-1  0.0.0.1  false      a    ---      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas ...  
                                 b                                                                                    
                                 c                                                                                    
  package-2  0.0.1    false      ---  xyz      XYZ                                                                    
  package-3      0.1  false      ---  ---      Lorem ipsum dolor sit amet, ...                                        
`, out.String())
}

func TestListPackagesFilterJson(t *testing.T) {
	var out bytes.Buffer
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	env.Out = &out
	ctx := mock.NewContext(env)

	client := &mocks.Client{}
	packages := []cosmos.Package{
		{Name: "package-1", Apps: []string{"a", "b", "some-app"}, Description: "package-2", Version: "0.0.0.1"},
		{Name: "package-2", Command: &dcos.CosmosPackageCommand{Name: "xyz"}, Description: "XYZ", Version: "0.0.1"},
	}
	client.On("PackageList").Return(packages, nil)

	var testCases = []struct {
		in  listOptions
		out string
	}{
		{listOptions{jsonOutput: true}, `[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"},{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
		{listOptions{query: "package-2", jsonOutput: true}, `[{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
		{listOptions{query: "-2", jsonOutput: true}, `[{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
		{listOptions{query: "package", jsonOutput: true}, `[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"},{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
		{listOptions{query: "a", jsonOutput: true}, `[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"},{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
		{listOptions{query: "some-app", jsonOutput: true}, `[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
		{listOptions{appID: "some-app", jsonOutput: true}, `[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
		{listOptions{appID: "a", jsonOutput: true}, `[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
		{listOptions{appID: "b", jsonOutput: true}, `[{"apps":["a","b","some-app"],"description":"package-2","framework":false,"name":"package-1","selected":false,"version":"0.0.0.1"}]`},
		{listOptions{appID: "package-2", jsonOutput: true}, `[{"command":{"name":"xyz"},"description":"XYZ","framework":false,"name":"package-2","selected":false,"version":"0.0.1"}]`},
		{listOptions{cliOnly: true, jsonOutput: true}, `[]`},
		{listOptions{appID: "not found", jsonOutput: true}, `[]`},
		{listOptions{query: "not found", jsonOutput: true}, `[]`},
	}
	for _, tt := range testCases {
		t.Run(fmt.Sprintf("%#v", tt.in), func(t *testing.T) {
			var out bytes.Buffer
			env.Out = &out

			err := listPackages(ctx, tt.in, client)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.out, out.String(), out.String())
		})
	}
}

func TestListErrorWhenNoPackageFound(t *testing.T) {
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	ctx := mock.NewContext(env)

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
			var out bytes.Buffer
			env.Out = &out

			err := listPackages(ctx, options, client)
			assert.EqualError(t, err, "cannot find packages matching the provided filter")
			assert.Empty(t, out.Bytes())
		})
	}
}

func TestListErrorWhenNoPackageReturned(t *testing.T) {
	var out bytes.Buffer
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	env.Out = &out
	ctx := mock.NewContext(env)

	client := &mocks.Client{}
	client.On("PackageList").Return(nil, nil)

	err := listPackages(ctx, listOptions{}, client)
	assert.EqualError(t, err, "cannot find packages matching the provided filter")
	assert.Empty(t, out.Bytes())
}

func TestListErrorWhenCosmosError(t *testing.T) {
	var out bytes.Buffer
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	env.Out = &out
	ctx := mock.NewContext(env)

	client := &mocks.Client{}
	client.On("PackageList").Return(nil, errors.New("cosmos error"))

	err := listPackages(ctx, listOptions{}, client)
	assert.EqualError(t, err, "cosmos error")
	assert.Empty(t, out.Bytes())
}
