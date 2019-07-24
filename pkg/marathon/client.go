package marathon

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	marathon "github.com/gambol99/go-marathon"
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

	dcosClient := pluginutil.NewHTTPClient(baseURL)

	config := marathon.NewDefaultConfig()
	config.URL = baseURL
	config.HTTPClient = dcosClient

	client, err := marathon.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{API: client, baseURL: baseURL}, nil
}

// GroupsAsQuotas returns the Marathon groups' names as quotas.
func (c *Client) GroupsAsQuotas() (map[string]bool, error) {
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
		return nil, errors.New("unable to get groups")
	}

}
