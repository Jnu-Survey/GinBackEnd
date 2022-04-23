package services

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"wechatGin/common"
	"wechatGin/dao"
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

//MakeFormCache 添加缓存
func MakeFormCache(uid, randomKey string) error {
	TimeLocation, _ := time.LoadLocation("Asia/Shanghai")
	nowTime := time.Now().In(TimeLocation).Unix()
	ago := nowTime - 86400*3
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("ZADD", uid, nowTime, randomKey)
		c.Send("zremrangebyscore", uid, 1648742400, ago) // 从顺便把3天前也删除了
	}); err != nil {
		return err
	}
	return nil
}

// PackInfo2Queue 把消息送过去
func PackInfo2Queue(orderID string, uid int) error {
	res := fmt.Sprintf("%v_%v_%v", "0", uid, orderID)
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
		c.Send("EXPIRE", orderID, 86400*3-3600)
	}); err != nil {
		return err
	}
	return nil
}

// DeleteCacheDone 删除已经记录的缓存
func DeleteCacheDone(orderID, uid string) error {
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("DEL", orderID)       // 删除不断更新的缓存
		c.Send("ZREM", uid, orderID) // 删除有序集合中的member
	}); err != nil {
		return err
	}
	return nil
}

func GetCacheInfo(forms []dao.Form) ([]dao.Form, error) {
	// todo 因为使用mget只要有一个找不到就返回nil
	var ansForms []dao.Form
	for k, _ := range forms {
		res, err := common.RedisConfDo("get", forms[k].RandomId)
		if err != nil {
			return nil, err
		}
		one := forms[k]
		if res == nil {
			continue
		} else {
			judge, ok := res.([]byte)
			if !ok {
				return nil, errors.New("断言错误")
			}
			decompress, err := public.JsonDecompress(public.Base64Decoding(string(judge)))
			if err != nil {
				return nil, err
			}
			info := strings.Split(string(decompress), "_")
			if len(info) != 3 {
				return nil, errors.New("切字符串错误")
			}
			title, tip := info[0], info[1]
			one.FormInfos.Title = title
			one.FormInfos.Tip = tip
		}
		ansForms = append(ansForms, one)
	}
	return ansForms, nil
}
