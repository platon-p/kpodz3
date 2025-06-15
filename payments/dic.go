package main

import (
	"context"

	"github.com/platon-p/kpodz3/payments/application/server"
	"github.com/platon-p/kpodz3/payments/application/services"
	"github.com/platon-p/kpodz3/payments/application/workers"
	"github.com/platon-p/kpodz3/payments/infra"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Config struct {
	Redis          string `env:"REDIS"`
	ServerPort     int    `env:"SERVER_PORT"`
	MQAddr         string `env:"MQ_ADDR"`
	QueueOrder2Pay string `env:"QUEUE_ORDER_2_PAY"`
	QueuePay2Order string `env:"QUEUE_PAY_2_ORDER"`
}

func NewDefaultConfig() Config {
	return Config{
		Redis:          "localhost:6379",
		ServerPort:     8080,
		MQAddr:         "amqp://user:password@localhost:5672/",
		QueueOrder2Pay: "order2pay",
		QueuePay2Order: "pay2order",
	}
}

func run(cfg Config) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis,
	})

	conn, err := amqp091.Dial(cfg.MQAddr)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	order2pay, err := ch.QueueDeclare(cfg.QueueOrder2Pay, true, false, false, false, nil)
	if err != nil {
		return err
	}
	pay2order, err := ch.QueueDeclare(cfg.QueuePay2Order, true, false, false, false, nil)
	if err != nil {
		return err
	}

	accountRepo := infra.NewRedisAccountRepo(redisClient)
	accountService := services.NewAccountService(accountRepo)

	httpServer := server.NewHTTPServer(cfg.ServerPort, accountService)
	httpServer.Setup()

	inbox := workers.NewInboxWorker(zap.L(), ch, &order2pay, accountRepo, accountService)
	outbox := workers.NewOutboxWorker(accountRepo, ch, &pay2order)
	tasker := workers.NewTaskWorker(accountService, accountRepo, ch, &pay2order)

	task := func() error {
		ctx := context.Background()
		go func() { inbox.Run(ctx) }()
		go func() { outbox.Run(ctx) }()
		go func() { tasker.Run(ctx) }()

		return httpServer.Serve()
	}
	return task()
}
