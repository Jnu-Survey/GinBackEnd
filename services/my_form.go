package services

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
	"wechatGin/common"
	"wechatGin/dao"
	"wechatGin/public"
)

// GetCacheInfo 通过订单数组去拿到详细信息
func GetCacheInfo(uid string, forms []dao.Form) ([]dao.Form, error) {
	// todo 因为使用mget只要有一个找不到就返回nil
	var ansForms []dao.Form
	for k, _ := range forms {
		res, err := common.RedisConfDo("get", forms[k].RandomId+"_"+uid) // 修复是找某个人的
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
			decompress, err := public.JsonDeTool(string(judge))
			if err != nil {
				return nil, err
			}
			info := strings.Split(string(decompress), public.SplitSymbol) // 使用特殊的标记进行切割
			if len(info) != 4 {
				return nil, errors.New("切割错误")
			}
			timeStr, title, tip, json := info[0], info[1], info[2], info[3]
			one.FormInfos.Title = title
			one.FormInfos.Tip = tip
			one.FormInfos.FormJson = json
			timeNum, _ := strconv.Atoi(timeStr)
			one.UpdatedAt = time.Unix(int64(timeNum), 0)
		}
		ansForms = append(ansForms, one)
	}
	return ansForms, nil
}
