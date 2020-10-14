package marathon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/sirupsen/logrus"
)

const UnfulfilledRole = "UnfulfilledRole"
const UnfulfilledConstraint = "UnfulfilledConstraint"
const InsufficientCpus = "InsufficientCpus"
const InsufficientMemory = "InsufficientMemory"
const InsufficientDisk = "InsufficientDisk"
const InsufficientPorts = "InsufficientPorts"
const DeclinedScarceResources = "DeclinedScarceResources"

var httpRegexp = regexp.MustCompile("^(http|https)$")

// Client to interact with the Marathon API.
type Client struct {
	API     goMarathon.Marathon
	baseURL string
}

type ErrAppAlreadyExists struct {
	appID string
}

func (e ErrAppAlreadyExists) Error() string {
	return fmt.Sprintf("Application '/%s' already exists", strings.TrimPrefix(e.appID, "/"))
}

var ErrCannotReadAppDefinition = errors.New("cannot read app definition")

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

	config := goMarathon.NewDefaultConfig()
	config.URL = baseURL
	config.HTTPClient = pluginutil.NewHTTPClient()
	if ctx.Logger().IsLevelEnabled(logrus.InfoLevel) {
		config.LogOutput = ctx.Logger().Out
	}

	client, err := goMarathon.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{API: client, baseURL: baseURL}, nil
}

// AddApp creates a deployment from the app definition referenced in appFile which can be a local
// file name or an HTTP URL. If appFile is empty, the definition is read from ctx.Input().
func (c *Client) AddApp(ctx api.Context, appLocation string) (*goMarathon.Application, error) {
	appBytes, err := getAppDefinition(ctx, appLocation)
	if err != nil {
		return nil, ErrCannotReadAppDefinition
	}

	var app map[string]interface{}
	err = json.Unmarshal(appBytes, &app)
	if err != nil {
		return nil, err
	}

	id, ok := app["id"]
	if !ok {
		return nil, fmt.Errorf("application ID must be set")
	}

	existingApp, err := c.API.ApplicationBy(id.(string), &goMarathon.GetAppOpts{})
	if err != nil {
		if apiErr, ok := err.(*goMarathon.APIError); !ok {
			return nil, err
		} else if apiErr.ErrCode != goMarathon.ErrCodeNotFound {
			return nil, err
		}
	}
	if existingApp != nil {
		return nil, ErrAppAlreadyExists{appID: existingApp.ID}
	}

	var resApp goMarathon.Application
	err = c.API.ApiPost("v2/apps/", app, &resApp)
	return &resApp, err
}

// getAppDefinition loads the app definition JSON from either a file pointed to by location or an HTTP
// URL pointed to by location or, if location is empty, from stdin.
func getAppDefinition(ctx api.Context, location string) ([]byte, error) {
	var appBytes []byte
	var err error

	if location == "" {
		appBytes, err = ioutil.ReadAll(ctx.Input())
		if err != nil {
			return nil, err
		}
		return appBytes, nil
	}

	url, err := url.Parse(location)
	if err == nil && httpRegexp.MatchString(url.Scheme) {
		resp, err := http.Get(location) // nolint: gosec // using the user-provided URL here is exactly what we want
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("could not fetch app definition from %s", location)
		}
		return ioutil.ReadAll(resp.Body)
	}

	_, err = os.Stat(location)
	if err == nil {
		return ioutil.ReadFile(location)
	}

	return nil, err
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

// Info returns the content of the Marathon endpoint 'v2/info'.
func (c *Client) Info() (map[string]interface{}, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read response body: %s", err)
		}

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal response body: %s", err)
		}
		return result, nil
	default:
		return nil, httpResponseToError(resp)
	}
}

// RawQueue returns the content of a Marathon endpoint url.
func (c *Client) RawQueue() (*RawQueue, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/queue?embed=lastUnusedOffers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read response body: %s", err)
		}

		var result RawQueue
		err = json.Unmarshal(data, &result)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal response body: %s", err)
		}
		return &result, nil
	default:
		return nil, httpResponseToError(resp)
	}
}

