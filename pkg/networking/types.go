package networking

import "time"

// Node is the struct representing what is returned by the networking API.
type Node struct {
	Updated   time.Time `json:"updated"`
	PublicIPs []string  `json:"public_ips"`
	PrivateIP string    `json:"private_ip"`
	Hostname  string    `json:"hostname"`
}
