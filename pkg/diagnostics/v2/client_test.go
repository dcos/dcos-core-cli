package v2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/dcos/dcos-cli/pkg/httpclient"

	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	response := []Bundle{
		{
			ID:      "test",
			Size:    100,
			Status:  Done,
			Started: time.Now().UTC(),
			Stopped: time.Now().UTC(),
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/diagnostics", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(response))
	}))

	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	bundles, err := c.List()
	require.NoError(t, err)

	assert.EqualValues(t, response, bundles)
}

func TestEmptyList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/diagnostics", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode([]Bundle{}))
	}))

	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	bundles, err := c.List()
	require.NoError(t, err)

	assert.Empty(t, bundles)
}

func TestCreateHappyPath(t *testing.T) {
	var newID string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		re := regexp.MustCompile("^/system/health/v1/diagnostics/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})$")
		urlMatch := re.FindStringSubmatch(r.URL.Path)
		newID = urlMatch[1]
		require.True(t, len(urlMatch) == 2, fmt.Sprintf("URL %s doesn't match %s", r.URL.Path, re.String()))
		require.Equal(t, "PUT", r.Method)
		fmt.Fprint(w, fmt.Sprintf(`{"id":"%s","status":"Started","started_at":"2019-08-19T08:07:24.404239211Z","stopped_at":"0001-01-01T00:00:00Z"}`, newID))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	id, err := c.Create()
	require.NoError(t, err)

	assert.Equal(t, newID, id)
}

func TestCreateServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	id, err := c.Create()
	require.Error(t, err)
	require.Empty(t, id)

	httpError, ok := err.(*httpclient.HTTPError)
	require.True(t, ok, "unexpected error: %#v", err)
	require.NotNil(t, httpError.Response)
	require.Equal(t, httpError.Response.StatusCode, 500)
	require.Equal(t, "HTTP 500 error", httpError.Error())
}

func TestDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/diagnostics/bundle-0/file", r.URL.String())
		assert.Equal(t, "GET", r.Method)

		http.ServeFile(w, r, "testdata/test_bundle.zip")
	}))
	defer ts.Close()

	file, err := ioutil.TempFile(os.TempDir(), "bundle-0.zip")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	err = c.Download("bundle-0", file)
	require.NoError(t, err)
	defer file.Close()

	outputData, err := ioutil.ReadFile(file.Name())
	require.NoError(t, err)

	testData, err := ioutil.ReadFile("testdata/test_bundle.zip")
	require.NoError(t, err)

	assert.EqualValues(t, testData, outputData)
}

func TestDownloadBundleErrors(t *testing.T) {
	type testDef struct {
		name       string
		returnCode int
		errMessage string
	}

	tests := []testDef{
		{
			name:       "not found",
			returnCode: http.StatusNotFound,
			errMessage: "no bundle bundle-0 found",
		},
		{
			name:       "internal server error",
			returnCode: http.StatusInternalServerError,
			errMessage: "bundle bundle-0 not readable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/system/health/v1/diagnostics/bundle-0/file", r.URL.String())
				assert.Equal(t, "GET", r.Method)

				w.WriteHeader(test.returnCode)
			}))
			defer ts.Close()

			file, err := ioutil.TempFile(os.TempDir(), "bundle-0.zip")
			require.NoError(t, err)
			defer os.Remove(file.Name())

			c := NewClient(pluginutil.HTTPClient(ts.URL))
			err = c.Download("bundle-0", file)
			assert.Error(t, err)
			assert.EqualError(t, err, test.errMessage)

			file.Close()

			outputData, err := ioutil.ReadFile(file.Name())
			require.NoError(t, err)

			assert.Empty(t, outputData)
		})
	}
}
