package services

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"wechatGin/common"
	"wechatGin/public"
)

// GetHomeCache 获取首页缓存
func GetHomeCache(key string) (string, error) {
	do, err := common.RedisConfDo("GET", key)
	if err != nil {
		return "", err
	}
	if do == nil { // 缓存不存在
		return "", nil
	}
	value, ok := do.([]byte)
	if !ok {
		return "", errors.New("断言失败")
	}
	return string(value), nil
}

// MakeHomeCache 为首页添加缓存
func MakeHomeCache(key, value string) error {
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("SET", key, value)
		c.Send("EXPIRE", key, 86400-public.RandomInt(10)*50) // 防止缓存雪崩
	}); err != nil {
		return err
	}
	return nil
}
