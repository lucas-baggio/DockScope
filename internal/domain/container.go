package domain

import "time"

type Container struct {
	ID         string
	Names      []string
	Image      string
	ImageID    string
	Status     string
	State      string
	CreatedAt  time.Time
	Labels     map[string]string
	Ports      []PortBinding
	Mounts     []Mount
	HostConfig *HostConfig
}

type PortBinding struct {
	PrivatePort uint16
	PublicPort  uint16
	Type        string
	IP          string
}

type Mount struct {
	Type   string
	Source string
	Target string
}

type HostConfig struct {
	NetworkMode string
}
