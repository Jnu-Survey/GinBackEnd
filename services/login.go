package services

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/eddieivan01/nic"
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
	"wechatGin/common"
	"wechatGin/public"
)

// Code2Session 登录凭证校验拿到openid
func Code2Session(code string) (string, error) {
	aimUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%v&secret=%v&js_code=%v&grant_type=authorization_code", public.Appid, public.Secret, code)
	resp, err := nic.Get(aimUrl, nil)
	if err != nil {
		return "", errors.New("服务器错误")
	}
	buf := bytes.NewBuffer([]byte(resp.Text))
	jsonRes, err := simplejson.NewFromReader(buf)
	openid := jsonRes.Get("openid").MustString()
	errCode := jsonRes.Get("errcode").MustInt()
	if errCode != 0 {
		errMsg := jsonRes.Get("errmsg").MustString()
		return "", errors.New(errMsg)
	}
	return openid, nil
}

// JudgeTime 判断时间差是否大于12天
func JudgeTime(update time.Time) bool {
	TimeLocation, _ := time.LoadLocation("Asia/Shanghai")
	nowTime := time.Now().In(TimeLocation)
	if nowTime.Sub(update).Hours() > 24*12 { // 如果大于12天的话
		return true
	}
	return false
}

// MakeTokenCache 为Token添加Cache
func MakeTokenCache(id, token string) error {
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("SET", token, id)
		c.Send("EXPIRE", token, 86400*3-public.RandomInt(10)*50) // 防止缓存雪崩
	}); err != nil {
		return err
	}
	return nil
}

func TokenOpenId(openid string) string {
	timestamp := time.Now().UnixNano()
	wantEn := fmt.Sprintf("%v_%v", timestamp, openid)
	return public.Base64Encoding(public.CBCEncrypt([]byte(wantEn), []byte(public.TempTokenKey), public.TempTokenIv))
}

func GetOpenIdFormToken(token string) (res string, curErr error) {
	defer func() {
		if err := recover(); err != nil {
			curErr = errors.New("解析错误")
			return
		}
	}()
	tokenByte := public.Base64Decoding(token)
	token = string(public.CBCDecrypt(tokenByte, []byte(public.TempTokenKey), public.TempTokenIv))
	res = strings.Split(token, "_")[0]
	return
}
