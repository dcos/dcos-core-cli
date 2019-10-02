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

	uuid "github.com/satori/go.uuid"

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

func TestGet(t *testing.T) {
	response := &Bundle{
		ID:      "test",
		Size:    100,
		Status:  Done,
		Started: time.Now().UTC(),
		Stopped: time.Now().UTC(),
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/diagnostics/bundle-id", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
		assert.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	bundle, err := c.Get("bundle-id")
	require.NoError(t, err)

	assert.EqualValues(t, response, bundle)
}

func TestGetErrors(t *testing.T) {
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
			name:       "not modified",
			returnCode: http.StatusNotModified,
			errMessage: "bundle bundle-0 has already been deleted",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/system/health/v1/diagnostics/bundle-0", r.URL.String())
				assert.Equal(t, "GET", r.Method)

				w.WriteHeader(test.returnCode)
			}))
			defer ts.Close()

			c := NewClient(pluginutil.HTTPClient(ts.URL))
			bundle, err := c.Get("bundle-0")
			assert.Error(t, err)
			assert.EqualError(t, err, test.errMessage)

			assert.Nil(t, bundle)
		})
	}
}

func TestDeleteHappyPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/diagnostics/this-is-the-bundle-id", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	err := c.Delete("this-is-the-bundle-id")
	assert.NoError(t, err)
}

func TestDeleteWithEmptyID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/diagnostics/", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	err := c.Delete("")
	assert.Error(t, err)
	httpError, ok := err.(*httpclient.HTTPError)
	require.True(t, ok, "unexpected error: %#v", err)
	require.NotNil(t, httpError.Response)
	assert.Equal(t, httpError.Response.StatusCode, http.StatusMethodNotAllowed)
	assert.Equal(t, "HTTP 405 error", httpError.Error())
}

func TestDeleteWithUnknownID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/diagnostics/never-heard-of-this-id", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	err := c.Delete("never-heard-of-this-id")
	assert.Error(t, err)
	assert.Equal(t, "no bundle never-heard-of-this-id found", err.Error())
}

func TestDeleteWithDeletedBundle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/diagnostics/already-deleted", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNotModified)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	err := c.Delete("already-deleted")
	assert.Error(t, err)
	assert.Equal(t, "bundle already-deleted has already been deleted", err.Error())
}

func TestCreateHappyPath(t *testing.T) {
	var newID uuid.UUID
	re := regexp.MustCompile("^/system/health/v1/diagnostics/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})$")
	handler := func(t *testing.T, expectedBody string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			urlMatch := re.FindStringSubmatch(r.URL.Path)
			idMatch := urlMatch[1]
			var err error
			newID, err = uuid.FromString(idMatch)
			require.NoError(t, err)
			body, err := ioutil.ReadAll(r.Body)

			assert.NoError(t, err)
			assert.Equal(t, "PUT", r.Method)
			assert.NoError(t, err)
			assert.Equal(t, expectedBody, string(body))

			_, err = fmt.Fprint(w, fmt.Sprintf(
				`{
						  "id":"%s",
						  "status":"Started",
						  "started_at":"2019-08-19T08:07:24.404239211Z",
						  "stopped_at":"0001-01-01T00:00:00Z"
						}`, newID))
			assert.NoError(t, err)
		})
	}

	for _, tc := range []struct {
		expected string
		given    Options
	}{
		{expected: `{"masters":false,"agents":false}`, given: Options{}},
		{expected: `{"masters":true,"agents":false}`, given: Options{Masters: true}},
		{expected: `{"masters":false,"agents":true}`, given: Options{Agents: true}},
		{expected: `{"masters":true,"agents":true}`, given: Options{Masters: true, Agents: true}},
	} {
		t.Run(tc.expected, func(t *testing.T) {
			ts := httptest.NewServer(handler(t, tc.expected))
			defer ts.Close()
			c := NewClient(pluginutil.HTTPClient(ts.URL))
			id, err := c.Create(tc.given)
			assert.NoError(t, err)
			assert.Equal(t, newID.String(), id)
		})
	}
}

func TestCreateServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	id, err := c.Create(Options{})
	assert.Error(t, err)
	assert.Empty(t, id)

	httpError, ok := err.(*httpclient.HTTPError)
	require.True(t, ok, "unexpected error: %#v", err)
	require.NotNil(t, httpError.Response)
	assert.Equal(t, httpError.Response.StatusCode, 500)
	assert.Equal(t, "HTTP 500 error", httpError.Error())
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
