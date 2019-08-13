package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dcos/dcos-cli/pkg/httpclient"
)

// Bundle represents a bundle object received from the diagnostics API
type Bundle struct {
	ID      string    `json:"id,omitempty"`
	Size    int64     `json:"size,omitempty"` // length in bytes for regular files; 0 when Canceled or Deleted
	Status  Status    `json:"status"`
	Started time.Time `json:"started_at,omitempty"`
	Stopped time.Time `json:"stopped_at,omitempty"`
	Errors  []string  `json:"errors,omitempty"`
}

// Client is a REST API wrapper around the new Diagnostics API.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new Client.
func NewClient(baseClient *httpclient.Client) *Client {
	return &Client{
		http: baseClient,
	}
}

// List gets a list of all cluster bundles.
func (c *Client) List() ([]Bundle, error) {
	resp, err := c.http.Get("/system/health/v1/diagnostics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var bundles []Bundle
		err = json.NewDecoder(resp.Body).Decode(&bundles)
		if err != nil {
			return nil, err
		}
		return bundles, err
	default:
		return nil, httpResponseToError(resp)
	}
}

func httpResponseToError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return &httpclient.HTTPError{
		Response: resp,
	}
}
