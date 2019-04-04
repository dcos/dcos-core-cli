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

// JournalctlJSONEntry are the fields that are printed for output=[json, json-pretty].
type JournalctlJSONEntry struct {
	Cursor             string `json:"__CURSOR"`
	MonotonicTimestamp int64  `json:"__MONOTONIC_TIMESTAMP"`
	RealtimeTimestamp  int64  `json:"__REALTIME_TIMESTAMP"`
	Message            string `json:"MESSAGE"`
	Priority           string `json:"PRIORITY"`
	SyslogFacility     string `json:"SYSLOG_FACILITY"`
	SyslogIdentifier   string `json:"SYSLOG_IDENTIFIER"`
	BootID             string `json:"_BOOT_ID"`
	Hostname           string `json:"_HOSTNAME"`
	MachineID          string `json:"_MACHINE_ID"`
	Transport          string `json:"_TRANSPORT"`
}

// JournalctlJSON generates a journalctl JSON log entry from a DC/OS log entry.
func (e *Entry) JournalctlJSON() JournalctlJSONEntry {
	return JournalctlJSONEntry{
		Cursor:             e.Cursor,
		MonotonicTimestamp: e.MonotonicTimestamp,
		RealtimeTimestamp:  e.RealtimeTimestamp,
		Message:            e.Fields.Message,
		Priority:           e.Fields.Priority,
		SyslogFacility:     e.Fields.SyslogFacility,
		SyslogIdentifier:   e.Fields.SyslogIdentifier,
		BootID:             e.Fields.BootID,
		Hostname:           e.Fields.Hostname,
		MachineID:          e.Fields.MachineID,
		Transport:          e.Fields.Transport,
	}
}
