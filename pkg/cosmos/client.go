package cosmos

import (
	"context"
	"strings"

	"github.com/antihax/optional"
	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
)

// Client is a diagnostics client for DC/OS.
type Client struct {
	cosmos *dcos.CosmosApiService
}

// NewClient creates a new Cosmos client.
func NewClient(ctx api.Context, baseClient *httpclient.Client) (*Client, error) {
	dcosConfigStore := dcos.NewConfigStore(&dcos.ConfigStoreOpts{
		Fs: ctx.Fs(),
	})
	dcosConfig := dcos.NewConfig(dcosConfigStore)
	dcosConfig.SetURL(baseClient.BaseURL().String())
	pluginutil.SetConfigFromEnv(dcosConfig)

	dcosClient, err := dcos.NewClientWithConfig(dcosConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		cosmos: dcosClient.Cosmos,
	}, nil
}

// PackageDescribe returns the content of '/package/describe'.
func (c *Client) PackageDescribe(name string, version string) (*Description, error) {
	desc, _, err := c.cosmos.PackageDescribe(context.TODO(), &dcos.PackageDescribeOpts{
		CosmosPackageDescribeV1Request: optional.NewInterface(dcos.CosmosPackageDescribeV1Request{
			PackageName:    name,
			PackageVersion: version,
		}),
	})
	if err != nil {
		return nil, err
	}

	backwardCompatibleDesc := Description{
		Package: desc.Package,
	}
	return &backwardCompatibleDesc, nil
}

// PackageSearch returns the packages found using the given query.
func (c *Client) PackageSearch(query string) (*SearchResult, error) {
	desc, _, err := c.cosmos.PackageSearch(context.TODO(), dcos.CosmosPackageSearchV1Request{Query: query})
	if err != nil {
		return nil, err
	}

	backwardCompatibleSearchResult := SearchResult{
		Packages: desc.Packages,
	}
	return &backwardCompatibleSearchResult, nil
}

// PackageList returns the packages installed in a cluster.
func (c *Client) PackageList() (*[]Package, error) {
	desc, _, err := c.cosmos.PackageList(context.TODO(), &dcos.PackageListOpts{
		CosmosPackageListV1Request: optional.NewInterface(dcos.CosmosPackageListV1Request{})})
	if err != nil {
		return nil, err
	}

	var ps []Package
	pAppNames := make(map[string]int)
	index := 0
	for _, p := range desc.Packages {
		// The app already exists in the list, we add it to the list of apps.
		if val, ok := pAppNames[p.PackageInformation.PackageDefinition.Name]; ok {
			ps[val].Apps = append(ps[val].Apps, p.AppId)
		} else {
			pAppNames[p.PackageInformation.PackageDefinition.Name] = index
			index++
			ps = append(ps, Package{
				Apps:               []string{p.AppId},
				Command:            strings.Join(p.PackageInformation.PackageDefinition.Command.Pip, " "),
				Description:        p.PackageInformation.PackageDefinition.Description,
				Framework:          p.PackageInformation.PackageDefinition.Framework,
				Licenses:           p.PackageInformation.PackageDefinition.Licenses,
				Maintainer:         p.PackageInformation.PackageDefinition.Maintainer,
				Name:               p.PackageInformation.PackageDefinition.Name,
				PackagingVersion:   p.PackageInformation.PackageDefinition.PackagingVersion,
				PostInstallNotes:   p.PackageInformation.PackageDefinition.PostInstallNotes,
				PostUninstallNotes: p.PackageInformation.PackageDefinition.PostUninstallNotes,
				PreInstallNotes:    p.PackageInformation.PackageDefinition.PreInstallNotes,
				Scm:                p.PackageInformation.PackageDefinition.Scm,
				Selected:           p.PackageInformation.PackageDefinition.Selected,
				Tags:               p.PackageInformation.PackageDefinition.Tags,
				Version:            p.PackageInformation.PackageDefinition.Version,
				Website:            p.PackageInformation.PackageDefinition.Website,
			})
		}
	}
	if err != nil {
		return nil, err
	}
	return &ps, nil
}
