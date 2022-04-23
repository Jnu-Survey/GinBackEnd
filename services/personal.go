package services

import (
	"errors"
	"time"
	"wechatGin/common"
)

// GetStillDoing 获取仍然在有序集合中的订单
func GetStillDoing(uid string) ([]string, error) {
	TimeLocation, _ := time.LoadLocation("Asia/Shanghai")
	nowTime := time.Now().In(TimeLocation).Unix()
	ago := nowTime - 86400*3
	do, err := common.RedisConfDo("ZREVRANGEBYSCORE", uid+"_from", nowTime, ago)
	if err != nil {
		return nil, err
	}
	interfaceArr, ok := do.([]interface{})
	if !ok {
		return nil, errors.New("断言错误")
	}
	var ans []string
	for k, _ := range interfaceArr {
		each, ok := interfaceArr[k].([]byte)
		if !ok {
			return nil, errors.New("断言错误")
		}
		ans = append(ans, string(each))
	}
	return ans, nil
}
