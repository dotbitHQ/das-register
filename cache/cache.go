package cache

import (
	"github.com/go-redis/redis"
	"github.com/scorpiotzh/mylog"
)

type RedisCache struct {
	red *redis.Client
}

var (
	log = mylog.NewLogger("cache", mylog.LevelDebug)
)

func Initialize(red *redis.Client) *RedisCache {
	return &RedisCache{red: red}
}

func (r *RedisCache) GetRedisClient() *redis.Client {
	return r.red
}
