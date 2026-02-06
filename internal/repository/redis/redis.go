package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type IRedisRepo interface {
	Save(ctx context.Context, key string, value interface{}, expires time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key ...string) error
}

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {

	return &RedisRepo{client: client}
}

func (r *RedisRepo) Save(ctx context.Context, key string, value interface{}, expires time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, jsonValue, expires).Err()
}

func (r *RedisRepo) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisRepo) Delete(ctx context.Context, key ...string) error {
	return r.client.Del(ctx, key...).Err()
}
