package diagnostics

// Bundle represents a diagnostics bundle.
type Bundle struct {
	File string `json:"file_name"`
	Size int64  `json:"file_size"`
}

// BundleReportStatus is the response from /system/health/v1/report/diagnostics/status/all.
type BundleReportStatus struct {
	// job related fields
	Running               bool     `json:"is_running"`
	Status                string   `json:"status"`
	Errors                []string `json:"errors"`
	LastBundlePath        string   `json:"last_bundle_dir"`
	JobStarted            string   `json:"job_started"`
	JobEnded              string   `json:"job_ended"`
	JobDuration           string   `json:"job_duration"`
	JobProgressPercentage float32  `json:"job_progress_percentage"`

	// config related fields
	DiagnosticBundlesBaseDir                 string `json:"diagnostics_bundle_dir"`
	DiagnosticsJobTimeoutMin                 int    `json:"diagnostics_job_timeout_min"`
	DiagnosticsUnitsLogsSinceHours           string `json:"journald_logs_since_hours"`
	DiagnosticsJobGetSingleURLTimeoutMinutes int    `json:"diagnostics_job_get_since_url_timeout_min"`
	CommandExecTimeoutSec                    int    `json:"command_exec_timeout_sec"`

	// metrics related
	DiskUsedPercent float64 `json:"diagnostics_partition_disk_usage_percent"`
}

// BundleCreate is the json request used to create a diagnostics bundle.
type BundleCreate struct {
	Nodes []string `json:"nodes"`
}

// BundleCreateResponseJSONStruct is the response from /system/health/v1/report/diagnostics/create.
type BundleCreateResponseJSONStruct struct {
	Status string `json:"status"`
	Extra  struct {
		BundleName string `json:"bundle_name"`
	}
}

// BundleGenericResponseJSONStruct is the generic response from /system/health/v1/report/diagnostics endppoints.
type BundleGenericResponseJSONStruct struct {
	Status string `json:"status"`
}

// UnitsHealthResponseJSONStruct is the response from /system/health/v1.
type UnitsHealthResponseJSONStruct struct {
	Array       []HealthResponseValues `json:"units"`
	Hostname    string                 `json:"hostname"`
	IPAddress   string                 `json:"ip"`
	DcosVersion string                 `json:"dcos_version"`
	Role        string                 `json:"node_role"`
	MesosID     string                 `json:"mesos_id"`
	TdtVersion  string                 `json:"dcos_diagnostics_version"`
}

// HealthResponseValues is the response from
type HealthResponseValues struct {
	UnitID     string `json:"id"`
	UnitHealth int    `json:"health"`
	UnitOutput string `json:"output"`
	UnitTitle  string `json:"description"`
	Help       string `json:"help"`
	PrettyName string `json:"name"`
}
