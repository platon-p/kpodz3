package main

import (
	"context"

	"github.com/platon-p/kpodz3/payments/application/server"
	"github.com/platon-p/kpodz3/payments/application/services"
	"github.com/platon-p/kpodz3/payments/application/workers"
	"github.com/platon-p/kpodz3/payments/infra"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Redis      string `json:"redis" env:"REDIS" default:"localhost:6379"`
	ServerPort int    `json:"server_port" env:"SERVER_PORT" default:"8080"`
}

func run(cfg Config) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis,
	})
	accountRepo := infra.NewRedisAccountRepo(redisClient)
	accountService := services.NewAccountService(accountRepo)

	httpServer := server.NewHTTPServer(cfg.ServerPort, accountService)
	httpServer.Setup()

	conn, err := amqp091.Dial("amqp://user:password@localhost:5672/")
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare("myqueue", true, false, false, false, nil)
	if err != nil {
		return err
	}

	inbox := workers.NewInboxWorker(ch, &queue, accountRepo, accountService)
	go func() { inbox.Run(context.Background()) }()

	outbox := workers.NewOutboxWorker(accountRepo, ch, &queue)
	go func() { outbox.Run(context.Background()) }()

	tasker := workers.NewTaskWorker(accountService, accountRepo, ch, &queue)
	go func() { tasker.Run(context.Background()) }()

	return httpServer.Serve()
}
