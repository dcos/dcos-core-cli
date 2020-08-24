package cosmos

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"runtime"
	"sort"

	"github.com/antihax/optional"
	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"

	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
)

// Client is a Cosmos client abstraction
type Client interface {
	PackageDescribe(name string, version string) (*Description, error)
	PackageList() ([]Package, error)
	PackageListVersions(name string) ([]string, error)
	PackageRender(appID string, name string, version string, optionsPath string) (map[string]interface{}, error)
	PackageSearch(query string) (*SearchResult, error)
	PackageAddRepo(name string, uri string, index *int) ([]dcos.CosmosPackageRepo, error)
	PackageDeleteRepo(name string) error
	PackageListRepo() (*dcos.CosmosPackageListRepoV1Response, error)
	PackageInstall(appID string, name string, version string, optionsPath string) error
	PackageUninstall(packageName string, all bool, appID string) (*dcos.CosmosPackageUninstallV1Response, error)
}

type client struct {
	cosmos *dcos.CosmosApiService
}

// NewClient creates a new Cosmos client.
func NewClient(ctx api.Context, baseClient *httpclient.Client) (Client, error) {
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
	return &client{
		cosmos: dcosClient.Cosmos,
	}, nil
}

// PackageDescribe returns the content of '/package/describe'.
func (c *client) PackageDescribe(name string, version string) (*Description, error) {
	desc, _, err := c.cosmos.PackageDescribe(context.TODO(), &dcos.PackageDescribeOpts{
		CosmosPackageDescribeV1Request: optional.NewInterface(dcos.CosmosPackageDescribeV1Request{
			PackageName:    name,
			PackageVersion: version,
		}),
	})
	if err != nil {
		return nil, cosmosErrUnwrap(err)
	}

	backwardCompatibleDesc := Description{
		Package: desc.Package,
	}
	return &backwardCompatibleDesc, nil
}

// PackageList returns the packages installed in a cluster.
func (c *client) PackageList() ([]Package, error) {
	desc, _, err := c.cosmos.PackageList(context.TODO(), &dcos.PackageListOpts{
		CosmosPackageListV1Request: optional.NewInterface(dcos.CosmosPackageListV1Request{})})
	if err != nil {
		return nil, err
	}

	ps := make([]Package, 0, len(desc.Packages))
	pAppNames := make(map[string]int)
	index := 0
	for _, p := range desc.Packages {
		pd := p.PackageInformation.PackageDefinition
		nameWithVersion := pd.Name + pd.Version
		if val, ok := pAppNames[nameWithVersion]; ok {
			ps[val].Apps = append(ps[val].Apps, p.AppId)
		} else {
			pAppNames[nameWithVersion] = index
			index++

			newPackage := Package{
				Apps:               []string{p.AppId},
				Description:        pd.Description,
				Framework:          pd.Framework,
				Licenses:           pd.Licenses,
				Maintainer:         pd.Maintainer,
				Name:               pd.Name,
				PackagingVersion:   pd.PackagingVersion,
				PostInstallNotes:   pd.PostInstallNotes,
				PostUninstallNotes: pd.PostUninstallNotes,
				PreInstallNotes:    pd.PreInstallNotes,
				Scm:                pd.Scm,
				Selected:           pd.Selected,
				Tags:               pd.Tags,
				Version:            pd.Version,
				Website:            pd.Website,
			}

			if pd.PackagingVersion < "4.0" {
				newPackage.Command = &dcos.CosmosPackageCommand{Name: pd.Name}
				newPackage.ReleaseVersion = &pd.ReleaseVersion
			}

			ps = append(ps, newPackage)
		}
	}

	return ps, nil
}

// PackageListVersions returns the versions of a package.
func (c *client) PackageListVersions(name string) ([]string, error) {
	list, _, err := c.cosmos.PackageListVersions(context.TODO(), dcos.CosmosPackageListVersionsV1Request{
		PackageName:            name,
		IncludePackageVersions: true,
	})

	if err != nil {
		return nil, cosmosErrUnwrap(err)
	}

	var versions []string
	for version := range list.Results {
		versions = append(versions, version)
	}

	sort.Strings(versions)

	return versions, nil
}

