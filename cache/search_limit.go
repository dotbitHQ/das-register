package cache

import (
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"strings"
	"time"
)

func (r *RedisCache) getSearchLimitKey(chainType common.ChainType, address, action string) string {
	key := fmt.Sprintf("search:limit:%s:%d:%s", action, chainType, address)
	return strings.ToLower(key)
}

func (r *RedisCache) SetSearchLimit(chainType common.ChainType, address, action string) error {
	if r.red == nil {
		return fmt.Errorf("redis is nil")
	}
	key := r.getSearchLimitKey(chainType, address, action)
	return r.red.Set(key, 1, time.Millisecond*600).Err()
}

func (r *RedisCache) SearchLimitExist(chainType common.ChainType, address, action string) bool {
	if r.red == nil {
		return false
	}
	key := r.getSearchLimitKey(chainType, address, action)
	res, _ := r.red.Exists(key).Result()
	if res == 1 {
		return true
	}
	return false
}
