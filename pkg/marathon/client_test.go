package marathon

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseToError(t *testing.T) {
	resp := http.Response{StatusCode: 200}
	err := httpResponseToError(&resp)
	assert.EqualError(t, err, `unexpected status code 200`)

	resp = http.Response{StatusCode: 404}
	err = httpResponseToError(&resp)
	assert.EqualError(t, err, `HTTP 404 error`)
}
