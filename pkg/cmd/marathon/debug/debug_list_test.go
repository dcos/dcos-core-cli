package debug

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	goMarathon "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugList(t *testing.T) {
	response := goMarathon.Queue{
		Items: []goMarathon.Item{
			{
				Application: &goMarathon.Application{
					ID: "/test-app",
				},
				Since: "2019-11-28T14:33:09.156Z",
				Count: 1,
				Delay: goMarathon.Delay{
					Overdue: true,
				},
				ProcessedOffersSummary: goMarathon.ProcessedOffersSummary{
					ProcessedOffersCount: 90,
					UnusedOffersCount:    80,
					LastUnusedOfferAt:    strPointer("2019-11-28T14:33:07.631Z"),
					LastUsedOfferAt:      strPointer("2019-11-28T14:30:07.631Z"),
				},
			},
			{
				Pod: &goMarathon.Pod{
					ID: "/test-pod",
				},
				Since: "2019-11-28T14:31:09.156Z",
				Count: 2,
				Delay: goMarathon.Delay{
					Overdue: false,
				},
				ProcessedOffersSummary: goMarathon.ProcessedOffersSummary{
					ProcessedOffersCount: 100,
					UnusedOffersCount:    90,
					LastUnusedOfferAt:    strPointer("2019-11-28T14:25:07.631Z"),
					LastUsedOfferAt:      strPointer("2019-11-28T14:24:07.631Z"),
				},
			},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/queue", r.URL.String())
		json.NewEncoder(w).Encode(&response)
	}))

	ctx, out := newContext(ts)

	err := debugList(ctx, false)
	require.NoError(t, err)

	expected := "     ID               SINCE            INSTANCES TO LAUNCH  WAITING  PROCESSED OFFERS  " +
		"UNUSED OFFERS     LAST UNUSED OFFER          LAST USED OFFER       \n" +
		"  /test-app  2019-11-28T14:33:09.156Z                    1  true                   90          " +
		"   80  2019-11-28T14:33:07.631Z  2019-11-28T14:30:07.631Z  \n" +
		"  /test-pod  2019-11-28T14:31:09.156Z                    2  false                 100          " +
		"   90  2019-11-28T14:25:07.631Z  2019-11-28T14:24:07.631Z  \n"
	assert.Equal(t, expected, out.String())
}

func strPointer(s string) *string {
	return &s
}
