package domain

type Status string

const (
	Running    Status = "running"
	Stopped    Status = "stopped"
	NotCreated Status = "not_created"
)

type ExhibitRuntimeInfo struct {
	Status            Status   `json:"status"`
	LastAccessed      string   `json:"last_accessed"`
	Hostname          string   `json:"hostname"`
	RelatedContainers []string `json:"related_containers"`
}
