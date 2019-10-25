package pkg

import (
	"errors"
	"testing"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/cosmos/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSearchPackages(t *testing.T) {
	out, ctx := setupCluster("multiple_commands")

	client := &mocks.Client{}
	searchResult := cosmos.SearchResult{Packages: []dcos.CosmosPackageSearchDetails{
		{
			Name:           "package-1",
			Description:    "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas cursus nec diam non fringilla.",
			CurrentVersion: "1",
		},
		{
			Name:           "package-3",
			Description:    "Lorem ipsum dolor sit amet, \nconsectetur adipiscing elit. \nMaecenas cursus nec diam non fringilla.",
			CurrentVersion: "3",
			Framework:      true,
		},
		{
			Name:           "package-2",
			Description:    "XYZ",
			CurrentVersion: "2",
			Selected:       true,
		},
	},
	}
	client.On("PackageSearch", "").Return(&searchResult, nil)

	err := search(ctx, "", false, client)
	assert.NoError(t, err)
	assert.Equal(t, `    NAME     VERSION  CERTIFIED  FRAMEWORK                                    DESCRIPTION                                    
  package-1        1  false      false      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas cursus nec...  
  package-3        3  false      true       Lorem ipsum dolor sit amet,                                                      
                                            consectetur adipiscing elit.                                                     
                                            Maecenas cursus n...                                                             
  package-2        2  true       false      XYZ                                                                              
`, out.String())
}

func TestSearchPackagesNoPackagesFound(t *testing.T) {
	client := &mocks.Client{}
	client.On("PackageSearch", "").Return(&cosmos.SearchResult{}, nil)
	client.On("PackageSearch", "some query").Return(&cosmos.SearchResult{}, nil)

	out, ctx := setupCluster(NoPlugins)
	err := search(ctx, "", false, client)
	assert.NoError(t, err)
	assert.Equal(t, "  NAME  VERSION  CERTIFIED  FRAMEWORK  DESCRIPTION  \n", out.String())

	out, ctx = setupCluster(NoPlugins)
	err = search(ctx, "some query", false, client)
	assert.EqualError(t, err, "no packages found")
	assert.Empty(t, out.String())

	out, ctx = setupCluster(NoPlugins)
	err = search(ctx, "", true, client)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"packages": null}`, out.String())

	out, ctx = setupCluster(NoPlugins)
	err = search(ctx, "some query", true, client)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"packages": null}`, out.String())
}

func TestSearchPackagesCosmosError(t *testing.T) {
	client := &mocks.Client{}
	client.On("PackageSearch", "").Return(nil, errors.New("cosmos error"))

	for _, jsonOutput := range []bool{true, false} {
		out, ctx := setupCluster(NoPlugins)
		err := search(ctx, "", jsonOutput, client)
		assert.EqualError(t, err, "cosmos error")
		assert.Empty(t, out.String())
	}
}
