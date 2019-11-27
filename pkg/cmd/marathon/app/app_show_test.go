package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppShowNoVersion(t *testing.T) {
	expected := struct {
		App marathon.Application `json:"app"`
	}{
		App: marathon.Application{
			ID: "/kafka",
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/v2/apps/kafka":
			json.NewEncoder(w).Encode(expected)
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	ctx, out := newContext(ts)

	err := appShow(ctx, "/kafka", "")
	require.NoError(t, err)

	expectedJSON := new(bytes.Buffer)
	enc := json.NewEncoder(expectedJSON)
	enc.SetIndent("", "    ")
	err = enc.Encode(expected.App)
	require.NoError(t, err)

	assert.Equal(t, expectedJSON.String(), out.String())
}

func TestAppShowRelativeVersion(t *testing.T) {
	expected := marathon.Application{
		ID: "/kafka",
	}
	versions := marathon.ApplicationVersions{
		Versions: []string{
			"2019-11-19T07:12:44.466Z",
			"2019-11-18T07:12:44.466Z",
			"2019-11-17T07:12:44.466Z",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		// nolint:goconst
		case "/service/marathon/v2/apps/kafka/versions":
			json.NewEncoder(w).Encode(versions)
		case "/service/marathon/v2/apps/kafka/versions/2019-11-17T07:12:44.466Z":
			json.NewEncoder(w).Encode(expected)
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	ctx, out := newContext(ts)

	err := appShow(ctx, "/kafka", "-2")
	require.NoError(t, err)

	expectedJSON := new(bytes.Buffer)
	enc := json.NewEncoder(expectedJSON)
	enc.SetIndent("", "    ")
	err = enc.Encode(expected)
	require.NoError(t, err)

	assert.Equal(t, expectedJSON.String(), out.String())
}

func TestAppShowRelativePostiveVersion(t *testing.T) {

	// required to make it possible to set up a client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "no calls to DCOS expected")
	}))

	ctx, _ := newContext(ts)

	err := appShow(ctx, "/kafka", "2")
	assert.EqualError(t, err, "relative versions must be negative: 2")
}

func TestAppShowAbsoluteVersion(t *testing.T) {
	expected := marathon.Application{
		ID: "/kafka",
	}
	versions := marathon.ApplicationVersions{
		Versions: []string{
			"2019-11-19T07:12:44.466Z",
			"2019-11-18T07:12:44.466Z",
			"2019-11-17T07:12:44.466Z",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/v2/apps/kafka/versions":
			json.NewEncoder(w).Encode(versions)
		case "/service/marathon/v2/apps/kafka/versions/2019-11-18T07:12:44.466Z":
			json.NewEncoder(w).Encode(expected)
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	ctx, out := newContext(ts)

	err := appShow(ctx, "/kafka", "2019-11-18T07:12:44.466Z")
	require.NoError(t, err)

	expectedJSON := new(bytes.Buffer)
	enc := json.NewEncoder(expectedJSON)
	enc.SetIndent("", "    ")
	err = enc.Encode(expected)
	require.NoError(t, err)

	assert.Equal(t, expectedJSON.String(), out.String())
}

func TestAppShowNonExistantVersion(t *testing.T) {
	versions := marathon.ApplicationVersions{
		Versions: []string{
			"2019-11-19T07:12:44.466Z",
			"2019-11-18T07:12:44.466Z",
			"2019-11-17T07:12:44.466Z",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/v2/apps/kafka/versions":
			json.NewEncoder(w).Encode(versions)
		case "/service/marathon/v2/apps/kafka/versions/2019-11-15T07:12:44.466Z":
			w.WriteHeader(404)
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	ctx, _ := newContext(ts)

	err := appShow(ctx, "/kafka", "2019-11-15T07:12:44.466Z")
	assert.EqualError(t, err, "app '/kafka' does not exist")
}
