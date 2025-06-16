package main

import (
	"context"
	"time"

	"github.com/platon-p/kpodz3/orders/application/server"
	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/application/workers"
	"github.com/platon-p/kpodz3/orders/infra"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Config struct {
	RedisAddr      string `env:"REDIS_ADDR"`
	ServerPort     int    `env:"SERVER_PORT"`
	MQAddr         string `env:"MQ_ADDR"`
	QueueOrder2Pay string `env:"QUEUE_ORDER_2_PAY"`
	QueuePay2Order string `env:"QUEUE_PAY_2_ORDER"`
}

func NewDefaultConfig() Config {
	return Config{
		RedisAddr:      "localhost:6378",
		ServerPort:     8888,
		MQAddr:         "amqp://user:password@localhost:5672/",
		QueueOrder2Pay: "order2pay",
		QueuePay2Order: "pay2order",
	}
}

func run(cfg Config) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	time.Sleep(8 * time.Second)
	conn, err := amqp091.Dial(cfg.MQAddr)
	if err != nil {
		return err
	}
	mqchan, err := conn.Channel()
	if err != nil {
		return err
	}
	pay2order, err := mqchan.QueueDeclare(cfg.QueuePay2Order, true, false, false, false, nil)
	if err != nil {
		return err
	}
	order2pay, err := mqchan.QueueDeclare(cfg.QueueOrder2Pay, true, false, false, false, nil)
	if err != nil {
		return err
	}

	orderRepo := infra.NewRedisOrderRepo(redisClient)
	orderService := services.NewOrderService(orderRepo)

	outbox := workers.NewOutboxWorker(orderRepo, mqchan, &order2pay)
	tasker := workers.NewTaskWorker(zap.L(), orderService, mqchan, &pay2order)

	httpServer := server.NewHTTPServer(cfg.ServerPort, orderService)
	httpServer.Setup()

	task := func() error {
		ctx := context.Background()
		go func() { outbox.Run(ctx) }()
		go func() { tasker.Run(ctx) }()
		return httpServer.Run()
	}

	return task()
}
