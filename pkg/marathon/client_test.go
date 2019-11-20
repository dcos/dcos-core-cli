package marathon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	marathonmocks "github.com/dcos/dcos-core-cli/pkg/marathon/mocks"

	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-cli/pkg/mock"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseToError(t *testing.T) {
	resp := http.Response{StatusCode: 200}
	err := httpResponseToError(&resp)
	assert.EqualError(t, err, `unexpected status code 200`)

	resp = http.Response{StatusCode: 404}
	err = httpResponseToError(&resp)
	assert.EqualError(t, err, `HTTP 404 error`)
}

func TestApplications(t *testing.T) {
	expected := goMarathon.Applications{
		Apps: []goMarathon.Application{
			{
				ID:      "test-id",
				Env:     new(map[string]string),
				Secrets: new(map[string]goMarathon.Secret),
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/apps", r.URL.String())
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := Client{
		baseURL: ts.URL,
	}

	apps, err := client.Applications()
	require.NoError(t, err)

	jsonEqual(t, expected.Apps, apps)
}

func TestApplicationsErrors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError string
	}{
		{
			name:          "unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: "unable to get Marathon apps",
		},
		{
			name:          "forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: "unable to get Marathon apps",
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(test.statusCode)
		}))

		client := Client{
			baseURL: ts.URL,
		}

		apps, err := client.Applications()

		assert.EqualError(t, err, test.expectedError)
		assert.Nil(t, apps)

		ts.Close()
	}
}

func TestQueue(t *testing.T) {
	expected := goMarathon.Queue{
		Items: []goMarathon.Item{
			{
				Count: 1,
				Delay: goMarathon.Delay{
					Overdue:         false,
					TimeLeftSeconds: 5,
				},
				Application: goMarathon.Application{
					ID: "test-id",
				},
			},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/queue", r.URL.String())
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := Client{
		baseURL: ts.URL,
	}

	queue, err := client.Queue()
	require.NoError(t, err)

	jsonEqual(t, expected, queue)
}

func TestQueueErrors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError string
	}{
		{
			name:          "unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: "unable to get Marathon queue",
		},
		{
			name:          "forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: "unable to get Marathon queue",
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(test.statusCode)
		}))

		client := Client{
			baseURL: ts.URL,
		}

		queue, err := client.Queue()

		assert.EqualError(t, err, test.expectedError)
		assert.EqualValues(t, goMarathon.Queue{}, queue)

		ts.Close()
	}
}

func TestDeployments(t *testing.T) {
	expected := []goMarathon.Deployment{
		{
			ID: "test-id",
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/deployments", r.URL.String())
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := Client{
		baseURL: ts.URL,
	}

	deployments, err := client.Deployments()
	require.NoError(t, err)

	jsonEqual(t, expected, deployments)
}

func TestDeploymentErrors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError string
	}{
		{
			name:          "unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: "unable to get Marathon deployments",
		},
		{
			name:          "forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: "unable to get Marathon deployments",
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(test.statusCode)
		}))

		client := Client{
			baseURL: ts.URL,
		}

		deployments, err := client.Deployments()

		assert.EqualError(t, err, test.expectedError)
		assert.Nil(t, deployments)

		ts.Close()
	}
}

func TestApplicationVersions(t *testing.T) {
	expected := goMarathon.ApplicationVersions{
		Versions: []string{
			"2019-11-18T22:48:41.138Z",
			"2019-11-18T22:27:28.504Z",
			"2019-11-19T07:12:44.466Z",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/apps/kafka/versions", r.URL.String())
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := Client{
		baseURL: ts.URL,
	}

	versions, err := client.ApplicationVersions("/kafka")
	require.NoError(t, err)

	jsonEqual(t, expected, versions)
}

func TestApplicationVersionsErrors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError string
	}{
		{
			name:          "unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: "unable to get versions for Marathon app /kafka",
		},
		{
			name:          "forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: "unable to get versions for Marathon app /kafka",
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(test.statusCode)
		}))

		client := Client{
			baseURL: ts.URL,
		}

		versions, err := client.ApplicationVersions("/kafka")

		assert.EqualError(t, err, test.expectedError)
		assert.Empty(t, versions.Versions)

		ts.Close()
	}
}

func TestApplicationByVersionWithNoVersion(t *testing.T) {
	expected := goMarathon.Application{
		ID: "/kafka",
	}
	response := struct {
		App goMarathon.Application `json:"app"`
	}{
		App: expected,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/apps/kafka", r.URL.String())
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	client := Client{
		baseURL: ts.URL,
	}
	application, err := client.ApplicationByVersion("/kafka", "")

	require.NoError(t, err)
	jsonEqual(t, expected, application)
}

func TestApplicationByVersionWithVersion(t *testing.T) {
	expected := goMarathon.Application{
		ID: "/kafka",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/apps/kafka/versions/2019-11-19T07:12:44.466Z", r.URL.String())
		assert.Equal(t, "GET", r.Method)

		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := Client{
		baseURL: ts.URL,
	}
	application, err := client.ApplicationByVersion("/kafka", "2019-11-19T07:12:44.466Z")

	require.NoError(t, err)
	jsonEqual(t, expected, application)
}

func TestApplicationByVersionErrors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError string
	}{
		{
			name:          "unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: "unable to get version 2019-11-19T07:12:44.466Z for app /kafka",
		},
		{
			name:          "forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: "unable to get version 2019-11-19T07:12:44.466Z for app /kafka",
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(test.statusCode)
		}))

		client := Client{
			baseURL: ts.URL,
		}

		application, err := client.ApplicationByVersion("/kafka", "2019-11-19T07:12:44.466Z")

		assert.EqualError(t, err, test.expectedError)
		assert.Nil(t, application)

		ts.Close()
	}
}
func jsonEqual(t *testing.T, expected interface{}, actual interface{}) {
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)
	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}

