package metrics

import "time"

// Node represents the metrics of a node.
type Node struct {
	Datapoints []Datapoint `json:"datapoints"`
	Dimensions struct {
		MesosID   string `json:"mesos_id"`
		ClusterID string `json:"cluster_id"`
		Hostname  string `json:"hostname"`
	} `json:"dimensions"`
}

// Container represents the metrics of a container.
type Container struct {
	Datapoints []Datapoint `json:"datapoints"`
	Dimensions struct {
		MesosID       string `json:"mesos_id"`
		ClusterID     string `json:"cluster_id"`
		ContainerID   string `json:"container_id"`
		FrameworkName string `json:"framework_name"`
		TaskName      string `json:"task_name"`
		Hostname      string `json:"hostname"`
		Labels        struct {
			DcosClusterID     string `json:"dcos_cluster_id"`
			DcosClusterName   string `json:"dcos_cluster_name"`
			FaultDomainRegion string `json:"fault_domain_region"`
			FaultDomainZone   string `json:"fault_domain_zone"`
			Host              string `json:"host"`
		} `json:"labels"`
	} `json:"dimensions"`
}

// Datapoint represents a datapoint of a node.
type Datapoint struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      map[string]string `json:"tags,omitempty"`
}