// KillTasks kill all the tasks of a Marathon app.
func (c *Client) KillTasks(appID string, host string) (map[string]interface{}, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	url := fmt.Sprintf("/v2/apps/%s/tasks", strings.Trim(appID, "/"))
	if host != "" {
		url += "?host=" + host
	}

	resp, err := dcosClient.Delete(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read response body: %s", err)
		}

		var tasksKilled map[string]interface{}
		err = json.Unmarshal(data, &tasksKilled)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal response body: %s", err)
		}
		return tasksKilled, nil
	case 404:
		return nil, fmt.Errorf("app '/%s' does not exist", appID)
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

func (c *Client) Applications() ([]goMarathon.Application, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/apps")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var applications goMarathon.Applications

		err = json.NewDecoder(resp.Body).Decode(&applications)
		return applications.Apps, err
	default:
		return nil, errors.New("unable to get Marathon apps")
	}
}

func (c *Client) Deployments() ([]goMarathon.Deployment, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/deployments")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var result []goMarathon.Deployment
		err = json.NewDecoder(resp.Body).Decode(&result)
		return result, err
	default:
		return nil, errors.New("unable to get Marathon deployments")
	}
}

func (c *Client) Queue() (goMarathon.Queue, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/queue")
	if err != nil {
		return goMarathon.Queue{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var result goMarathon.Queue
		err = json.NewDecoder(resp.Body).Decode(&result)
		return result, err
	default:
		return goMarathon.Queue{}, errors.New("unable to get Marathon queue")
	}
}

func (c *Client) QueueWithLastUnusedOffers() (goMarathon.Queue, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get("/v2/queue?embed=lastUnusedOffers")
	if err != nil {
		return goMarathon.Queue{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var result goMarathon.Queue
		err = json.NewDecoder(resp.Body).Decode(&result)
		return result, err
	default:
		return goMarathon.Queue{}, errors.New("unable to get Marathon queue")
	}
}

func (c *Client) ApplicationVersions(appID string) (goMarathon.ApplicationVersions, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	resp, err := dcosClient.Get(fmt.Sprintf("/v2/apps%s/versions", NormalizeAppID(appID)))
	if err != nil {
		return goMarathon.ApplicationVersions{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var result goMarathon.ApplicationVersions
		err = json.NewDecoder(resp.Body).Decode(&result)
		return result, err
	default:
		return goMarathon.ApplicationVersions{}, fmt.Errorf("unable to get versions for Marathon app %s", appID)
	}
}

func (c *Client) ApplicationByVersion(appID string, version string) (*goMarathon.Application, error) {
	dcosClient := pluginutil.HTTPClient(c.baseURL)
	url := fmt.Sprintf("/v2/apps%s", NormalizeAppID(appID))
	if version != "" {
		url = fmt.Sprintf("/v2/apps%s/versions/%s", NormalizeAppID(appID), version)
	}

	resp, err := dcosClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		if version == "" {
			var result struct {
				App goMarathon.Application `json:"app"`
			}
			err = json.NewDecoder(resp.Body).Decode(&result)
			return &result.App, err
		}
		var result goMarathon.Application
		err = json.NewDecoder(resp.Body).Decode(&result)
		return &result, err
	case 404:
		return nil, fmt.Errorf("app '%s' does not exist", NormalizeAppID(appID))
	case 422:
		return nil, fmt.Errorf("invalid timestamp provided '%s', expecting ISO-8601 datetime string", version)
	default:
		return nil, fmt.Errorf("unable to get version %s for app %s", version, appID)
	}
}

// NormalizeAppID will return a string with the correct Application ID form based on the given string
func NormalizeAppID(appID string) string {
	return fmt.Sprintf("/%s", strings.Trim(appID, "/"))
}
