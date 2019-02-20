package logs

// Entry refers to a DC/OS log entry.
type Entry struct {
	Fields             EntryFields `json:"fields"`
	Cursor             string      `json:"cursor"`
	MonotonicTimestamp int64       `json:"monotonic_timestamp"`
	RealtimeTimestamp  int64       `json:"realtime_timestamp"`
}

// EntryFields are the fields in a DC/OS log entry.
type EntryFields struct {
	Message          string `json:"MESSAGE"`
	Priority         string `json:"PRIORITY"`
	SyslogFacility   string `json:"SYSLOG_FACILITY"`
	SyslogIdentifier string `json:"SYSLOG_IDENTIFIER"`
	BootID           string `json:"_BOOT_ID"`
	CapEffective     string `json:"_CAP_EFFECTIVE"`
	Cmdline          string `json:"_CMDLINE"`
	Comm             string `json:"_COMM"`
	Exe              string `json:"_EXE"`
	GID              string `json:"_GID"`
	Hostname         string `json:"_HOSTNAME"`
	MachineID        string `json:"_MACHINE_ID"`
	PID              string `json:"_PID"`
	SelinuxContext   string `json:"_SELINUX_CONTEXT"`
	StreamID         string `json:"_STREAM_ID"`
	SystemdCgroup    string `json:"_SYSTEMD_CGROUP"`
	SystemdSlice     string `json:"_SYSTEMD_SLICE"`
	SystemdUnit      string `json:"_SYSTEMD_UNIT"`
	Transport        string `json:"_TRANSPORT"`
	UID              string `json:"_UID"`
}
