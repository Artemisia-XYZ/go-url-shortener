package database

import (
	"context"
	"fmt"
	"url-shortener/helpers"

	"github.com/redis/go-redis/v9"
)

func NewRedis() *redis.Client {
	addr := helpers.Getenv("REDIS_HOST", "127.0.0.1") + ":" + helpers.Getenv("REDIS_PORT", "6379")
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: helpers.Getenv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		panic(fmt.Sprintf("can't connect to redis: %v", err))
	}

	return client
}
