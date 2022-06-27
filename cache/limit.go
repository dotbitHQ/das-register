package cache

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"strings"
	"time"
)

func (r *RedisCache) getApiLimitKey(chainType common.ChainType, address, action string) string {
	key := fmt.Sprintf("api:limit:%s:%d:%s", action, chainType, address)
	return strings.ToLower(key)
}

func (r *RedisCache) SetApiLimit(chainType common.ChainType, address, action string) error {
	if r.red == nil {
		return fmt.Errorf("redis is nil")
	}
	key := r.getApiLimitKey(chainType, address, action)
	return r.red.Set(key, 1, time.Minute*3).Err()
}

func (r *RedisCache) ApiLimitExist(chainType common.ChainType, address, action string) bool {
	if r.red == nil {
		return false
	}
	key := r.getApiLimitKey(chainType, address, action)
	res, _ := r.red.Exists(key).Result()
	if res == 1 {
		return true
	}
	return false
}

// account limit

func (r *RedisCache) getAccountLimitKey(account string) string {
	key := fmt.Sprintf("limit:%s", account)
	return strings.ToLower(key)
}

func (r *RedisCache) SetAccountLimit(account string, expiration time.Duration) error {
	if r.red == nil {
		return fmt.Errorf("redis is nil")
	}
	key := r.getAccountLimitKey(account)
	return r.red.Set(key, 1, expiration).Err()
}

func (r *RedisCache) AccountLimitExist(account string) bool {
	if r.red == nil {
		return false
	}
	key := r.getAccountLimitKey(account)
	res, _ := r.red.Exists(key).Result()
	if res == 1 {
		return true
	}
	return false
}

// register limit
func (r *RedisCache) getRegisterLimitKey(chainType common.ChainType, address, account, action string) string {
	key := fmt.Sprintf("register:limit:%s:%d:%s:%s", account, chainType, address, action)
	return strings.ToLower(key)
}

func (r *RedisCache) SetRegisterLimit(chainType common.ChainType, address, account, action string, expiration time.Duration) error {
	if r.red == nil {
		return fmt.Errorf("redis is nil")
	}
	key := r.getRegisterLimitKey(chainType, address, account, action)
	return r.red.Set(key, 1, expiration).Err()
}

func (r *RedisCache) RegisterLimitExist(chainType common.ChainType, address, account, action string) bool {
	if r.red == nil {
		return false
	}
	key := r.getRegisterLimitKey(chainType, address, account, action)
	res, _ := r.red.Exists(key).Result()
	if res == 1 {
		return true
	}
	return false
}
