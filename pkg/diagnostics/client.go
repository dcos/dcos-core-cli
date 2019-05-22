package diagnostics

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dcos/dcos-cli/pkg/httpclient"
)

// Client is a diagnostics client for DC/OS.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new diagnostics client.
func NewClient(baseClient *httpclient.Client) *Client {
	return &Client{
		http: baseClient,
	}
}

// Units returns the units of a certain node.
func (c *Client) Units(node string) (*UnitsHealthResponseJSONStruct, error) {
	resp, err := c.http.Get(fmt.Sprintf("/system/health/v1/nodes/%s/units", node))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var units UnitsHealthResponseJSONStruct
		err = json.NewDecoder(resp.Body).Decode(&units)
		if err != nil {
			return nil, err
		}
		return &units, nil
	default:
		return nil, httpResponseToError(resp)
	}
}

// Cancel cancel a running job creating a bundle.
func (c *Client) Cancel() (*BundleGenericResponseJSONStruct, error) {
	resp, err := c.http.Post("/system/health/v1/report/diagnostics/cancel", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var cancelBundle BundleGenericResponseJSONStruct
		err = json.NewDecoder(resp.Body).Decode(&cancelBundle)
		return &cancelBundle, err
	case 503:
		var cancelBundle BundleGenericResponseJSONStruct
		err = json.NewDecoder(resp.Body).Decode(&cancelBundle)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(cancelBundle.Status)
	default:
		return nil, httpResponseToError(resp)
	}
}

// Create requests a bundle creation for given nodes on the cluster.
func (c *Client) Create(nodes []string) (*BundleCreateResponseJSONStruct, error) {
	bundleDelete := BundleCreate{Nodes: nodes}
	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(bundleDelete); err != nil {
		return nil, err
	}

	resp, err := c.http.Post("/system/health/v1/report/diagnostics/create", "application/json", &reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var createdBundle BundleCreateResponseJSONStruct
		err = json.NewDecoder(resp.Body).Decode(&createdBundle)
		return &createdBundle, err
	case 409:
		return nil, fmt.Errorf("another bundle already in progress")
	case 503:
		return nil, fmt.Errorf("Requested nodes %v not found", nodes)
	default:
		return nil, httpResponseToError(resp)
	}
}

// Get returns a given diagnostic bundle.
func (c *Client) Get(bundle string) (*http.Response, error) {
	if !strings.HasSuffix(bundle, ".zip") {
		return nil, fmt.Errorf("Format allowed bundle-*.zip")
	}

	return c.http.Get("/system/health/v1/report/diagnostics/serve/" + bundle)
}

// Delete deletes a diagnostics bundle in the cluster.
func (c *Client) Delete(bundle string) (*BundleGenericResponseJSONStruct, error) {
	if !strings.HasSuffix(bundle, ".zip") {
		return nil, fmt.Errorf("Format allowed bundle-*.zip")
	}

	resp, err := c.http.Post("/system/health/v1/report/diagnostics/delete/"+bundle, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var createdBundle BundleGenericResponseJSONStruct
		err = json.NewDecoder(resp.Body).Decode(&createdBundle)
		return &createdBundle, err
	case 404:
		return nil, fmt.Errorf("Bundle '%s' not found, unable to delete it", bundle)
	default:
		return nil, httpResponseToError(resp)
	}
}

// List returns all the diagnostics bundles.
func (c *Client) List() (map[string][]Bundle, error) {
	resp, err := c.http.Get("/system/health/v1/report/diagnostics/list/all")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var list map[string][]Bundle
		err = json.NewDecoder(resp.Body).Decode(&list)
		return list, err
	default:
		return nil, httpResponseToError(resp)
	}
}

// Status returns the status.
func (c *Client) Status() (map[string]BundleReportStatus, error) {
	resp, err := c.http.Get("/system/health/v1/report/diagnostics/status/all")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var status map[string]BundleReportStatus
		err = json.NewDecoder(resp.Body).Decode(&status)
		if err != nil {
			return nil, err
		}
		return status, nil
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
