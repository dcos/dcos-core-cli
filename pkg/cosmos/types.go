package cosmos

import "github.com/dcos/client-go/dcos"

// Description returns the backward-compatible description of a Cosmos package.
type Description struct {
	Package dcos.CosmosPackage `json:"package"`
}

// SearchResult returns the backward-compatible description of a search on Cosmos.
type SearchResult struct {
	Packages []dcos.CosmosPackageSearchDetails `json:"packages"`
}

// Package is a struct representing a package.
type Package struct {
	Apps        []string                    `json:"apps,omitempty"`
	Command     *dcos.CosmosPackageCommand  `json:"command,omitempty"`
	Config      map[string]interface{}      `json:"config,omitempty"`
	Description string                      `json:"description,omitempty"`
	Framework   bool                        `json:"framework"`
	Licenses    []dcos.CosmosPackageLicense `json:"licenses,omitempty"`
	Maintainer  string                      `json:"maintainer,omitempty"`
	Marathon    *struct {
		V2AppMustacheTemplate string `json:"v2AppMustacheTemplate,omitempty"`
	} `json:"marathon,omitempty"`
	Name               string   `json:"name,omitempty"`
	PackagingVersion   string   `json:"packagingVersion,omitempty"`
	PostInstallNotes   string   `json:"postInstallNotes,omitempty"`
	PostUninstallNotes string   `json:"postUninstallNotes,omitempty"`
	PreInstallNotes    string   `json:"preInstallNotes,omitempty"`
	Scm                string   `json:"scm,omitempty"`
	ReleaseVersion     *int64   `json:"releaseVersion,omitempty"`
	Selected           bool     `json:"selected"`
	Tags               []string `json:"tags,omitempty"`
	Version            string   `json:"version,omitempty"`
	Website            string   `json:"website,omitempty"`
}
