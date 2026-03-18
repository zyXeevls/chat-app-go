package database

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
}
