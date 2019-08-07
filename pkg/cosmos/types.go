package cosmos

import "github.com/dcos/client-go/dcos"

// Description returns the backward-compatible description of a Cosmos package.
type Description struct {
	Package dcos.CosmosPackage `json:"package"`
}
