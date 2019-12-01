package marathon

// Groups represents a stripped down version of the Marathon groups
type Groups struct {
	Groups []struct {
		ID string `json:"id"`
	} `json:"groups"`
}

// RawQueue represents a queue returned by /v2/queue?embed=lastUnusedOffers
type RawQueue struct {
	Queue []map[string]interface{} `json:"queue"`
}
