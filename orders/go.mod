module github.com/platon-p/kpodz3/orders

go 1.23.8

replace github.com/platon-p/kpodz3/proto => ../proto

require (
	github.com/platon-p/kpodz3/proto v0.0.0
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/redis/go-redis/v9 v9.10.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
