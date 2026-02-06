package db

import (
	"context"
	"fmt"
	"github.com/begenov/real-estate/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, config config.RedisConfig) (*redis.Client, error) {
	redisInfo := fmt.Sprintf("%s:%s", config.Host, config.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     redisInfo,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("could not connect to Redis: %v", err)
	}

	return client, nil
}
