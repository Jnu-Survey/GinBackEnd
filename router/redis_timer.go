package router

import (
	"strings"
	"time"
	"wechatGin/common"
	"wechatGin/public"
)

type RedisTimer struct {
	exitSingle chan int // 退出信号
	info       string   // 绑定的信号
}

func NewRedisTimer() *RedisTimer {
	one := &RedisTimer{
		exitSingle: make(chan int, 1),
		info:       "",
	}
	return one
}

func (r *RedisTimer) RedisTimerStart() {
	// todo 创建一个定时器用于服务端心跳
	noticeTicker := time.NewTicker(time.Second * 60)     // 每60秒
	renewTicker := time.NewTicker(time.Second * 60 * 60) // 每1小时
	// todo 开启协程
	go func() {
		for {
			select {
			case _ = <-r.exitSingle:
				return
			case <-noticeTicker.C: // 从redis拿数据
				resInfo := r.GetRedisListPopInfo()
				if len(resInfo) != 0 { // 进行处理
					r.HandleEmail(resInfo)
				}
			case <-renewTicker.C: // 刷新
				common.BloomFilterService = common.NewBloomFilter()
			}
		}
	}()
}

func (r *RedisTimer) RedisTimerStop() {
	r.exitSingle <- 1
}

func (r *RedisTimer) GetRedisListPopInfo() (resInfo []string) {
	do, err := common.RedisConfDo("RPOP", "email_timer")
	if err != nil {
		return
	}
	if do == nil {
		return
	}
	info, ok := do.([]byte)
	if !ok {
		return
	}
	splitStr := strings.Split(string(info), public.SplitSymbol)
	if len(splitStr) != 4 {
		return
	}
	uid, connection, title, infoStr := splitStr[0], splitStr[1], splitStr[2], splitStr[3]
	resInfo = append(resInfo, uid)
	resInfo = append(resInfo, connection)
	resInfo = append(resInfo, title)
	resInfo = append(resInfo, infoStr)
	return
}

func (r *RedisTimer) HandleEmail(resInfo []string) {
	body := common.WaterBody(common.EmailUser, common.AdminEmail, resInfo)
	common.SendMailUsingTLS(common.EmailUser, common.AdminEmail, []byte(body))
}
