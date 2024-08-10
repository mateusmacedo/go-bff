package adapter

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() redis.UniversalClient {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

var Ctx = context.Background()
