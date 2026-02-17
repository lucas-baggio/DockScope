package domain

import "time"

type Image struct {
	ID          string
	RepoTags    []string
	RepoDigests []string
	CreatedAt   time.Time
	Size        int64
	SharedSize  int64
	VirtualSize int64
	Labels      map[string]string
	ParentID    string
}
