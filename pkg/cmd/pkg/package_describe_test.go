package pkg

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pkgDescribe(t *testing.T) {
	var out bytes.Buffer
	env := mock.NewEnvironment()
	env.Fs = afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)
	env.Out = &out

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method == http.MethodPost && r.URL.Path == "/package/describe" {
			path := filepath.Join("testdata", "cosmos_chronos_describe_response.json")
			bytes, err := ioutil.ReadFile(path)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/vnd.dcos.service.describe-request+json;charset=utf-8;version=v1")
			w.Write(bytes)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.CloseClientConnections()

	oldDcosUrl := os.Getenv("DCOS_URL")
	err := os.Setenv("DCOS_URL", testServer.URL)
	require.NoError(t, err)
	defer func() {
		err := os.Setenv("DCOS_URL", oldDcosUrl)
		require.NoError(t, err)
	}()

	ctx := mock.NewContext(env)

	var testCases = []struct {
		in  describeOptions
		out string
	}{
		{describeOptions{cliOnly: true}, "cosmos_chronos_describe_cli.expected.txt"},
		{describeOptions{appOnly: true}, "cosmos_chronos_describe_app.expected.txt"},
		{describeOptions{config: true}, "cosmos_chronos_describe_config.expected.json"},
	}
	for _, tt := range testCases {
		t.Run(tt.out, func(t *testing.T) {
			var out bytes.Buffer
			env.Out = &out

			path := filepath.Join("testdata", tt.out)
			expected, err := ioutil.ReadFile(path)
			assert.NoError(t, err)

			err = pkgDescribe(ctx, "helloworld", tt.in)
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(out.Bytes()))
		})
	}
}
