package cache

import (
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"strings"
	"time"
)

func (r *RedisCache) getRegisterLimitLockWithRedisKey(chainType common.ChainType, address, action, account string) string {
	key := fmt.Sprintf("register:lock:%s:%s:%d:%s", account, action, chainType, address)
	return strings.ToLower(key)
}

func (r *RedisCache) RegisterLimitLockWithRedis(chainType common.ChainType, address, action, account string, expiration time.Duration) error {
	key := r.getRegisterLimitLockWithRedisKey(chainType, address, action, account)
	ret := r.red.SetNX(key, "", expiration)
	if err := ret.Err(); err != nil {
		return fmt.Errorf("redis set nx-->%s", err.Error())
	}
	ok := ret.Val()
	if !ok {
		return ErrDistributedLockPreemption
	}
	return nil
}
