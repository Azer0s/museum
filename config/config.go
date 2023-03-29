package config

type Config interface {
	GetKafkaBrokers() []string
	GetKafkaTopic() string
	GetRedisHost() string
	GetRedisBaseKey() string
	GetDockerHost() string
	GetHostname() string
	GetPort() string
	GetJaegerHost() string
	GetEnvironment() string
}
