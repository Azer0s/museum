package main

import (
	"fmt"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"museum/ioc"
	"museum/persistence"
)

func main() {
	// register redis
	var redisClientDeferFunc func()
	ioc.RegisterImpl[*goredislib.Client](func() *goredislib.Client {
		client, deferFunc := persistence.NewRedisClient()
		redisClientDeferFunc = deferFunc
		return client
	})
	defer redisClientDeferFunc()

	ioc.RegisterImpl[persistence.SharedPersistentState](persistence.NewRedisStateConnector)

	ioc.RegisterImpl[*kafka.ConsumerGroup](func() *kafka.ConsumerGroup {
		consumerGroup, err := persistence.NewKafkaConsumerGroup()
		if err != nil {
			panic(err)
		}
		return consumerGroup
	})
	ioc.RegisterImpl[*kafka.Writer](persistence.NewKafkaWriter)
	ioc.RegisterImpl[persistence.Emitter](persistence.NewKafkaEmitter)

	graph := ioc.GenerateDependencyGraph()
	fmt.Println(graph)
}
