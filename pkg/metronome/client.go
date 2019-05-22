package metronome

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/sirupsen/logrus"
)

// Client is a client for Cosmos.
type Client struct {
	http   *httpclient.Client
	logger *logrus.Logger
}

// JobsOption is a functional Option to set the `embed` query parameters.
type JobsOption func(query url.Values)

// EmbedActiveRun sets the `embed` option to activeRuns.
func EmbedActiveRun() JobsOption {
	return func(query url.Values) {
		query.Add("embed", "activeRuns")
	}
}

// EmbedSchedule sets the `embed` option to schedules.
func EmbedSchedule() JobsOption {
	return func(query url.Values) {
		query.Add("embed", "schedules")
	}
}

// EmbedHistory sets the `embed` option to history.
func EmbedHistory() JobsOption {
	return func(query url.Values) {
		query.Add("embed", "history")
	}
}

// EmbedHistorySummary sets the `embed` option to historySummary.
func EmbedHistorySummary() JobsOption {
	return func(query url.Values) {
		query.Add("embed", "historySummary")
	}
}

// NewClient creates a new Metronome client.
func NewClient(baseClient *httpclient.Client, logger *logrus.Logger) *Client {
	return &Client{
		http:   baseClient,
		logger: logger,
	}
}

// Job returns a Job for the given jobID.
func (c *Client) Job(jobID string, opts ...JobsOption) (*Job, error) {

	req, err := c.http.NewRequest("GET", "/v1/jobs/"+jobID, nil)
	if err != nil {
		return nil, err
	}

	// Add embed query parameters to the request URL.
	q := req.URL.Query()
	for _, opt := range opts {
		opt(q)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var job Job
		if err = json.NewDecoder(resp.Body).Decode(&job); err != nil {
			return nil, err
		}
		return &job, nil
	case 404:
		return nil, fmt.Errorf(`job "%s" does not exist`, jobID)
	default:
		return httpResponseToError(resp)
	}
}

// Jobs returns a list of all job definitions.
func (c *Client) Jobs(opts ...JobsOption) ([]Job, error) {

	req, err := c.http.NewRequest("GET", "/v1/jobs", nil)
	if err != nil {
		return nil, err
	}

	// Add embed query parameters to the request URL.
	q := req.URL.Query()
	for _, opt := range opts {
		opt(q)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var jobs []Job
		err = json.NewDecoder(resp.Body).Decode(&jobs)
		return jobs, err
	default:
		return httpResponseToError(resp)
	}
}

func (c *Client) addOrUpdateJob(job *Job, add bool) (*Job, error) {
	jsonBytes, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	buf := bytes.NewBuffer(jsonBytes)
	if add {
		req, err = c.http.NewRequest("POST", "/v1/jobs", buf)
	} else {
		req, err = c.http.NewRequest("PUT", "/v1/jobs/"+job.ID, buf)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 201:
		var j Job
		if err = json.NewDecoder(resp.Body).Decode(&j); err != nil {
			return nil, err
		}
		return &j, nil
	default:
		return httpResponseToError(resp)
	}
}

// AddJob creates a job.
func (c *Client) AddJob(job *Job) (*Job, error) {
	return c.addOrUpdateJob(job, true)
}

// UpdateJob updates an existing job.
func (c *Client) UpdateJob(job *Job) (*Job, error) {
	return c.addOrUpdateJob(job, false)
}

// RunJob triggers a run of the job with a given jobID right now.
func (c *Client) RunJob(jobID string) (*Run, error) {
	resp, err := c.http.Post("/v1/jobs/"+jobID+"/runs", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 201:
		var run Run
		if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
			return nil, err
		}
		return &run, nil
	case 404:
		return nil, fmt.Errorf(`job "%s" does not exist`, jobID)
	default:
		return httpResponseToError(resp)
	}
}

// Run returns the run object for a given jobID and runID.
func (c *Client) Run(jobID, runID string) (*Run, error) {
	resp, err := c.http.Get("/v1/jobs/" + jobID + "/runs/" + runID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var run Run
		if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
			return nil, err
		}
		return &run, nil
	case 404:
		return nil, fmt.Errorf(`job "%s" does not exist`, jobID)
	default:
		return httpResponseToError(resp)
	}
}

