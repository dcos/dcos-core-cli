package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dcos/dcos-core-cli/pkg/marathon"
	marathonmocks "github.com/dcos/dcos-core-cli/pkg/marathon/mocks"

	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-cli/pkg/mock"
	goMarathon "github.com/gambol99/go-marathon"
)

func TestAppAdd(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		location string
		err      bool
		errOut   string
		out      string
	}{
		{
			name:     "empty input",
			input:    "",
			location: "",
			err:      true,
			errOut:   "error loading JSON: unexpected end of JSON input",
			out:      "",
		},
		{
			name:     "input from absent file",
			location: "/this/file/does/not/exist",
			err:      true,
			errOut:   "can't read from resource: /this/file/does/not/exist. Please check that it exists",
			out:      "",
		},
		{
			name:     "valid input",
			input:    `{"id":"the id"}`,
			location: "",
			err:      false,
			errOut:   "",
			out:      "Created deployment some id\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marathonMock := marathonmocks.MarathonMock{}
			marathonMock.ApplicationByFn = func(id string, opts *goMarathon.GetAppOpts) (*goMarathon.Application, error) {
				return nil, nil
			}
			marathonMock.ApiPostFn = func(path string, app interface{}, res interface{}) error {
				if resApp, ok := res.(*goMarathon.Application); ok {
					resApp.Deployments = []map[string]string{{"id": "some id"}}
					return nil
				}
				return fmt.Errorf("res not of type Application")
			}
			var output strings.Builder
			ctx := mock.NewContext(&cli.Environment{Input: strings.NewReader(tt.input), Out: &output})
			client := marathon.Client{API: &marathonMock}
			err := marathonAppAdd(ctx, client, tt.location)

			if tt.err {
				assert.Error(t, err)
				assert.Equal(t, tt.errOut, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.out, output.String())
		})
	}
}
