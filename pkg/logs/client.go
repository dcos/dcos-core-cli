package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/r3labs/sse"
)

// Client is a logs client for DC/OS.
type Client struct {
	http *httpclient.Client
	out  io.Writer
}

// NewClient creates a new logs client.
func NewClient(baseClient *httpclient.Client, out io.Writer) *Client {
	return &Client{
		http: baseClient,
		out:  out,
	}
}

// PrintComponent prints a component logs.
func (c *Client) PrintComponent(route string, service string, skip int, filters []string, follow bool) error {
	requestFilters := ""
	if len(filters) > 0 {
		requestFilters = "&filter=" + strings.Join(filters, "&filter=")
	}
	endpoint := fmt.Sprintf("/system/v1%s/logs/v2/component%s?skip=%d%s", route, service, skip, requestFilters)
	if follow {
		client := sse.NewClient(c.http.BaseURL().String() + endpoint)
		client.Connection = c.http.BaseClient()
		client.Headers["Authorization"] = c.http.Header().Get("Authorization")
		client.Headers["User-Agent"] = c.http.Header().Get("User-Agent")

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		events := make(chan *sse.Event)
		err := client.SubscribeChanRaw(events)
		if err != nil {
			return err
		}
		var data SSEEventDataField

		for {
			select {
			case msg := <-events:
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					client.Unsubscribe(events)
					return err
				}
				date := time.Unix(data.RealtimeTimestamp/int64(math.Pow(10, 6)), 0)
				fmt.Fprintln(c.out, date.Format("2006-01-02 15:04:05 MST")+": "+data.Fields.Message)
			case <-sig:
				fmt.Fprintln(c.out, "User interrupted command with Ctrl-C")
				client.Unsubscribe(events)
				return nil
			}
		}
	}
	resp, err := c.http.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Fprintln(c.out, strings.TrimSpace(string(bodyBytes)))
		return nil
	}
	return fmt.Errorf("HTTP %d error", resp.StatusCode)
}
