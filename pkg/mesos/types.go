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
	GPUs  float64 `json:"gpus"`
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
	Status              string                 `json:"status"`
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
	Version             string      `json:"version"`
	ID                  string      `json:"id"`
	PID                 string      `json:"pid"`
	Hostname            string      `json:"hostname"`
	ActivatedSlaves     float64     `json:"activated_slaves"`
	DeactivatedSlaves   float64     `json:"deactivated_slaves"`
	Domain              Domain      `json:"domain"`
	Cluster             string      `json:"cluster"`
	Leader              string      `json:"leader"`
	Slaves              []Slave     `json:"slaves"`
	Frameworks          []Framework `json:"frameworks"`
	CompletedFrameworks []Framework `json:"completed_frameworks"`
}

// StateSummary summarizes the state of a mesos master.
type StateSummary struct {
	Hostname string
	Cluster  string
	Slaves   []Slave
}

// Framework represent a single framework of a mesos node.
type Framework struct {
	Active             bool       `json:"active"`
	Capabilities       []string   `json:"capabilities"`
	Checkpoint         bool       `json:"checkpoint"`
	CompletedTasks     []Task     `json:"completed_tasks"`
	Executors          []Executor `json:"executors"`
	CompletedExecutors []Executor `json:"completed_executors"`
	FailoverTimeout    float64    `json:"failover_timeout"`
	Hostname           string     `json:"hostname"`
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	PID                string     `json:"pid"`
	OfferedResources   Resources  `json:"offered_resources"`
	Offers             []Offer    `json:"offers"`
	RegisteredTime     float64    `json:"registered_time"`
	ReregisteredTime   float64    `json:"reregistered_time"`
	Resources          Resources  `json:"resources"`
	Role               string     `json:"role"`
	Tasks              []Task     `json:"tasks"`
	UnregisteredTime   float64    `json:"unregistered_time"`
	UsedResources      Resources  `json:"used_resources"`
	User               string     `json:"user"`
	WebuiURL           string     `json:"webui_url"`
	Labels             []Label    `json:"label"`
}

// Offer represents a single offer from a Mesos Slave to a Mesos master
type Offer struct {
	ID          string            `json:"id"`
	FrameworkID string            `json:"framework_id"`
	SlaveID     string            `json:"slave_id"`
	Hostname    string            `json:"hostname"`
	URL         URL               `json:"url"`
	Resources   Resources         `json:"resources"`
	Attributes  map[string]string `json:"attributes"`
}

// URL represents a single URL
type URL struct {
	Scheme     string      `json:"scheme"`
	Address    Address     `json:"address"`
	Path       string      `json:"path"`
	Parameters []Parameter `json:"parameters"`
}

// Address represents a single address.
// e.g. from a Slave or from a Master
type Address struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
}

// Parameter represents a single key / value pair for parameters
type Parameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Label represents a single key / value pair for labeling
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Task represent a single Mesos task
type Task struct {
	ExecutorID  string        `json:"executor_id"`
	FrameworkID string        `json:"framework_id"`
	ID          string        `json:"id"`
	Labels      []Label       `json:"labels"`
	Name        string        `json:"name"`
	Resources   Resources     `json:"resources"`
	SlaveID     string        `json:"slave_id"`
	State       string        `json:"state"`
	Statuses    []TaskStatus  `json:"statuses"`
	Discovery   TaskDiscovery `json:"discovery"`
	Container   Container     `json:"container"`
}

// Executor represents a single executor of a framework
type Executor struct {
	CompletedTasks []Task    `json:"completed_tasks"`
	Container      string    `json:"container"`
	Directory      string    `json:"directory"`
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Resources      Resources `json:"resources"`
	Source         string    `json:"source"`
	QueuedTasks    []Task    `json:"queued_tasks"`
	Tasks          []Task    `json:"tasks"`
}

// File represents an element returned when hitting the '/browse' endpoint.
type File struct {
	GID   string  `json:"gid"`
	Mode  string  `json:"mode"`
	MTime float64 `json:"mtime"`
	NLink float64 `json:"nlink"`
	Path  string  `json:"path"`
	Size  float64 `json:"size"`
	UID   string  `json:"uid"`
}

// Container represents one way a Mesos task can be ran
type Container struct {
	Type   string `json:"type"`
	Docker Docker `json:"docker,omitempty"`
}

// Docker is one type of Container
type Docker struct {
	Image          string        `json:"image"`
	Network        string        `json:"network"`
	PortMappings   []PortMapping `json:"port_mappings"`
	Priviledge     bool          `json:"priviledge"`
	Parameters     []Parameter   `json:"parameters"`
	ForcePullImage bool          `json:"force_pull_image"`
}

// PortMapping represents how containers ports map to host ports
type PortMapping struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
}

// TaskDiscovery represents the dicovery information of a task
type TaskDiscovery struct {
	Visibility string `json:"visibility"`
	Name       string `json:"name"`
	Ports      Ports  `json:"ports"`
}

// Ports represents a number of PortDetails
type Ports struct {
	Ports []PortDetails `json:"ports"`
}

// PortDetails represents details about a single port
type PortDetails struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
}

// TaskStatus represents the status of a single task
type TaskStatus struct {
	State           string          `json:"state"`
	Timestamp       float64         `json:"timestamp"`
	ContainerStatus ContainerStatus `json:"container_status"`
}

// ContainerStatus represents the status of a single container inside a task
type ContainerStatus struct {
	ContainerID  ContainerID   `json:"container_id"`
	NetworkInfos []NetworkInfo `json:"network_infos"`
}

// ContainerID represents the ID of a container
type ContainerID struct {
	Value  string       `json:"value"`
	Parent *ContainerID `json:"parent"`
}

// NetworkInfo represents information about the network of a container
type NetworkInfo struct {
	IPAddress   string      `json:"ip_address"`
	IPAddresses []IPAddress `json:"ip_addresses"`
}

// IPAddress represents a single IpAddress
type IPAddress struct {
	IPAddress string `json:"ip_address"`
}

// Roles represents a stripped down representation of mesos/roles
type Roles struct {
	Roles []struct {
		Quota Quota `json:"quota"`
	} `json:"roles"`
}

// Quota represents a role's quota
type Quota struct {
	Role      string                 `json:"role"`
	Consumed  map[string]interface{} `json:"consumed,omitempty"`
	Guarantee map[string]interface{} `json:"guarantee,omitempty"`
	Limit     map[string]interface{} `json:"limit,omitempty"`
}
