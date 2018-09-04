package metronome

import (
	"encoding/json"
	"net/url"

	"github.com/dcos/dcos-cli/pkg/httpclient"
)

// Client is a client for Cosmos.
type Client struct {
	http *httpclient.Client
}

// JobsOption is a fucntional Option to set the `embed` query parameters
type JobsOption func(query url.Values)

// EmbedActiveRun sets the `embed`option to activeRuns
func EmbedActiveRun() JobsOption {
	return func(query url.Values) {
		query.Add("embed", "activeRuns")
	}
}

// EmbedSchedule sets the `embed`option to schedules
func EmbedSchedule() JobsOption {
	return func(query url.Values) {
		query.Add("embed", "schedules")
	}
}

// EmbedHistory sets the `embed`option to history
func EmbedHistory() JobsOption {
	return func(query url.Values) {
		query.Add("embed", "history")
	}
}

// EmbedHistorySummary sets the `embed`option to historySummary
func EmbedHistorySummary() JobsOption {
	return func(query url.Values) {
		query.Add("embed", "historySummary")
	}
}

// NewClient creates a new Metronome client.
func NewClient(baseClient *httpclient.Client) *Client {
	return &Client{
		http: baseClient,
	}
}

// Jobs returns a list of all job definitions.
func (c *Client) Jobs(opts ...JobsOption) ([]Job, error) {

	req, err := c.http.NewRequest("GET", "/v1/jobs", nil, httpclient.FailOnErrStatus(true))
	if err != nil {
		return nil, err
	}

	// Add embed query parameters to the request URL
	q := req.URL.Query()
	for _, opt := range opts {
		opt(q)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jobs []Job
	err = json.NewDecoder(resp.Body).Decode(&jobs)

	return jobs, err
}
