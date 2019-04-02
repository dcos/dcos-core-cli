package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/r3labs/sse"
	"golang.org/x/crypto/ssh/terminal"
)

// Client is a logs client for DC/OS.
type Client struct {
	http    *httpclient.Client
	out     io.Writer
	colored bool
}

// NewClient creates a new logs client.
func NewClient(baseClient *httpclient.Client, out io.Writer) *Client {
	c := &Client{http: baseClient, out: out}

	// Enable colors on UNIX when Out is a terminal.
	if outFile, ok := out.(*os.File); ok {
		if runtime.GOOS != "windows" && terminal.IsTerminal(int(outFile.Fd())) {
			c.colored = true
		}
	}
	return c
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

		events := make(chan *sse.Event)
		err := client.SubscribeChanRaw(events)
		if err != nil {
			return err
		}
		defer client.Unsubscribe(events)

		for msg := range events {
			err := c.printEntry(msg.Data)
			if err != nil {
				return err
			}
		}
		return nil
	}

	resp, err := c.http.Get(endpoint, httpclient.Header("Accept", "application/json"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
	for scanner := bufio.NewScanner(resp.Body); scanner.Scan(); {
		err := c.printEntry(scanner.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) printEntry(rawEntry []byte) error {
	var entry Entry
	err := json.Unmarshal(rawEntry, &entry)
	if err != nil {
		return err
	}

	if c.colored {
		var color string
		switch entry.Fields.Priority {
		// EMERGENCY, ALERT, CRITICAL, ERROR are printed in red.
		case "0", "1", "2", "3":
			color = "31"
		// WARNING is printed in yellow.
		case "4":
			color = "33"
		// NOTICE is printed in bright blue.
		case "5":
			color = "34;1"
		default:
			color = "0"
		}
		fmt.Fprintf(c.out, "\033[0;%sm", color)
	}

	date := time.Unix(entry.RealtimeTimestamp/1000000, 0).UTC().Format("2006-01-02 15:04:05 MST")
	var pid string
	if entry.Fields.PID != "" {
		pid = fmt.Sprintf(" [%s]", entry.Fields.PID)
	}
	fmt.Fprint(
		c.out,
		date,
		entry.Fields.SyslogIdentifier,
		pid,
		": ",
		entry.Fields.Message,
	)
	if c.colored {
		fmt.Fprint(c.out, "\033[0m")
	}
	fmt.Fprintln(c.out)
	return nil
}
