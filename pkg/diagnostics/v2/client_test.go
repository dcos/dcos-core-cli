package v2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
