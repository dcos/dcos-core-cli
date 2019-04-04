package logs

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintComponent(t *testing.T) {
	infoEntry := &Entry{
		RealtimeTimestamp: 1550515267000000,
		Fields: EntryFields{
			Priority: "6",
			Message:  "info message",
		},
	}

	errorEntry := &Entry{
		RealtimeTimestamp: 1550515267000000,
		Fields: EntryFields{
			Priority: "3",
			Message:  "error message",
		},
	}

	fixtures := []struct {
		route          string
		service        string
		skip           int
		filters        []string
		format         string
		entries        []*Entry
		colored        bool
		expectedPath   string
		expectedOutput string
	}{
		{
			route:          "/leader/mesos",
			skip:           -10,
			entries:        []*Entry{infoEntry},
			expectedPath:   "/system/v1/leader/mesos/logs/v2/component?skip=-10",
			expectedOutput: "2019-02-18 18:41:07 UTC: info message\n",
		},
		{
			route:          "/agent/fe406283-4198-4aa4-ad77-3f2a034884ee-S1",
			skip:           5,
			entries:        []*Entry{infoEntry},
			expectedPath:   "/system/v1/agent/fe406283-4198-4aa4-ad77-3f2a034884ee-S1/logs/v2/component?skip=5",
			expectedOutput: "2019-02-18 18:41:07 UTC: info message\n",
		},
		{
			route:          "/leader/mesos",
			entries:        []*Entry{infoEntry},
			expectedPath:   "/system/v1/leader/mesos/logs/v2/component?skip=0",
			expectedOutput: "2019-02-18 18:41:07 UTC: info message\n",
		},
		{
			route:          "/leader/mesos",
			entries:        []*Entry{infoEntry},
			colored:        true,
			expectedPath:   "/system/v1/leader/mesos/logs/v2/component?skip=0",
			expectedOutput: "\x1b[0;0m2019-02-18 18:41:07 UTC: info message\x1b[0m\n",
		},
		{
			route:          "/leader/mesos",
			entries:        []*Entry{infoEntry},
			format:         "cat",
			colored:        true,
			expectedPath:   "/system/v1/leader/mesos/logs/v2/component?skip=0",
			expectedOutput: "\x1b[0;0minfo message\x1b[0m\n",
		},
		{
			route:        "/leader/mesos",
			entries:      []*Entry{infoEntry},
			format:       "json-pretty",
			colored:      true,
			expectedPath: "/system/v1/leader/mesos/logs/v2/component?skip=0",
		},
		{
			route:        "/leader/mesos",
			entries:      []*Entry{infoEntry},
			format:       "json",
			colored:      true,
			expectedPath: "/system/v1/leader/mesos/logs/v2/component?skip=0",
		},
		{
			route:          "/leader/mesos",
			entries:        []*Entry{errorEntry},
			colored:        true,
			expectedPath:   "/system/v1/leader/mesos/logs/v2/component?skip=0",
			expectedOutput: "\x1b[0;31m2019-02-18 18:41:07 UTC: error message\x1b[0m\n",
		},
	}

	for _, fixture := range fixtures {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, fixture.expectedPath, r.URL.String())

			for _, entry := range fixture.entries {
				assert.NoError(t, json.NewEncoder(w).Encode(entry))
			}
		}))

		var b bytes.Buffer
		c := NewClient(pluginutil.HTTPClient(ts.URL), &b)
		c.colored = fixture.colored

		opts := Options{
			Filters: fixture.filters,
			Format:  fixture.format,
			Skip:    fixture.skip,
		}

		err := c.PrintComponent(fixture.route, fixture.service, opts)
		require.NoError(t, err)

		switch fixture.format {
		case "json", "json-pretty":
			var entry JournalctlJSONEntry
			require.NoError(t, json.Unmarshal([]byte(b.String()), &entry))
			require.Equal(t, entry, infoEntry.JournalctlJSON())
		default:
			require.Equal(t, fixture.expectedOutput, b.String())
		}

		ts.Close()
	}
}
