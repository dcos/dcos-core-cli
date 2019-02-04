package mesos

import "strings"

// Domain holds information about a nodes region and zone.
type Domain struct {
	FaultDomain struct {
		Region struct {
			Name string `json:"name"`
		} `json:"region"`
		Zone struct {
			Name string `json:"name"`
		} `json:"zone"`
	} `json:"fault_domain"`
}

// Host holds information about a host and its IP.
type Host struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
}

// Master represents a single mesos master node.
type Master struct {
	Host      string   `json:"host"`
	IP        string   `json:"ip"`
	PublicIPs []string `json:"public_ips"`
	Type      string   `json:"type"`
	Region    string   `json:"region"`
	Zone      string   `json:"zone"`
	ID        string   `json:"id"`
	PID       string   `json:"pid"`
	Version   string   `json:"version"`
}

// Resources represents a resource type for a task.
type Resources struct {
	CPUs  float64 `json:"cpus"`
	Disk  float64 `json:"disk"`
	Mem   float64 `json:"mem"`
	Ports string  `json:"ports"`
}

// Slave represents a single mesos slave node.
type Slave struct {
	TaskError           int                    `json:"TASK_ERROR"`
	TaskFailed          int                    `json:"TASK_FAILED"`
	TaskFinished        int                    `json:"TASK_FINISHED"`
	TaskKilled          int                    `json:"TASK_KILLED"`
	TaskKilling         int                    `json:"TASK_KILLING"`
	TaskLost            int                    `json:"TASK_LOST"`
	TaskRunning         int                    `json:"TASK_RUNNING"`
	TaskStaging         int                    `json:"TASK_STAGING"`
	TaskStarting        int                    `json:"TASK_STARTING"`
	TaskUnreachable     int                    `json:"TASK_UNREACHABLE"`
	Active              bool                   `json:"active"`
	Attributes          map[string]interface{} `json:"attributes"`
	Capabilities        []string               `json:"capabilities"`
	Domain              Domain                 `json:"domain"`
	FrameworkIDs        []string               `json:"framework_ids"`
	Hostname            string                 `json:"hostname"`
	PublicIPs           []string               `json:"public_ips"`
	ID                  string                 `json:"id"`
	PID                 string                 `json:"pid"`
	Port                int                    `json:"port"`
	Region              string                 `json:"region"`
	RegisteredTime      float64                `json:"registered_time"`
	Resources           Resources              `json:"resources"`
	UsedResources       Resources              `json:"used_resources"`
	OfferedResources    Resources              `json:"offered_resources"`
	ReservedResources   map[string]Resources   `json:"reserved_resources"`
	UnreservedResources Resources              `json:"unreserved_resources"`
	Type                string                 `json:"type"`
	Version             string                 `json:"version"`
	Zone                string                 `json:"zone"`
}

// IP returns the IP of a slave parsed from its PID.
func (s *Slave) IP() string {
	// PID format: Host@IP:Port
	s1 := strings.Split(s.PID, "@")
	s2 := strings.Split(s1[1], ":")
	return s2[0]
}

// State represents a state.json returned by a mesos master.
type State struct {
	Version           string  `json:"version"`
	ID                string  `json:"id"`
	PID               string  `json:"pid"`
	Hostname          string  `json:"hostname"`
	ActivatedSlaves   float64 `json:"activated_slaves"`
	DeactivatedSlaves float64 `json:"deactivated_slaves"`
	Domain            Domain  `json:"domain"`
	Cluster           string  `json:"cluster"`
	Leader            string  `json:"leader"`
	Slaves            []Slave `json:"slaves"`
}

// StateSummary summarizes the state of a mesos master.
type StateSummary struct {
	Hostname string
	Cluster  string
	Slaves   []Slave
}
