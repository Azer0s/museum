package domain

type Status string

const (
	Starting   Status = "starting"
	Running    Status = "running"
	Stopping   Status = "stopping"
	Stopped    Status = "stopped"
	NotCreated Status = "not_created"
)

type ExhibitRuntimeInfo struct {
	Status            Status   `json:"status"`
	Hostname          string   `json:"hostname"`
	RelatedContainers []string `json:"related_containers"`
	LastAccessed      int64    `json:"-"`
}

func (e *ExhibitRuntimeInfo) ToDto() RuntimeInfoDto {
	return RuntimeInfoDto{
		Status:       e.Status,
		LastAccessed: e.LastAccessed,
	}
}
