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
