package cache

import (
	"errors"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"strings"
	"time"
)

func (r *RedisCache) getSearchLimitKey(chainType common.ChainType, address, action string) string {
	key := fmt.Sprintf("search:limit:%s:%d:%s", action, chainType, address)
	return strings.ToLower(key)
}

var ErrDistributedLockPreemption = errors.New("distributed lock preemption")

func (r *RedisCache) LockWithRedis(chainType common.ChainType, address, action string, expiration time.Duration) error {
	log.Info("LockWithRedis:", chainType, address, action)
	key := r.getSearchLimitKey(chainType, address, action)
	ret := r.red.SetNX(key, "", expiration)
	if err := ret.Err(); err != nil {
		return fmt.Errorf("redis set order nx-->%s", err.Error())
	}
	ok := ret.Val()
	log.Info("LockWithRedis:", ok)
	if !ok {
		return ErrDistributedLockPreemption
	}
	return nil
}
