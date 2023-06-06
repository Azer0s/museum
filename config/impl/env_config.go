package impl

import (
	proxy_mode "museum/config/proxy-mode"
)

type EnvConfig struct {
	EtcdHost    string `env:"ETCD_HOST,required"`
	EtcdBaseKey string `env:"ETCD_BASE_KEY" envDefault:"museum"`
	DockerHost  string `env:"DOCKER_HOST" envDefault:"unix:///var/run/docker.sock"`
	Hostname    string `env:"HOSTNAME" envDefault:"localhost"`
	Port        string `env:"PORT" envDefault:"8080"`
	JaegerHost  string `env:"JAEGER_HOST"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	ProxyMode   string `env:"PROXY_MODE" envDefault:"swarm-ext"`
	DevProxyUrl string `env:"DEV_PROXY_URL" envDefault:"http://localhost:3000"`
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

func (e EnvConfig) GetProxyMode() proxy_mode.Mode {
	switch e.ProxyMode {
	case "swarm":
		return proxy_mode.ModeSwarm
	case "swarm-ext":
		return proxy_mode.ModeSwarmExt
	default:
		panic("invalid proxy mode" + e.ProxyMode)
	}
}

func (e EnvConfig) GetDevProxyUrl() string {
	return e.DevProxyUrl
}
