package domain

type ExhibitDto struct {
	Id          string              `json:"id"`
	Name        string              `json:"name"`
	RuntimeInfo *ExhibitRuntimeInfo `json:"runtime_info"`
}
