package services

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
	"wechatGin/common"
	"wechatGin/public"
	"wechatGin/rabbitmq"
)

// MakeFormDetailCache 已经完成的表单打上缓存
func MakeFormDetailCache(orderID, jsonInfo string) error {
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("SETNX", orderID, jsonInfo)
		c.Send("EXPIRE", orderID, 86400)
	}); err != nil {
		return err
	}
	return nil
}

//MakeFormCacheTo 该用户正在填写别人的表单
func MakeFormCacheTo(uid, randomKey string) error {
	TimeLocation, _ := time.LoadLocation("Asia/Shanghai")
	nowTime := time.Now().In(TimeLocation).Unix()
	ago := nowTime - 86400*3
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("ZREM", uid+"_to", randomKey)                   // 有的话先把上一个删除了
		c.Send("ZADD", uid+"_to", nowTime, randomKey)          // 相当于记录上一次进入的时间
		c.Send("zremrangebyscore", uid+"_to", 1648742400, ago) // 从顺便把3天前也删除了
	}); err != nil {
		return err
	}
	return nil
}

// GetFormDetail 拿到Detail缓存
func GetFormDetail(order string) (string, error) {
	res, err := common.RedisConfDo("get", order) // 修复是找某个人的
	if res == nil {
		return "", nil
	}
	if err != nil {
		return "", nil
	}
	jsonStr, ok := res.([]byte)
	if !ok {
		return "", errors.New("断言错误")
	}
	decompress, err := public.JsonDeTool(string(jsonStr))
	if err != nil {
		return "", errors.New("解压错误")
	}
	return string(decompress), nil
}

// MakeFormInfoCacheTo 填写表单的人记录填写记录（用于更新的）
func MakeFormInfoCacheTo(uid, orderID, jsonInfo string) error {
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("SETNX", uid+"_to_"+orderID, jsonInfo)
		c.Send("SET", uid+"_to_"+orderID, jsonInfo)
		c.Send("EXPIRE", orderID, 86400*3-3600)
	}); err != nil {
		return err
	}
	return nil
}

// FindToCache 寻找填表人的缓存
func FindToCache(uid, orderId string) (string, error) {
	do, err := common.RedisConfDo("GET", uid+"_to_"+orderId)
	if do == nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	judge, ok := do.([]byte)
	if !ok {
		return "", errors.New("断言错误")
	}
	decompress, err := public.JsonDeTool(string(judge))
	if err != nil {
		return "", err
	}
	list := strings.Split(string(decompress), public.SplitSymbol)
	if len(list) != 4 {
		return "", errors.New("缓存错误")
	}
	return list[3], nil
}

// PackInfo2QueueToMakeRecord 把填写者的消息送入消息队列中
func PackInfo2QueueToMakeRecord(json, uid, order, nickName string) error {
	res := fmt.Sprintf("%v%v%v%v%v%v%v%v%v", "1", public.SplitSymbol, uid, public.SplitSymbol, order, public.SplitSymbol, nickName, public.SplitSymbol, json)
	err := rabbitmq.PushStrToAimQueue(res)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// PackInfo2QueueToCreateFillRecord 当有人第一次填写表单的时候 尽快返回表单
func PackInfo2QueueToCreateFillRecord(from, to, order string) error {
	res := fmt.Sprintf("%v%v%v%v%v%v%v", "2", public.SplitSymbol, from, public.SplitSymbol, to, public.SplitSymbol, order)
	err := rabbitmq.PushStrToAimQueue(res)
	if err != nil {
		return err
	}
	return nil
}
