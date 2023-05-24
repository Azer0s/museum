package impl

type EnvConfig struct {
	EtcdHost    string `env:"ETCD_HOST,required"`
	EtcdBaseKey string `env:"ETCD_BASE_KEY" envDefault:"museum"`
	DockerHost  string `env:"DOCKER_HOST" envDefault:"unix:///var/run/docker.sock"`
	Hostname    string `env:"HOSTNAME" envDefault:"localhost"`
	Port        string `env:"PORT" envDefault:"8080"`
	JaegerHost  string `env:"JAEGER_HOST"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

func (e EnvConfig) GetEtcdHost() string {
	return e.EtcdHost
}

func (e EnvConfig) GetEtcdBaseKey() string {
	return e.EtcdBaseKey
}

func (e EnvConfig) GetDockerHost() string {
	return e.DockerHost
}

func (e EnvConfig) GetHostname() string {
	return e.Hostname
}

func (e EnvConfig) GetPort() string {
	return e.Port
}

func (e EnvConfig) GetJaegerHost() string {
	return e.JaegerHost
}

func (e EnvConfig) GetEnvironment() string {
	return e.Environment
}
