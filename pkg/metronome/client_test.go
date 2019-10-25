package metronome

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJob(t *testing.T) {
	expectedJob := Job{
		ID:          "test-job",
		Description: "Job that sleeps 10 seconds",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedJob))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	job, err := c.Job("test-job")
	require.NoError(t, err)
	require.Equal(t, &expectedJob, job)
}

func TestJobs(t *testing.T) {
	expectedJobs := []Job{
		{
			ID:          "test-job-1",
			Description: "Job that sleeps 10 seconds",
		},
		{
			ID:          "test-job-2",
			Description: "Job that sleeps 20 seconds",
		},
		{
			ID:          "test-job-3",
			Description: "Job that sleeps 30 seconds",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedJobs))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	jobs, err := c.Jobs()
	require.NoError(t, err)
	require.Equal(t, expectedJobs, jobs)
}

func TestAddJob(t *testing.T) {
	expectedJob := Job{
		ID:          "test-job",
		Description: "Job that sleeps 10 seconds",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		var payload Job
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, expectedJob, payload)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedJob))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	job, err := c.AddJob(&expectedJob)
	require.NoError(t, err)
	require.Equal(t, &expectedJob, job)
}

func TestUpdateJob(t *testing.T) {
	expectedJob := Job{
		ID:          "test-job",
		Description: "Job that sleeps 20 seconds",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job", r.URL.String())
		assert.Equal(t, "PUT", r.Method)
		var payload Job
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, expectedJob, payload)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedJob))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	job, err := c.UpdateJob(&expectedJob)
	require.NoError(t, err)
	require.Equal(t, &expectedJob, job)
}

func TestRunJob(t *testing.T) {
	expectedRun := Run{
		JobID: "test-job",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job/runs", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedRun))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	run, err := c.RunJob("test-job")
	require.NoError(t, err)
	require.Equal(t, &expectedRun, run)
}

func TestRun(t *testing.T) {
	expectedRun := Run{
		ID:    "20190307204634kC8Rs",
		JobID: "test-job",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job/runs/20190307204634kC8Rs", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedRun))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	run, err := c.Run("test-job", "20190307204634kC8Rs")
	require.NoError(t, err)
	require.Equal(t, &expectedRun, run)
}

func TestRuns(t *testing.T) {
	expectedRuns := []Run{
		{
			JobID: "test-job",
		},
		{
			JobID: "test-job",
		},
		{
			JobID: "test-job",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job/runs", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedRuns))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	runs, err := c.Runs("test-job")
	require.NoError(t, err)
	require.Equal(t, expectedRuns, runs)
}

func TestKill(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job/runs/20190307204634kC8Rs/actions/stop", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	err := c.Kill("test-job", "20190307204634kC8Rs")
	require.NoError(t, err)
}

func TestRemoveJob(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job?stopCurrentJobRuns=true", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	err := c.RemoveJob("test-job", true)
	require.NoError(t, err)
}

func TestSchedules(t *testing.T) {
	expectedSchedules := []Schedule{
		{
			ID:      "test-schedule-1",
			Enabled: true,
		},
		{
			ID:      "test-schedule-2",
			Enabled: false,
		},
		{
			ID:      "test-schedule-3",
			Enabled: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job/schedules", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedSchedules))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	schedules, err := c.Schedules("test-job")
	require.NoError(t, err)
	require.Equal(t, expectedSchedules, schedules)
}

func TestAddSchedule(t *testing.T) {
	expectedSchedule := Schedule{
		ID:      "test-schedule",
		Enabled: true,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job/schedules", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		var payload Schedule
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, expectedSchedule, payload)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedSchedule))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	schedule, err := c.AddSchedule("test-job", &expectedSchedule)
	require.NoError(t, err)
	require.Equal(t, &expectedSchedule, schedule)
}

func TestUpdateSchedule(t *testing.T) {
	expectedSchedule := Schedule{
		ID:      "test-schedule",
		Enabled: false,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job/schedules/test-schedule", r.URL.String())
		assert.Equal(t, "PUT", r.Method)
		var payload Schedule
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, expectedSchedule, payload)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedSchedule))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	schedule, err := c.UpdateSchedule("test-job", &expectedSchedule)
	require.NoError(t, err)
	require.Equal(t, &expectedSchedule, schedule)
}

func TestRemoveSchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/jobs/test-job/schedules/test-schedule", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	err := c.RemoveSchedule("test-job", "test-schedule")
	require.NoError(t, err)
}

func TestQueued(t *testing.T) {
	expectedQueues := []Queue{
		{
			JobID: "test-job",
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/queue", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(&expectedQueues))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL), pluginutil.Logger())

	q1, err := c.Queued("")
	require.NoError(t, err)
	require.Equal(t, expectedQueues, q1)

	q2, err := c.Queued("test-job")
	require.NoError(t, err)
	require.Equal(t, expectedQueues, q2)

	_, err = c.Queued("not-existing-job")
	require.Errorf(t, err, `job "not-existing-job" does not exist`)
}
