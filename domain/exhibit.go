package domain

type Exhibit struct {
	Id          string              `json:"id"`
	Name        string              `json:"name" yaml:"name"`
	Expose      string              `json:"expose" yaml:"expose"`
	Rewrite     *bool               `json:"rewrite" yaml:"rewrite"`
	Objects     []Object            `json:"objects" yaml:"objects"`
	Lease       string              `json:"lease" yaml:"lease"`
	Order       []string            `json:"order" yaml:"order"`
	RuntimeInfo *ExhibitRuntimeInfo `json:"-"`
}

func (e Exhibit) ToDto() ExhibitDto {
	objects := make([]ObjectDto, 0)
	for _, o := range e.Objects {
		objects = append(objects, o.ToDto())
	}

	return ExhibitDto{
		Id:          e.Id,
		Name:        e.Name,
		RuntimeInfo: e.RuntimeInfo.ToDto(),
		Lease:       e.Lease,
		Objects:     objects,
	}
}

func (e Exhibit) GetTotalSteps() int {
	steps := 0
	for _, object := range e.Objects {
		steps += 4
		if object.Livecheck != nil {
			steps++
		}
	}
	return steps
}
