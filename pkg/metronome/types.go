package metronome

import (
	"time"
)

const apiTimeFormat = "2006-01-02T15:04:05.000-0700"

// Job represents a Job returned by the Metronome API.
type Job struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels,omitempty"`
	// The run property of a Job represents the run configuration for that Job
	Run struct {
		Args                       []string               `json:"args,omitempty"`
		Artifacts                  []artifact             `json:"artifacts,omitempty"`
		Cmd                        string                 `json:"cmd"`
		Cpus                       float32                `json:"cpus"`
		Gpus                       float32                `json:"gpus"`
		Disk                       int                    `json:"disk"`
		Docker                     *docker                `json:"docker,omitempty"`
		Env                        map[string]interface{} `json:"env,omitempty"`
		MaxLaunchDelay             int                    `json:"maxLaunchDelay,omitempty"`
		Mem                        int                    `json:"mem"`
		Placement                  *placement             `json:"placement,omitempty"`
		Secrets                    map[string]interface{} `json:"secrets,omitempty"`
		TaskKillGracePeriodSeconds float64                `json:"taskKillGracePeriodSeconds"`
		UCR                        *ucr                   `json:"ucr,omitempty"`
		User                       string                 `json:"user,omitempty"`
		Restart                    *restart               `json:"restart,omitempty"`
		Volumes                    []volume               `json:"volumes,omitempty"`
	} `json:"run"`

	// These properties depend on the embed parameters when querying the /v1/jobs endpoints.
	ActiveRuns     []Run              `json:"activeRuns,omitempty"`
	HistorySummary *JobHistorySummary `json:"historySummary,omitempty"`
	History        *JobHistory        `json:"history,omitempty"`
	Schedules      []Schedule         `json:"schedules,omitempty"`
}

type artifact struct {
	URI        string `json:"uri"`
	Executable bool   `json:"executable"`
	Extract    bool   `json:"extract"`
	Cache      bool   `json:"cache"`
}

type docker struct {
	Image          string `json:"image,omitempty"`
	ForcePullImage bool   `json:"forcePullImage"`
	Privileged     bool   `json:"privileged"`
}

type ucr struct {
	Image      map[string]interface{} `json:"image,omitempty"`
	Privileged bool                   `json:"privileged"`
}

type placement struct {
	Constraints []constraint `json:"constraints"`
}

type constraint struct {
	Attribute string `json:"attribute"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
}

type restart struct {
	Policy                string `json:"policy,omitempty"`
	ActiveDeadlineSeconds int    `json:"activeDeadlineSeconds,omitempty"`
}

type volume struct {
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	Mode          string `json:"mode,omitempty"`
	Secret        string `json:"secret,omitempty"`
}

// Run contains information about a run of a Job.
type Run struct {
	ID          string       `json:"id"`
	JobID       string       `json:"jobId"`
	Status      string       `json:"status"`
	CreatedAt   string       `json:"createdAt"`
	CompletedAt string       `json:"completedAt"`
	Tasks       []activeTask `json:"tasks"`
}

type activeTask struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	StartedAt   string `json:"startedAt"`
	CompletedAt string `json:"completedAt"`
}

// JobHistory contains statistics and information about past runs of a Job.
type JobHistory struct {
	SuccessCount           int          `json:"successCount"`
	FailureCount           int          `json:"failureCount"`
	LastSuccessAt          string       `json:"lastSuccessAt"`
	LastFailureAt          string       `json:"lastFailureAt"`
	SuccessfulFinishedRuns []runHistory `json:"successfulFinishedRuns"`
	FailedRuns             []runHistory `json:"failedFinishedRuns"`
}

type runHistory struct {
	ID         string   `json:"id"`
	CreatedAt  string   `json:"createdAt"`
	FinishedAt string   `json:"finishedAt"`
	Tasks      []string `json:"tasks,omitempty"`
}

// JobHistorySummary contains statistics about past runs of a Job.
type JobHistorySummary struct {
	SuccessCount  int    `json:"successCount"`
	FailureCount  int    `json:"failureCount"`
	LastSuccessAt string `json:"lastSuccessAt"`
	LastFailureAt string `json:"lastFailureAt"`
}

// Schedule of a Job.
type Schedule struct {
	ID                      string `json:"id"`
	Cron                    string `json:"cron"`
	TimeZone                string `json:"timeZone,omitempty"`
	StartingDeadlineSeconds int    `json:"startingDeadlineSeconds,omitempty"`
	ConcurrencyPolicy       string `json:"concurrencyPolicy"`
	Enabled                 bool   `json:"enabled"`
	NextRunAt               string `json:"nextRunAt"`
}

// Queue contains all queued runs of a Job.
type Queue struct {
	JobID string      `json:"jobId"`
	Runs  []queuedRun `json:"runs"`
}

type queuedRun struct {
	ID string `json:"runId"`
}

// Status returns the status of the job depending on its active runs and its schedule.
func (j *Job) Status() string {
	switch {
	case j.ActiveRuns != nil:
		return "Running"
	case len(j.Schedules) == 0:
		return "Unscheduled"
	default:
		return "Scheduled"
	}
}

// LastRunStatus returns the status of the last run of this job.
func (j *Job) LastRunStatus() string {
	if j.HistorySummary.LastSuccessAt == "" && j.HistorySummary.LastFailureAt == "" {
		return "N/A"
	} else if j.HistorySummary.LastFailureAt == "" {
		return "Success"
	} else if j.HistorySummary.LastSuccessAt == "" {
		return "Failure"
	}

	lastSuccess, err := time.Parse(apiTimeFormat, j.HistorySummary.LastSuccessAt)
	if err != nil {
		return "N/A"
	}

	lastFailure, err := time.Parse(apiTimeFormat, j.HistorySummary.LastFailureAt)
	if err != nil {
		return "N/A"
	}

	if lastSuccess.After(lastFailure) {
		return "Success"
	}
	return "Failure"

}
