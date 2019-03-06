package networking

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

func TestNodes(t *testing.T) {
	expectedNodes := []Node{
		Node{
			Updated:   time.Date(1996, 3, 6, 20, 34, 58, 651387237, time.UTC),
			PublicIPs: []string{"192.168.0.1", "192.168.1.1"},
			PrivateIP: "1.1.1.1",
			Hostname:  "localhost",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/net/v1/nodes", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(expectedNodes))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	nodes, err := c.Nodes()
	require.NoError(t, err)
	require.Equal(t, expectedNodes, nodes)
}
