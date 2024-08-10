package adapter

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() redis.UniversalClient {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Endereço do Redis
		Password: "",               // Senha, se necessário
		DB:       0,                // Número do banco de dados
	})
}

var Ctx = context.Background()
