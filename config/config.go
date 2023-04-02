package config

type Config interface {
	GetNatsHost() string
	GetNatsSubject() string
	GetRedisHost() string
	GetRedisBaseKey() string
	GetDockerHost() string
	GetHostname() string
	GetPort() string
	GetJaegerHost() string
	GetEnvironment() string
}
