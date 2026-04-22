package device

import "time"

type Service struct {
	Port int    `json:"port"`
	Name string `json:"name"`
}

type Device struct {
	IP       string    `json:"ip"`
	Hostname string    `json:"hostname,omitempty"`
	LastSeen time.Time `json:"last_seen"`
	Online   bool      `json:"online"`
	Services []Service `json:"services,omitempty"`
}
