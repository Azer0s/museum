package main

import (
	"fmt"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"museum/ioc"
	"museum/persistence"
)

func main() {
	c := ioc.NewContainer()

	// register redis
	var redisClientDeferFunc func()
	ioc.RegisterSingleton[*goredislib.Client](c, func() *goredislib.Client {
		client, deferFunc := persistence.NewRedisClient()
		redisClientDeferFunc = deferFunc
		return client
	})
	defer redisClientDeferFunc()

	ioc.RegisterSingleton[persistence.SharedPersistentState](c, persistence.NewRedisStateConnector)

	// register kafka consumer group
	ioc.RegisterSingleton[*kafka.ConsumerGroup](c, func() *kafka.ConsumerGroup {
		consumerGroup, err := persistence.NewKafkaConsumerGroup()
		if err != nil {
			panic(err)
		}
		return consumerGroup
	})
	ioc.RegisterSingleton[persistence.Consumer](c, persistence.NewKafkaConsumer)

	// register kafka producer
	ioc.RegisterSingleton[*kafka.Writer](c, persistence.NewKafkaWriter)
	ioc.RegisterSingleton[persistence.Emitter](c, persistence.NewKafkaEmitter)

	// register shared state
	ioc.RegisterSingleton[persistence.SharedPersistentEmittedState](c, persistence.NewSharedPersistentEmittedState)

	graph := ioc.GenerateDependencyGraph(c)
	fmt.Println(graph)
}
