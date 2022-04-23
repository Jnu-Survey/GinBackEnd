package rabbitmq

import (
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"wechatGin/dao"
	"wechatGin/public"
)

type DealFunc func(msg string) error

// PushStrToAimQueue 将字符串送入消息队列
func PushStrToAimQueue(aim string) error {
	rabbitmq, err := NewRabbitMQSimple("Simple")
	if err != nil {
		return errors.New("创建rabbitmq的实例错误")
	}
	err = rabbitmq.PublishSimple(aim)
	if err != nil {
		return errors.New("向通道里面写入数据错误")
	}
	return nil
}

// 函数处理这里必须著名先后顺序
func makeDealFuncList() []DealFunc {
	// todo 创建第一个处理函数
	var func1toFormIdWithUid = func(msg string) (err error) {
		defer func() {
			if err := recover(); err != nil {
				err = errors.New("崩溃错误")
			}
		}()
		msgList := strings.Split(msg, "_")
		uid, orderId := msgList[0], msgList[1]
		dsn := public.RabbitDsn // 与配置中相通
		tx, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}
		curUid, err := strconv.Atoi(uid)
		if err != nil {
			return
		}
		var formInfo = &dao.Form{
			RandomId: orderId,
			Uid:      curUid,
		}
		formInfo, err = formInfo.AddForm2Uid(tx, formInfo)
		if err != nil {
			return
		}
		return nil
	}
	// todo 等待创建第二个函数
	return []DealFunc{func1toFormIdWithUid}
}

func Consume() {
	// 申请到对应的管道
	rabbitmq, _ := NewRabbitMQSimple("Simple")
	handle := func(msg string, dealList ...DealFunc) (curErr error) {
		defer func() {
			if err := recover(); err != nil {
				curErr = errors.New("崩溃错误")
			}
		}()
		msgList := strings.Split(msg, "_")
		which, uid, orderId := msgList[0], msgList[1], msgList[2]
		// todo 解析当前信息是要第几个函数
		whichNum, err := strconv.Atoi(which) // 拿到对应的第几个
		if err != nil {
			curErr = err
			return
		}
		// todo 找到对应的函数处理
		for k, f := range dealList {
			if k == whichNum { // 处理对应的函数
				err = f(uid + "_" + orderId)
				if err != nil {
					curErr = err
					return
				}
			}
		}
		return nil
	}
	rabbitmq.ConsumeSimple(5, makeDealFuncList(), handle) // 5个写锁
}
