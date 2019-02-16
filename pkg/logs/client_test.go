package logs

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/stretchr/testify/require"
)

func TestPrintComponentLeaderNoService(t *testing.T) {
	correctURL := "/system/v1/leader/mesos/logs/v2/component?skip=-10"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.String())
	}))
	defer ts.Close()

	var b bytes.Buffer
	c := NewClient(pluginutil.HTTPClient(ts.URL), &b)
	err := c.PrintComponent("/leader/mesos", "", -10, []string{}, false)
	require.Equal(t, nil, err)
	require.Equal(t, correctURL, strings.TrimSpace(b.String()))
}
