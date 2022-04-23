package services

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"strconv"
	"time"
	"wechatGin/common"
	"wechatGin/public"
	"wechatGin/rabbitmq"
)

func RandomFormId(token string) string {
	TimeLocation, _ := time.LoadLocation("Asia/Shanghai")
	detail := time.Now().In(TimeLocation).UnixNano()
	detailStr := strconv.Itoa(int(detail))
	detailStr = detailStr[len(detailStr)-6:]
	rand.Seed(detail)
	res := ""
	other1, other2 := rand.Intn(4), rand.Intn(8)+4
	res = detailStr + res + token[other1:other2]
	return public.HashSHA256Encoding(res)[:16]
}

//MakeFormCache 添加缓存_score是时间_记录的是随机表单号
func MakeFormCache(uid, randomKey string) error {
	TimeLocation, _ := time.LoadLocation("Asia/Shanghai")
	nowTime := time.Now().In(TimeLocation).Unix()
	ago := nowTime - 86400*3
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("ZADD", uid+"_from", nowTime, randomKey)
		c.Send("zremrangebyscore", uid+"_from", 1648742400, ago) // 从顺便把3天前也删除了
	}); err != nil {
		return err
	}
	return nil
}

// PackInfo2Queue 把消息送过去
func PackInfo2Queue(orderID string, uid int) error {
	res := fmt.Sprintf("%v%v%v%v%v", "0", public.SplitSymbol, uid, public.SplitSymbol, orderID)
	err := rabbitmq.PushStrToAimQueue(res)
	if err != nil {
		return err
	}
	return nil
}

// MakeFormInfoCache 将订单号与JSON建立缓存
func MakeFormInfoCache(orderID, jsonInfo string) error {
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("SETNX", orderID, jsonInfo)
		c.Send("SET", orderID, jsonInfo)
		c.Send("EXPIRE", orderID, 86400*3-3600)
	}); err != nil {
		return err
	}
	return nil
}

// DeleteCacheDone 删除表单记录以及个人有序集合中的order
func DeleteCacheDone(orderInfo, orderID, uid string) error {
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("DEL", orderInfo)             // 删除不断更新的缓存
		c.Send("ZREM", uid+"_from", orderID) // 删除有序集合中的member
	}); err != nil {
		return err
	}
	return nil
}
