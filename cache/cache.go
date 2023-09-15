package cache

import (
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/go-redis/redis"
)

type RedisCache struct {
	red *redis.Client
}

var (
	log = logger.NewLogger("cache", logger.LevelDebug)
)

func Initialize(red *redis.Client) *RedisCache {
	return &RedisCache{red: red}
}

func (r *RedisCache) GetRedisClient() *redis.Client {
	return r.red
}
