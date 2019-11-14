package marathon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gambol99/go-marathon"
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
	expected := marathon.Applications{
		Apps: []marathon.Application{
			{
				ID:      "test-id",
				Env:     new(map[string]string),
				Secrets: new(map[string]marathon.Secret),
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
	expected := marathon.Queue{
		Items: []marathon.Item{
			{
				Count: 1,
				Delay: marathon.Delay{
					Overdue:         false,
					TimeLeftSeconds: 5,
				},
				Application: marathon.Application{
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
		assert.EqualValues(t, marathon.Queue{}, queue)

		ts.Close()
	}
}

func TestDeployments(t *testing.T) {
	expected := []marathon.Deployment{
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

func jsonEqual(t *testing.T, expected interface{}, actual interface{}) {
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)
	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}
