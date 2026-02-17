package domain

import "time"

type ContainerMetrics struct {
	CPUPercentage float64   `json:"cpu_percentage"`
	MemoryUsage   uint64    `json:"memory_usage"`
	MemoryLimit   uint64    `json:"memory_limit"`
	MemoryPercent float64   `json:"memory_percent,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}
