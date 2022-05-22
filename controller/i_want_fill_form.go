package controller

import (
	"encoding/json"
	"strconv"
	"time"
	"wechatGin/common"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
	"wechatGin/public"
	"wechatGin/services"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type OrderController struct {
}

func OrderRegister(group *gin.RouterGroup) {
	orderController := &OrderController{}
	group.GET("/wantFill", orderController.WantFillTheForm) // 开始填写目标问卷
	group.POST("/updateForm", orderController.UpdateForm)   // 更新参与者当前的填写记录
	group.POST("/commit", orderController.CommitDone)       // 参与者提交填写记录
}

func (order *OrderController) WantFillTheForm(c *gin.Context) {
	// 主要分为3种情况
	// 1. flag == 2 我已经提交了最终的结果了
	// 2. flag == 1 我已经提交过了，找到缓存继续填写
	// 3. flag == 0 我是第一次提交，那么找到表格最原始的并建立映射关系
	// 说一下我的想法当为0的是那么就会向数据库中插入一条记录说明我开始填写了
	// 可能出现大量的人开始同时填写，那么写入数据库的操作就放到消息队列中
	// 而出现1/2的情况的时候就是读的场景，读基本无所谓了
	// todo 拿到参数
	params := &dto.FormOrderInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 5001, err)
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 5000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	// todo 判断是否在数据库中记录
	var commitInfo *dao.Commit
	flag, str, err := commitInfo.IsExistDoingInfo(c, base.tx, params.Order, base.uid) // 并没有关心是否被自己/对方删除
	if err != nil {
		middleware.ResponseError(c, 5002, errors.New("数据库错误"))
		return
	}
	var resMsg dto.FormOrderOutput
	if flag == 2 { // 已经存在记录了且已经确定了最后的提交了
		// 即使后面被发布人移除了该记录那么也不能重新提交而是要发布人自己手动添加
		decompress, err := public.JsonDeTool(str)
		if err != nil {
			middleware.ResponseError(c, 5003, err)
			return
		}
		resMsg.Flag = 2
		resMsg.JsonMsg = string(decompress)
		middleware.ResponseSuccess(c, resMsg)
		return
	} else if flag == 1 { // 已经记录了所以去缓存中去找一下还有不有记录
		cache, _ := services.FindToCache(base.uid, params.Order)
		if cache != "" { // 找到了缓存那么就接着填写
			resMsg.Flag = 1
			resMsg.JsonMsg = cache
			middleware.ResponseSuccess(c, resMsg)
			return
		} // 如果为 "" 的话就是说要么错误要么没有了那么就开始查询数据库了
	}
	// todo 先去缓存中找有不有再去数据库中找
	var orderInfo *dao.Form
	fromJson := ""
	cache, _ := findCache(params.Order)
	if cache != "" {
		fromJson = cache
	} else {
		orderInfo, err = orderInfo.GetFormDetailByOrderId(c, base.tx, params.Order, false)
		if err != nil {
			middleware.ResponseError(c, 5004, err)
			return
		}
		decompress, err := public.JsonDeTool(orderInfo.FormInfos.FormJson)
		if err != nil {
			middleware.ResponseError(c, 5003, err)
			return
		}
		fromJson = string(decompress)
	}
	// todo 更新下最新的填表格时间
	services.MakeFormCacheTo(base.uid, params.Order)
	// todo 订单状态OK后如果是已经存在了记录
	if flag == 1 { // 如果到这了flag还是1的话那么就是没有找到，但是不用记录了
		resMsg.Flag = 1
		err = doCache(params.Order, fromJson)
		if err != nil {
			middleware.ResponseError(c, 5006, errors.New("服务器错误"))
			return
		}
		resMsg.JsonMsg = fromJson
		middleware.ResponseSuccess(c, resMsg)
		return
	}
	// todo 订单状态OK后但是是首次不存在记录的话
	toUid, err := orderInfo.GetUidFromOrder(c, base.tx, params.Order) // 拿到对谁进行填表
	if err != nil {
		middleware.ResponseError(c, 5007, err)
		return
	}
	// todo 异步加入消息队列里面进行创建
	err = services.PackInfo2QueueToCreateFillRecord(base.uid, strconv.Itoa(toUid), params.Order)
	if err != nil {
		middleware.ResponseError(c, 5008, errors.New("加入消息队列错误"))
		return
	}
	err = doCache(params.Order, fromJson) // 手动加入缓存中
	if err != nil {
		middleware.ResponseError(c, 5006, errors.New("服务器错误"))
		return
	}
	// todo 返回结果
	resMsg.Flag = 0
	resMsg.JsonMsg = fromJson
	middleware.ResponseSuccess(c, resMsg)
}