// PackageRender returns a rendered package.
func (c *client) PackageRender(appID string, name string, version string, optionsPath string) (map[string]interface{}, error) {
	var optionsInterface map[string]interface{}
	if optionsPath != "" {
		options, err := ioutil.ReadFile(optionsPath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(options, &optionsInterface)
		if err != nil {
			log.Fatal(err)
		}
	}

	render, _, err := c.cosmos.PackageRender(context.TODO(), &dcos.PackageRenderOpts{
		CosmosPackageRenderV1Request: optional.NewInterface(dcos.CosmosPackageRenderV1Request{
			AppId:          appID,
			PackageName:    name,
			PackageVersion: version,
			Options:        optionsInterface,
		}),
	})

	if err != nil {
		return nil, cosmosErrUnwrap(err)
	}

	return render.MarathonJson, nil
}

// PackageInstall installs package
func (c *client) PackageInstall(appID string, name string, version string, optionsPath string) error {
	var optionsInterface map[string]interface{}
	if optionsPath != "" {
		options, err := ioutil.ReadFile(optionsPath)
		if err != nil {
			return err
		}

		err = json.Unmarshal(options, &optionsInterface)
		if err != nil {
			return err
		}
	}

	_, _, err := c.cosmos.PackageInstall(context.TODO(), dcos.CosmosPackageInstallV1Request{
		AppId:          appID,
		PackageName:    name,
		PackageVersion: version,
		Options:        optionsInterface,
	})
	return cosmosErrUnwrap(err)
}

// PackageSearch returns the packages found using the given query.
func (c *client) PackageSearch(query string) (*SearchResult, error) {
	desc, _, err := c.cosmos.PackageSearch(context.TODO(), dcos.CosmosPackageSearchV1Request{Query: query})
	if err != nil {
		return nil, err
	}

	backwardCompatibleSearchResult := SearchResult{
		Packages: desc.Packages,
	}
	return &backwardCompatibleSearchResult, nil
}

// PackageAddRepo adds a package repository.
func (c *client) PackageAddRepo(name string, uri string, index *int) ([]dcos.CosmosPackageRepo, error) {
	addRepoRequest := dcos.CosmosPackageAddRepoV1Request{
		Name: name,
		Uri:  uri,
	}

	if index != nil {
		index32 := int32(*index)
		addRepoRequest.Index = &index32
	}

	desc, _, err := c.cosmos.PackageRepositoryAdd(context.TODO(), &dcos.PackageRepositoryAddOpts{
		CosmosPackageAddRepoV1Request: optional.NewInterface(addRepoRequest),
	})
	if err != nil {
		return nil, cosmosErrUnwrap(err)
	}
	return desc.Repositories, nil
}

// PackageDeleteRepo deletes a package repository.
func (c *client) PackageDeleteRepo(name string) error {
	_, _, err := c.cosmos.PackageRepositoryDelete(context.TODO(), &dcos.PackageRepositoryDeleteOpts{
		CosmosPackageDeleteRepoV1Request: optional.NewInterface(dcos.CosmosPackageDeleteRepoV1Request{
			Name: name,
		}),
	})

	return cosmosErrUnwrap(err)
}

// PackageListRepo returns a list of package repositories.
func (c *client) PackageListRepo() (*dcos.CosmosPackageListRepoV1Response, error) {
	desc, _, err := c.cosmos.PackageRepositoryList(context.TODO(), nil)
	if err != nil {
		return nil, cosmosErrUnwrap(err)
	}
	return &desc, nil
}

// PackageUninstall uninstalls the given package/appID from the cluster
func (c *client) PackageUninstall(packageName string, all bool, appID string) (*dcos.CosmosPackageUninstallV1Response, error) {
	reqParams := dcos.CosmosPackageUninstallV1Request{
		All:         all,
		AppId:       appID,
		PackageName: packageName,
	}
	resp, _, err := c.cosmos.PackageUninstall(context.TODO(), reqParams, &dcos.PackageUninstallOpts{})
	return &resp, cosmosErrUnwrap(err)
}

func cosmosErrUnwrap(err error) error {
	switch err := err.(type) {
	case dcos.GenericOpenAPIError:
		if err.Model() != nil {
			if e, ok := err.Model().(dcos.CosmosError); ok {
				return fmt.Errorf(e.Message)
			}
		}
		return err
	default:
		return err
	}
}

// CLIPluginInfo extracts plugin resource data from the Cosmos package and for the current platform.
func CLIPluginInfo(pkg dcos.CosmosPackage, baseURL *url.URL) (cliArtifact dcos.CosmosPackageResourceCliArtifact, err error) {
	switch runtime.GOOS {
	case "linux":
		cliArtifact = pkg.Resource.Cli.Binaries.Linux.X8664
	case "darwin":
		cliArtifact = pkg.Resource.Cli.Binaries.Darwin.X8664
	case "windows":
		cliArtifact = pkg.Resource.Cli.Binaries.Windows.X8664
	default:
		return cliArtifact, fmt.Errorf("%s is not supported", runtime.GOOS)
	}
	// Workaround for a Cosmos bug leading to wrong schemes in plugin resource URLs.
	// This happens on setups with TLS termination proxies, where Cosmos might rewrite
	// the scheme to HTTP while it is actually HTTPS. The other way around is also possible.
	// See https://jira.mesosphere.com/browse/COPS-3052 for more context.
	//
	// To prevent this we're rewriting such URLs with the scheme set in `core.dcos_url`.
	pluginURL, err := url.Parse(cliArtifact.Url)
	if err != nil {
		return cliArtifact, err
	}
	if pluginURL.Hostname() == baseURL.Hostname() {
		pluginURL.Scheme = baseURL.Scheme
		cliArtifact.Url = pluginURL.String()
	}
	if cliArtifact.Url == "" {
		err = fmt.Errorf("'%s' isn't available for '%s')", pkg.Name, runtime.GOOS)
	}
	return cliArtifact, err
}
