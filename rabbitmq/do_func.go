package rabbitmq

import (
	"errors"
	"github.com/bitly/go-simplejson"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
	"wechatGin/common"
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
	// todo 第一个函数用于创建者创建表单
	var func1toFormIdWithUid = func(msg string) (err error) {
		defer func() {
			if err := recover(); err != nil {
				err = errors.New("崩溃错误")
			}
		}()
		msgList := strings.Split(msg, public.SplitSymbol)
		uid, orderId := msgList[0], msgList[1]
		dsn := common.RabbitDsn // 与配置中相通
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
	// todo 第二个函数用于提交者最终提交写入数据库
	var func2ToMakeRecord = func(msg string) (err error) {
		defer func() {
			if err := recover(); err != nil {
				err = errors.New("崩溃错误")
			}
		}()
		msgList := strings.Split(msg, public.SplitSymbol)
		uid, orderId, nickName, json := msgList[0], msgList[1], msgList[2], msgList[3]
		dsn := common.RabbitDsn // 与配置中相通
		tx, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}
		// todo 添加到mongo与mysql
		err = insertMongoAndMysql(tx, json, uid, orderId, nickName)
		if err != nil {
			return
		}
		return nil
	}
	// todo 第三个函数用于提交者创建记录我进入这张表进行填写了
	var func3ToMakeFillRecord = func(msg string) (err error) {
		defer func() {
			if err := recover(); err != nil {
				err = errors.New("崩溃错误")
			}
		}()
		msgList := strings.Split(msg, public.SplitSymbol)
		from, to, order := msgList[0], msgList[1], msgList[2]
		dsn := common.RabbitDsn // 与配置中相通
		tx, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}
		// todo 添加到填写记录
		fromInt, err := strconv.Atoi(from)
		toInt, err := strconv.Atoi(to)
		if err != nil {
			return
		}
		var addInfo = &dao.Commit{
			FromUid: fromInt,
			ToUid:   toInt,
			OrderId: order,
		}
		addInfo, err = addInfo.AddFrom2To(tx, addInfo)
		if err != nil {
			return
		}
		return nil
	}
	return []DealFunc{func1toFormIdWithUid, func2ToMakeRecord, func3ToMakeFillRecord}
}

func Consume() {
	rabbitmq, _ := NewRabbitMQSimple("Simple")
	handle := func(msg string, dealList ...DealFunc) (curErr error) {
		defer func() {
			if err := recover(); err != nil {
				curErr = errors.New("崩溃错误")
			}
		}()
		msgList := strings.Split(msg, public.SplitSymbol)
		// todo 解析当前信息是要第几个函数
		which := msgList[0]
		whichNum, err := strconv.Atoi(which) // 拿到对应的第几个
		if err != nil {
			curErr = err
			return
		}
		// todo 找到对应的函数处理
		for k, f := range dealList {
			if whichNum == k && k == 0 { // 当匹配上第一个函数
				err = f(msgList[1] + public.SplitSymbol + msgList[2]) // uid%_%orderId
				if err != nil {
					curErr = err
					return
				}
			} else if whichNum == k && k == 1 {
				err = f(msgList[1] + public.SplitSymbol + msgList[2] + public.SplitSymbol + msgList[3] + public.SplitSymbol + msgList[4]) // uid%_%orderId%_%nickName%_%json
				if err != nil {
					curErr = err
					return
				}
			} else if whichNum == k && k == 2 {
				err = f(msgList[1] + public.SplitSymbol + msgList[2] + public.SplitSymbol + msgList[3]) // from%_%to%_%order
				if err != nil {
					curErr = err
					return
				}
			}
		}
		return nil
	}
	rabbitmq.ConsumeSimple(common.RabbitMysqlConsumeNum, makeDealFuncList(), handle)
}

func insertMongoAndMysql(tx *gorm.DB, json, uid, order, nickName string) error {
	tempCommit := &dao.Commit{}
	// todo 开启事务
	tx = tx.Begin()
	// todo 加入Mongo记录去拿ID
	mongoConnect, err := common.NewMongoDbPool()
	if err != nil {
		return err
	}
	json, err = HandleJsonInfo(json, order, nickName, uid)
	if err != nil {
		return errors.New("处理字段错误")
	}
	dbStr, err := mongoConnect.InsertToDb(json)
	if err != nil {
		return err
	}
	// todo 加入Mysql记录
	tempCommit, err = tempCommit.RewriteCommit(tx, order, uid, dbStr) // 修改状态
	if err != nil {
		tx.Rollback()
		return err
	}
	res := public.JsonCoTool(json) // 压缩
	formInfo := &dao.CommitInfo{FormJson: res, Out: tempCommit.Id}
	formInfo, err = formInfo.RecordJson(tx, formInfo)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// HandleJsonInfo 加工下Json
func HandleJsonInfo(jsonStr, order, nickName, fromUid string) (string, error) {
	j, err := simplejson.NewJson([]byte(jsonStr))
	if err != nil {
		return "", err
	}
	j.Del("order_id_key")
	j.Del("nick_name")
	j.Del("update_time")
	j.Del("from_uid")
	j.Set("order_id_key", order)
	j.Set("nick_name", nickName)
	j.Set("from_uid", fromUid)
	j.Set("update_time", time.Now().Format("2006-01-02 15:04"))
	marshalJSON, err := j.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(marshalJSON), nil
}