func TestAddAppWithEmptyInput(t *testing.T) {
	client := &Client{API: &marathonmocks.MarathonMock{}}
	ctx := mock.NewContext(&cli.Environment{Input: strings.NewReader("")})

	newApp, err := client.AddApp(ctx, "")
	assert.Error(t, err, nil)
	_, ok := err.(*json.SyntaxError)
	assert.True(t, ok)
	assert.Nil(t, newApp)
}

func TestAddAppWithNonExistingFile(t *testing.T) {
	client := &Client{API: &marathonmocks.MarathonMock{}}
	ctx := mock.NewContext(nil)

	newApp, err := client.AddApp(ctx, "/this/does/not/exist")
	assert.Equal(t, ErrCannotReadAppDefinition, err, nil)
	assert.Nil(t, newApp)
}

func TestAddAppWithUnsupportedScheme(t *testing.T) {
	client := &Client{API: &marathonmocks.MarathonMock{}}
	ctx := mock.NewContext(nil)

	newApp, err := client.AddApp(ctx, "ftp://example.org/whateva")
	assert.Error(t, err, nil)
	assert.Nil(t, newApp)
}

func TestAddAppWithHTTPURL(t *testing.T) {
	// spin up HTTP server that serves an app definition
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"some id", "another_property":"another value"}`))
	}))
	defer ts.Close()

	marathonMock := &marathonmocks.MarathonMock{}
	client := &Client{API: marathonMock}
	marathonMock.ApplicationByFn = func(string, *goMarathon.GetAppOpts) (*goMarathon.Application, error) {
		return nil, nil
	}
	marathonMock.ApiPostFn = func(path string, data interface{}, result interface{}) error {
		if app, ok := data.(map[string]interface{}); ok {
			assert.Equal(t, "some id", app["id"])
			assert.Equal(t, "another value", app["another_property"])
			return nil
		}
		assert.Fail(t, "wrong type provided to ApiPost", "%#v", data)
		return nil
	}
	ctx := mock.NewContext(nil)

	newApp, err := client.AddApp(ctx, ts.URL)

	assert.NoError(t, err, nil)
	assert.NotNil(t, newApp)
	assert.Equal(t, 1, marathonMock.ApplicationByInvocations, "Expected ApplicationBy to be invoked once")
	assert.Equal(t, 1, marathonMock.ApiPostInvocations, "Expected ApiPost to be invoked once")
	assert.Equal(t, &goMarathon.Application{}, newApp)
}

func TestAddAppWithBrokenURL(t *testing.T) {
	client := &Client{API: &marathonmocks.MarathonMock{}}
	ctx := mock.NewContext(nil)

	newApp, err := client.AddApp(ctx, "ft p://example.org/whateva")
	assert.Error(t, err, nil)
	assert.Nil(t, newApp)
}

func TestAddApp(t *testing.T) {
	tests := []struct {
		name             string
		applicationByErr error
		applicationByApp *goMarathon.Application
		expectedApp      *goMarathon.Application
		expectedErr      error
	}{
		{
			name:             "application does not exists, yet",
			applicationByErr: &goMarathon.APIError{ErrCode: goMarathon.ErrCodeNotFound},
			applicationByApp: nil,
			expectedApp:      &goMarathon.Application{ID: "some id"},
			expectedErr:      nil,
		},
		{
			name:             "arbitrary APIError",
			applicationByErr: &goMarathon.APIError{ErrCode: 999},
			applicationByApp: nil,
			expectedApp:      nil,
			expectedErr:      &goMarathon.APIError{ErrCode: 999},
		},
		{
			name:             "arbitrary error",
			applicationByErr: goMarathon.ErrMarathonDown,
			applicationByApp: nil,
			expectedApp:      nil,
			expectedErr:      goMarathon.ErrMarathonDown,
		},
		{
			name:             "application already exists",
			applicationByErr: nil,
			applicationByApp: &goMarathon.Application{},
			expectedApp:      nil,
			expectedErr:      ErrAppAlreadyExists{appID: ""},
		},
	}
	marathonMock := marathonmocks.MarathonMock{}
	client := &Client{API: &marathonMock}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marathonMock.ApplicationByFn = func(string, *goMarathon.GetAppOpts) (*goMarathon.Application, error) {
				return tt.applicationByApp, tt.applicationByErr
			}
			marathonMock.ApiPostFn = func(path string, data interface{}, result interface{}) error {
				resApp, ok := result.(*goMarathon.Application)
				if !ok {
					return fmt.Errorf("Parameter result not of type Application")
				}
				inApp, ok := data.(map[string]interface{})
				if !ok {
					return fmt.Errorf("Parameter data not of type map[string]interface{}")
				}
				resApp.ID = inApp["id"].(string)
				return nil
			}

			ctx := mock.NewContext(&cli.Environment{Input: strings.NewReader(`{"id":"some id"}`)})

			newApp, err := client.AddApp(ctx, "")

			assert.Equal(t, tt.expectedApp, newApp)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
