package cache

import (
	"fmt"
	"time"
)

//register:coupon:BQVXJ7LC
func (r *RedisCache) GetCouponLockWithRedis(coupon string, expiration time.Duration) error {
	key := fmt.Sprintf("register:coupon:%s", coupon)
	ret := r.red.SetNX(key, "", expiration)
	if err := ret.Err(); err != nil {
		return fmt.Errorf("get coupon lock: redis set nx-->%s", err.Error())
	}
	ok := ret.Val()
	if !ok {
		return fmt.Errorf("get coupon lock error")
	}
	return nil
}
func (r *RedisCache) DeleteCouponLockWithRedis(coupon string) error {
	key := fmt.Sprintf("register:coupon:%s", coupon)
	return r.red.Del(key).Err()
}
