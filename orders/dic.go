package main

import (
	"context"

	"github.com/platon-p/kpodz3/orders/application/server"
	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/application/workers"
	"github.com/platon-p/kpodz3/orders/infra"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Config struct {
	RedisAddr  string `env:"REDIS_ADDR"`
	ServerPort int    `env:"SERVER_PORT"`
	MQAddr     string `env:"MQ_ADDR"`
	QueueName  string `env:"QUEUE_NAME"`
}

func NewDefaultConfig() Config {
	return Config{
		RedisAddr:  "localhost:6378",
		ServerPort: 8888,
		MQAddr:     "amqp://user:password@localhost:5672/",
		QueueName:  "myqueue",
	}
}

func run(cfg Config) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	conn, err := amqp091.Dial(cfg.MQAddr)
	if err != nil {
		return err
	}
	mqchan, err := conn.Channel()
	if err != nil {
		return err
	}
	myQueue, err := mqchan.QueueDeclare(cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	orderRepo := infra.NewRedisOrderRepo(redisClient)
	orderService := services.NewOrderService(orderRepo)

	outbox := workers.NewOutboxWorker(orderRepo, mqchan, &myQueue)
	tasker := workers.NewTaskWorker(zap.L(), orderService, mqchan, &myQueue)

	httpServer := server.NewHTTPServer(cfg.ServerPort, orderService)
	httpServer.Setup()

	task := func() error {
		ctx := context.Background()
		go func() {
			outbox.Run(ctx)
		}()
		go func() {
			tasker.Run(ctx)
		}()
		return httpServer.Run()
	}

	return task()
}
