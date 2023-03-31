package domain

type Exhibit struct {
	Id      string   `json:"id"`
	Name    string   `json:"name" yaml:"name"`
	Expose  string   `json:"expose" yaml:"expose"`
	Objects []Object `json:"objects" yaml:"objects"`
	Lease   string   `json:"lease" yaml:"lease"`
}

type Object struct {
	Name        string    `json:"name" yaml:"name"`
	Image       string    `json:"image" yaml:"image"`
	Label       string    `json:"label" yaml:"label"`
	Environment StringMap `json:"environment" yaml:"environment"`
	Mounts      StringMap `json:"mounts" yaml:"mounts"`
	Volumes     []Volume  `json:"volumes" yaml:"volumes"`
	Port        *string   `json:"port" yaml:"port"`
}

type Volume struct {
	Name   string `json:"name" yaml:"name"`
	Driver Driver `json:"driver" yaml:"driver"`
}

type Driver struct {
	Type   string    `json:"type" yaml:"type"`
	Config StringMap `json:"config" yaml:"config"`
}

type StringMap map[string]string

func (e *StringMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var env map[string]string
	if err := unmarshal(&env); err != nil {
		return err
	}
	*e = env
	return nil
}
