package marathon

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/gambol99/go-marathon"
)

// Client to interact with the Marathon API.
type Client struct {
	API     marathon.Marathon
	baseURL string
}

// NewClient creates a new HTTP wrapper client to talk to the Marathon service.
func NewClient(ctx api.Context) (*Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}
	baseURL, _ := cluster.Config().Get("marathon.url").(string)
	if baseURL == "" {
		baseURL = cluster.URL() + "/service/marathon"
	}

	config := marathon.NewDefaultConfig()
	config.URL = baseURL
	config.HTTPClient = pluginutil.NewHTTPClient()

	client, err := marathon.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{API: client, baseURL: baseURL}, nil
}

// GroupsWithoutRootSlash returns the Marathon groups' names as a map without the first "/".
// This makes it easy to match Marathon groups with Mesos roles that cannot start with "/".
func (c *Client) GroupsWithoutRootSlash() (map[string]bool, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/groups")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var groups Groups
		groupsMap := make(map[string]bool)
		if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
			return nil, err
		}

		for _, group := range groups.Groups {
			groupsMap[strings.Replace(group.ID, "/", "", 1)] = true
		}
		return groupsMap, nil
	default:
		return nil, errors.New("unable to get Marathon groups")
	}

}

func (c *Client) Applications() ([]marathon.Application, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/apps")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var applications marathon.Applications

		err = json.NewDecoder(resp.Body).Decode(&applications)
		return applications.Apps, err
	default:
		return nil, errors.New("unable to get Marathon apps")
	}
}

func (c *Client) Deployments() ([]marathon.Deployment, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/deployments")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var result []marathon.Deployment
		err = json.NewDecoder(resp.Body).Decode(&result)
		return result, err
	default:
		return nil, errors.New("unable to get Marathon deployments")
	}
}

func (c *Client) Queue() (marathon.Queue, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/queue")
	if err != nil {
		return marathon.Queue{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var result marathon.Queue
		err = json.NewDecoder(resp.Body).Decode(&result)
		return result, err
	default:
		return marathon.Queue{}, errors.New("unable to get Marathon queue")
	}
}
