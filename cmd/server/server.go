package server

import (
	"fmt"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"museum/config"
	"museum/ioc"
	"museum/observability"
	"museum/persistence"
)

func Run() {
	c := ioc.NewContainer()

	// register config
	ioc.RegisterSingleton[config.Config](c, config.NewEnvConfig)

	// register jaeger
	ioc.RegisterSingleton[tracesdk.SpanExporter](c, observability.NewSpanExporter)
	ioc.RegisterSingleton[trace.TracerProvider](c, observability.NewTracerProvider)

	// register redis
	ioc.RegisterSingleton[*goredislib.Client](c, persistence.NewRedisClient)
	ioc.RegisterSingleton[persistence.SharedPersistentState](c, persistence.NewRedisStateConnector)

	// register kafka consumer group
	ioc.RegisterSingleton[*kafka.ConsumerGroup](c, persistence.NewKafkaConsumerGroup)
	ioc.RegisterSingleton[persistence.Consumer](c, persistence.NewKafkaConsumer)

	// register kafka producer
	ioc.RegisterSingleton[*kafka.Writer](c, persistence.NewKafkaWriter)
	ioc.RegisterSingleton[persistence.Emitter](c, persistence.NewKafkaEmitter)

	// register shared state
	ioc.RegisterSingleton[persistence.SharedPersistentEmittedState](c, persistence.NewSharedPersistentEmittedState)

	graph := ioc.GenerateDependencyGraph(c)
	fmt.Println(graph)
}
