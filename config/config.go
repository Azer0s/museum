package config

type Config interface {
	GetEtcdHost() string
	GetEtcdBaseKey() string
	GetDockerHost() string
	GetHostname() string
	GetPort() string
	GetJaegerHost() string
	GetEnvironment() string
}
