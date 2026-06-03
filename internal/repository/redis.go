package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client  *redis.Client
	baseTTL int
}

func NewRedisRepo(client *redis.Client, baseTTL int) *RedisRepo {
	return &RedisRepo{client: client, baseTTL: baseTTL}
}

func (r *RedisRepo) Get(code string) (string, bool) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, "short:"+code).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

func (r *RedisRepo) Set(code, originalURL string, ttl int) error {
	ctx := context.Background()
	if ttl <= 0 {
		ttl = r.baseTTL
	}
	return r.client.Set(ctx, "short:"+code, originalURL, time.Duration(ttl)*time.Second).Err()
}

func (r *RedisRepo) DeleteCache(code string) error {
	ctx := context.Background()
	return r.client.Del(ctx, "short:"+code).Err()
}
