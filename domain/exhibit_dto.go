package domain

type ExhibitDto struct {
	Id          string                 `json:"id"`
	Name        string                 `json:"name"`
	RuntimeInfo RuntimeInfoDto         `json:"runtime_info"`
	Lease       string                 `json:"lease"`
	Objects     []ObjectDto            `json:"objects"`
	Meta        map[string]interface{} `json:"meta"`
}

func (d ExhibitDto) ToExhibit() Exhibit {
	return Exhibit{
		Id:    d.Id,
		Name:  d.Name,
		Lease: d.Lease,
		Meta:  d.Meta,
	}
}
