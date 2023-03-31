package impl

type EnvConfig struct {
	KafkaBrokers []string `env:"KAFKA_BROKERS,required" envSeparator:","`
	KafkaTopic   string   `env:"KAFKA_TOPIC" envDefault:"museum"`
	RedisHost    string   `env:"REDIS_HOST,required"`
	RedisBaseKey string   `env:"REDIS_BASE_KEY" envDefault:"museum"`
	DockerHost   string   `env:"DOCKER_HOST" envDefault:"unix:///var/run/docker.sock"`
	Hostname     string   `env:"HOSTNAME" envDefault:"localhost"`
	Port         string   `env:"PORT" envDefault:"8080"`
	JaegerHost   string   `env:"JAEGER_HOST"`
	Environment  string   `env:"ENVIRONMENT" envDefault:"development"`
}

func (e EnvConfig) GetKafkaBrokers() []string {
	return e.KafkaBrokers
}

func (e EnvConfig) GetKafkaTopic() string {
	return e.KafkaTopic
}

func (e EnvConfig) GetRedisHost() string {
	return e.RedisHost
}

func (e EnvConfig) GetRedisBaseKey() string {
	return e.RedisBaseKey
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
