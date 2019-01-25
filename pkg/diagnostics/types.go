package diagnostics

// UnitsHealthResponseJSONStruct json response /system/health/v1
type UnitsHealthResponseJSONStruct struct {
	Array       []HealthResponseValues `json:"units"`
	Hostname    string                 `json:"hostname"`
	IPAddress   string                 `json:"ip"`
	DcosVersion string                 `json:"dcos_version"`
	Role        string                 `json:"node_role"`
	MesosID     string                 `json:"mesos_id"`
	TdtVersion  string                 `json:"dcos_diagnostics_version"`
}

// HealthResponseValues is a health values json response.
type HealthResponseValues struct {
	UnitID     string `json:"id"`
	UnitHealth int    `json:"health"`
	UnitOutput string `json:"output"`
	UnitTitle  string `json:"description"`
	Help       string `json:"help"`
	PrettyName string `json:"name"`
}
