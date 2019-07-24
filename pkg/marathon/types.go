package marathon

// Groups represents a stripped down version of the Marathon groups
type Groups struct {
	Groups []struct {
		ID string `json:"id"`
	} `json:"groups"`
}
