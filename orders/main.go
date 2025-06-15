package main

import (
	"context"
	"fmt"
	"time"

	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/application/workers"
	"github.com/platon-p/kpodz3/orders/domain"
	"github.com/platon-p/kpodz3/orders/infra"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	os := infra.NewRedisOrderRepo(redisClient)
	s := services.NewOrderService(os)

	s.SubscribeOrderCreated(func(o domain.Order) {
		fmt.Println(o)
	})

	q, err := amqp091.Dial("amqp://user:password@localhost:5672/")
	if err != nil {
		panic(err)
	}
	mqchan, err := q.Channel()
	if err != nil {
		panic(err)
	}

	w := workers.NewEventPublishWorker(mqchan)
	s.SubscribeOrderCreated(w.PublishOrderCreated)

	_, err = s.CreateOrder(context.Background(), 1, "test", 100)
	fmt.Println(err)

	time.Sleep(10 * time.Second)
}
