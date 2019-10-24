package diagnostics

import (
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateErrorsWhenGot503WithStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/report/diagnostics/create", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := w.Write([]byte(`{
			"response_http_code":503,
			"version":1,
			"status":"requested nodes: [] not found",
			"errors":null,
			"extra":{"bundle_name":""}
		}`))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	got, err := c.Create([]string{""})

	assert.Empty(t, got)
	assert.EqualError(t, err, "unexpected status code 503 Service Unavailable: requested nodes: [] not found")
}

func TestClient_CreateErrorsWhenGot500WithoutStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/report/diagnostics/create", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`500 Internal Server Error`))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	got, err := c.Create([]string{"all"})

	assert.Empty(t, got)
	assert.EqualError(t, err, "HTTP 500 error")
}

func TestClient_Create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/health/v1/report/diagnostics/create", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{
			"response_http_code":200,
			"version":1,
			"status":"Job has been successfully started",
			"errors":null,
			"extra":{
				"bundle_name":"bundle-2019-10-24-1571919252.zip"
			}
		}`))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))
	got, err := c.Create([]string{"all"})

	expected := BundleCreateResponseJSONStruct{
		Status: "Job has been successfully started",
		Extra: struct {BundleName string `json:"bundle_name"`}{"bundle-2019-10-24-1571919252.zip",},
	}
	assert.NoError(t, err)
	assert.Equal(t, &expected, got)
}