func (order *OrderController) UpdateForm(c *gin.Context) {
	params := &dto.FormUpdateInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 5001, err)
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 5000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	var orderInfo *dao.Form
	orderInfo, err = orderInfo.GetFormDetailByOrderId(c, base.tx, params.Order, false)
	if err != nil {
		middleware.ResponseError(c, 5004, err)
		return
	}
	// todo 对表单进行分析
	resInfo, err := ParasJson(params.FormJson)
	if err != nil {
		middleware.ResponseError(c, 5009, errors.New("分析表单错误"))
		return
	}
	// todo 订单号与压缩后的信息建立缓存
	err = services.MakeFormInfoCacheTo(base.uid, params.Order, public.JsonCoTool(resInfo)) // 此处更换了因为是多个人对同一个订单进行
	if err != nil {
		middleware.ResponseError(c, 5010, errors.New("更新记录发生错误"))
		return
	}
	returnInfo := dto.FormUpdateOutput{
		UpdateTime: strconv.Itoa(int(time.Now().Unix())),
		Msg:        "本次缓存更新成功",
	}
	middleware.ResponseSuccess(c, returnInfo)
}

func (order *OrderController) CommitDone(c *gin.Context) {
	params := &dto.FormUpdateInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 5001, err)
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 5000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	nickName, err := GetInfoByContext("nickName", c)
	if err != nil {
		middleware.ResponseError(c, 5998, errors.New("token解析错误"))
		return
	}
	// todo 判断时候存在创建记录/是否能够提交了
	tempCommit := &dao.Commit{}
	err = tempCommit.JudgeStatusIsBeWrittenAndCreated(base.tx, params.Order, base.uid)
	if err != nil {
		middleware.ResponseError(c, 5011, err)
		return
	}
	// todo 查询这个表单的详情
	var orderInfo *dao.Form
	orderInfo, err = orderInfo.GetFormDetailByOrderId(c, base.tx, params.Order, false)
	if err != nil {
		middleware.ResponseError(c, 5004, err)
		return
	}
	// todo 加入消息队列
	err = services.PackInfo2QueueToMakeRecord(params.FormJson, base.uid, params.Order, nickName)
	if err != nil {
		middleware.ResponseError(c, 5012, errors.New("加入消息队列错误"))
		return
	}
	// todo 删除uid_to中的缓存以及填表记录缓存
	deleteCacheCommit(params.Order, base.uid)
	// todo 包装websocket消息
	str, err := handlePackageInfo(c, params.Order, orderInfo.FormInfos.Title)
	if err != nil {
		middleware.ResponseError(c, 5013, errors.New("服务器错误"))
		return
	}
	common.WebsocketService.PushInfo(strconv.Itoa(orderInfo.Uid), str)
	middleware.ResponseSuccess(c, "记录成功")
}

func findCache(order string) (string, error) {
	jsonStr, err := services.GetFormDetail(order)
	if err != nil {
		return "", err
	}
	if jsonStr != "" { // 如果存在
		return jsonStr, nil
	}
	return "", nil
}

func doCache(order, jsonInfo string) error {
	err := services.MakeFormDetailCache(order, jsonInfo)
	if err != nil {
		return err
	}
	return nil
}

func handlePackageInfo(c *gin.Context, order, title string) (string, error) {
	var temp = dto.EachFromDoing{}
	avatar, err := GetInfoByContext("avatar", c)
	if err != nil {
		return "", err
	}
	nickName, err := GetInfoByContext("nickName", c)
	if err != nil {
		return "", err
	}
	temp.Nickname = nickName
	temp.AvatarUrl = avatar
	temp.OrderId = order
	temp.UpdateTime = time.Now().Format("2006-01-02 15:04")
	temp.Title = title
	marshal, err := json.Marshal(temp)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

// deleteCacheCommit 删除更新记录以及个人有序集合中的order
func deleteCacheCommit(orderID, uid string) error {
	if err := common.RedisConfPipline(func(c redis.Conn) {
		c.Send("DEL", uid+"_to_"+orderID)  // 删除不断更新的缓存
		c.Send("ZREM", uid+"_to", orderID) // 删除有序集合中的member
	}); err != nil {
		return err
	}
	return nil
}
