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

// Datapoint represents a datapoint of a node.
type Datapoint struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      map[string]string `json:"tags,omitempty"`
}
