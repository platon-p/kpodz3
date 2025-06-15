package main

import (
	"context"
	"time"

	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/application/workers"
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

	conn, err := amqp091.Dial("amqp://user:password@localhost:5672/")
	if err != nil {
		panic(err)
	}
	mqchan, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	myQueue, err := mqchan.QueueDeclare("myqueue", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	w := workers.NewEventPublishWorker(os, mqchan, &myQueue)
	_ = w

	_, err = s.CreateOrder(context.Background(), 1, "test", 100)

	go func() {
		w.Run(context.Background())
	}()
	time.Sleep(10 * time.Second)
}
