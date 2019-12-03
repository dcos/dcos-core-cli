package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/stretchr/testify/assert"
)

func TestMarathonAppVersionList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test/versions", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.Write([]byte(`{"versions": ["2019-11-21T16:15:35.114Z","2019-11-21T16:49:06.988Z"]}`))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	versions, err := marathonAppVersionList(*client, "app-test", 0)
	assert.NoError(t, err)
	assert.Equal(t, []string{"2019-11-21T16:15:35.114Z", "2019-11-21T16:49:06.988Z"}, versions)
}

func TestMarathonAppVersionListMaxCount(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test/versions", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.Write([]byte(`{"versions": ["2019-11-21T16:15:35.114Z","2019-11-21T16:49:06.988Z"]}`))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	versions, err := marathonAppVersionList(*client, "app-test", 1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"2019-11-21T16:15:35.114Z"}, versions)
}

func TestMarathonAppVersionListMaxCountTooBig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test/versions", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.Write([]byte(`{"versions": ["2019-11-21T16:15:35.114Z","2019-11-21T16:49:06.988Z"]}`))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	versions, err := marathonAppVersionList(*client, "app-test", 3)
	assert.NoError(t, err)
	assert.Equal(t, []string{"2019-11-21T16:15:35.114Z", "2019-11-21T16:49:06.988Z"}, versions)
}

func TestMarathonAppVersionListError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test/versions", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusExpectationFailed)
		w.Write([]byte("Something bad happened"))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	_, err = marathonAppVersionList(*client, "app-test", 3)
	assert.EqualError(t, err, "Marathon API error: Something bad happened")
}
