package domain

type Exhibit struct {
	Id          string              `json:"id"`
	Name        string              `json:"name" yaml:"name"`
	Expose      string              `json:"expose" yaml:"expose"`
	Objects     []Object            `json:"objects" yaml:"objects"`
	Lease       string              `json:"lease" yaml:"lease"`
	Order       []string            `json:"order" yaml:"order"`
	RuntimeInfo *ExhibitRuntimeInfo `json:"-"`
}

func (e Exhibit) ToDto() ExhibitDto {
	return ExhibitDto{
		Id:          e.Id,
		Name:        e.Name,
		RuntimeInfo: e.RuntimeInfo,
	}
}