// Runs returns the run objects for a given jobID.
func (c *Client) Runs(jobID string) ([]Run, error) {
	resp, err := c.http.Get("/v1/jobs/" + jobID + "/runs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var runs []Run
		if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
			return nil, err
		}
		return runs, nil
	case 404:
		return nil, fmt.Errorf(`job "%s" does not exist`, jobID)
	default:
		return httpResponseToError(resp)
	}
}

// Kill stops a run of a given jobID and runID.
func (c *Client) Kill(jobID, runID string) error {
	resp, err := c.http.Post("/v1/jobs/"+jobID+"/runs/"+runID+"/actions/stop", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		c.logger.Infof(`Run "%s" of job "%s" killed.`, runID, jobID)
		return nil
	case 404:
		return fmt.Errorf(`job "%s" or run "%s" does not exist`, jobID, runID)
	default:
		return httpResponseToError(resp)
	}
}

// RemoveJob removes a job.
func (c *Client) RemoveJob(jobID string, force bool) error {
	req, err := c.http.NewRequest("DELETE", "/v1/jobs/"+jobID, nil)
	if err != nil {
		return err
	}

	// Add embed query parameters to the request URL.
	q := req.URL.Query()
	if force {
		q.Add("stopCurrentJobRuns", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf(`job "%s" does not exist`, jobID)
	case 409:
		return fmt.Errorf(`job "%s" is running`, jobID)
	default:
		return httpResponseToError(resp)
	}
}

// Schedules returns the schedules for a given jobID.
func (c *Client) Schedules(jobID string) ([]Schedule, error) {
	resp, err := c.http.Get("/v1/jobs/" + jobID + "/schedules")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var schedules []Schedule
		err = json.NewDecoder(resp.Body).Decode(&schedules)
		if err != nil {
			return nil, err
		}
		return schedules, nil
	case 404:
		return nil, fmt.Errorf(`job "%s" does not exist`, jobID)
	default:
		return httpResponseToError(resp)
	}
}

func (c *Client) addOrUpdateSchedule(jobID string, schedule *Schedule, add bool) (*Schedule, error) {
	jsonBytes, err := json.Marshal(schedule)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	buf := bytes.NewBuffer(jsonBytes)

	if add {
		req, err = c.http.NewRequest("POST", "/v1/jobs/"+jobID+"/schedules", buf)
	} else {
		req, err = c.http.NewRequest("PUT", "/v1/jobs/"+jobID+"/schedules/"+schedule.ID, buf)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 201:
		var s Schedule
		if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
			return nil, err
		}
		return &s, nil
	default:
		return httpResponseToError(resp)
	}
}

// AddSchedule adds a schedule to the job with the given jobID.
func (c *Client) AddSchedule(jobID string, schedule *Schedule) (*Schedule, error) {
	return c.addOrUpdateSchedule(jobID, schedule, true)
}

// UpdateSchedule updates a schedule of a job with the given jobID.
func (c *Client) UpdateSchedule(jobID string, schedule *Schedule) (*Schedule, error) {
	return c.addOrUpdateSchedule(jobID, schedule, false)
}

// RemoveSchedule removes a schedule from a job with the given jobID and scheduleID.
func (c *Client) RemoveSchedule(jobID, scheduleID string) error {
	resp, err := c.http.Delete("/v1/jobs/" + jobID + "/schedules/" + scheduleID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf(`job "%s" or schedule "%s" does not exist`, jobID, scheduleID)
	default:
		return httpResponseToError(resp)
	}
}

// Queued returns all queued runs for the existing jobs.
func (c *Client) Queued(jobID string) ([]Queue, error) {
	resp, err := c.http.Get("/v1/queue")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var queued []Queue
		if err = json.NewDecoder(resp.Body).Decode(&queued); err != nil {
			return nil, err
		}
		if jobID == "" {
			return queued, nil
		}
		for _, queue := range queued {
			if queue.JobID == jobID {
				return []Queue{queue}, nil
			}
		}
		return nil, fmt.Errorf(`job "%s" does not exist`, jobID)
	default:
		return httpResponseToError(resp)
	}
}

func httpResponseToError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	var apiError *Error
	if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
		return &httpclient.HTTPError{resp}
	}
	apiError.Code = resp.StatusCode
	return apiError
}
