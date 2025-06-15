package main

import (
	"github.com/platon-p/kpodz3/payments/application/server"
	"github.com/platon-p/kpodz3/payments/application/services"
	"github.com/platon-p/kpodz3/payments/infra"
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
	return httpServer.Serve()
}
