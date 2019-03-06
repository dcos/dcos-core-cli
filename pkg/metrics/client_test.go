package metrics

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

func TestNode(t *testing.T) {
	var dimensions struct {
		MesosID   string `json:"mesos_id"`
		ClusterID string `json:"cluster_id"`
		Hostname  string `json:"hostname"`
	}
	expectedNode := Node{
		Datapoints: []Datapoint{
			Datapoint{
				Name:      "datapoint",
				Value:     0,
				Unit:      "percentage",
				Timestamp: time.Date(1996, 3, 6, 20, 34, 58, 651387237, time.UTC),
				Tags:      map[string]string{"hello": "world"},
			},
		},
		Dimensions: dimensions,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/system/v1/agent/mesosID/metrics/v0/node", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedNode))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	node, err := c.Node("mesosID")
	require.NoError(t, err)
	require.Equal(t, &expectedNode, node)
}
