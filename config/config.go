package config

import proxymode "museum/config/proxy-mode"

type Config interface {
	GetEtcdHost() string
	GetEtcdBaseKey() string
	GetNatsHost() string
	GetNatsBaseKey() string
	GetDockerHost() string
	GetHostname() string
	GetPort() string
	GetJaegerHost() string
	GetEnvironment() string
	GetProxyMode() proxymode.Mode
	GetCertFile() string
	GetKeyFile() string
	GetStartingTimeout() int
}
