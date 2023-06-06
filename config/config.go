package config

import proxy_mode "museum/config/proxy-mode"

type Config interface {
	GetEtcdHost() string
	GetEtcdBaseKey() string
	GetDockerHost() string
	GetHostname() string
	GetPort() string
	GetJaegerHost() string
	GetEnvironment() string
	GetProxyMode() proxy_mode.Mode
	GetDevProxyUrl() string
}
