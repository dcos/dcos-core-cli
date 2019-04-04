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

// Options encapsulates options that can be set via flags on the command.
type Options struct {
	Filters []string
	Follow  bool
	Format  string
	Skip    int
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
func (c *Client) PrintComponent(route string, service string, opts Options) error {
	requestFilters := ""
	if len(opts.Filters) > 0 {
		requestFilters = "&filter=" + strings.Join(opts.Filters, "&filter=")
	}
	endpoint := fmt.Sprintf("/system/v1%s/logs/v2/component%s?skip=%d%s", route, service, opts.Skip, requestFilters)
	if opts.Follow {
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
			err := c.printEntry(msg.Data, opts)
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
		err := c.printEntry(scanner.Bytes(), opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) printEntry(rawEntry []byte, opts Options) error {
	var entry Entry
	err := json.Unmarshal(rawEntry, &entry)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(c.out)
	switch opts.Format {
	case "json-pretty":
		enc.SetIndent("", "    ")
		fallthrough
	case "json":
		return enc.Encode(entry.JournalctlJSON())
	case "cat":
		c.setColor(entry.Fields.Priority)
		fmt.Fprint(c.out, entry.Fields.Message)
		c.resetColor()
	default:
		c.setColor(entry.Fields.Priority)
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
		c.resetColor()
	}

	fmt.Fprintln(c.out)
	return nil
}

func (c *Client) setColor(priority string) {
	if c.colored {
		var color string
		switch priority {
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
}

func (c *Client) resetColor() {
	if c.colored {
		fmt.Fprint(c.out, "\033[0m")
	}
}
