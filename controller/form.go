package controller

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
	"wechatGin/common"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
	"wechatGin/public"
	"wechatGin/services"
)

type FormController struct {
}

func FormRegister(group *gin.RouterGroup) {
	formController := &FormController{}
	group.GET("/getFormId", formController.GetFormId)
	group.POST("/tempUpdate", formController.TempUpdate)
	group.POST("/formDone", formController.FormDone)
	group.GET("/formOutDone", formController.GetFormOutListDone)
	group.GET("/formDoing", formController.GetFormOutListDoing)
}

// GetFormId godoc
// @Summary 获取创建订单ID编号
// @Description 获取创建订单ID编号
// @Tags 业务
// @ID /form/getFormId
// @Accept  json
// @Produce  json
// @Param token query string true "通行证"
// @Success 200 {object} middleware.Response{data=dto.FormRegisteredInput} "success"
// @Router /form/getFormId [get]
func (form *FormController) GetFormId(c *gin.Context) {
	// 进行了通行证与数量的鉴权
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3001, err) // 3001 参数不正确
		return
	}
	// todo 通过缓存来找人
	uid, err := GetUidByContext(c)
	if err != nil {
		middleware.ResponseError(c, 3999, err) // 上下文获取UID错误
		return
	}
	// todo 添加判断数量是否过关
	flag, err := JudgeNumOneHas(uid)
	if err != nil {
		middleware.ResponseError(c, 3015, err)
		return
	}
	if !flag {
		middleware.ResponseError(c, 3016, errors.New("正在创建表单超过规定数量"))
		return
	}
	// todo 此时生成随机ID 解析Token找到对应的人
	randomToken := services.RandomFormId(params.Token)
	// todo 把uid与随机表单号写在有序集合缓存里面
	err = services.MakeFormCache(uid, randomToken)
	if err != nil {
		middleware.ResponseError(c, 3004, errors.New("写入缓存错误")) // 3004 数据库连接失败
		return
	}
	// todo 将写入数据库的操作写入消息队列
	intUId, _ := strconv.Atoi(uid)
	err = services.PackInfo2Queue(randomToken, intUId)
	if err != nil {
		middleware.ResponseError(c, 3005, errors.New("消息队列发生错误")) // 3005 消息队列发生错误
		return
	}
	// todo 返回随机ID
	res := dto.FormRegisteredOutput{
		RandomId: randomToken,
		Msg:      "创建表单成功，已经加入消息队列",
	}
	middleware.ResponseSuccess(c, res)
}

// TempUpdate godoc
// @Summary 更新记录缓存
// @Description 更新记录缓存
// @Tags 业务
// @ID /form/tempUpdate
// @Accept  json
// @Produce  json
// @Param body body dto.FormUpdateInput true "body"
// @Success 200 {object} middleware.Response{data=dto.FormUpdateOutput} "success"
// @Router /form/tempUpdate [post]
func (form *FormController) TempUpdate(c *gin.Context) {
	params := &dto.FormUpdateInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3001, err) // 3001 参数不正确
		return
	}
	// todo 对表单进行分析
	resInfo, err := ParasJson(params.FormJson)
	if err != nil {
		middleware.ResponseError(c, 3012, err)
		return
	}
	// todo 进行压缩
	res := public.Base64Encoding(public.JsonCompress([]byte(resInfo)))
	// todo 订单号与压缩后的信息建立缓存
	err = services.MakeFormInfoCache(params.Order, res)
	if err != nil {
		middleware.ResponseError(c, 3006, errors.New("更新记录发生错误")) // 3006 缓存的问题
		return
	}
	returnInfo := dto.FormUpdateOutput{
		UpdateTime: strconv.Itoa(int(time.Now().Unix())),
		Msg:        "本次缓存更新成功",
	}
	middleware.ResponseSuccess(c, returnInfo)
}

// FormDone godoc
// @Summary 完成表单提交
// @Description 完成表单提交
// @Tags 业务
// @ID /form/formDone
// @Accept  json
// @Produce  json
// @Param body body dto.FormUpdateInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /form/formDone [post]
func (form *FormController) FormDone(c *gin.Context) {
	params := &dto.FormFinalInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3001, err) // 3001 参数不正确
		return
	}
	// todo 通过缓存来找人
	uid, err := GetUidByContext(c)
	if err != nil {
		middleware.ResponseError(c, 3999, err) // 上下文获取UID错误
		return
	}
	// todo 直接保存到Mysql因为不需要分析只需要给前端进行渲染就好了
	// 事务：判断主表是不是存在；是不是能够修改；修改状态；从表加入数据
	res := public.Base64Encoding(public.JsonCompress([]byte(params.FormJson)))
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误")) // 10000 池子不通畅
		return
	}
	// todo 判断状态
	tempForm := &dao.Form{}
	tempForm, err = tempForm.IsExistAndJudgeStatus(c, tx, params.Order, tempForm)
	if err != nil {
		middleware.ResponseError(c, 3007, errors.New("服务器错误")) // 3007 要么找不到要么不能修改
		return
	}
	// todo 开启事务
	tx.Begin()
	err = tempForm.RewriteStatus(c, tx, tempForm)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		middleware.ResponseError(c, 3008, errors.New("服务器错误")) // 3008
		return
	}
	formInfo := &dao.FormInfo{FormJson: res, Title: params.FormTitle, Tip: params.FormTip, Out: tempForm.Id}
	formInfo, err = formInfo.RecordJson(c, tx, formInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3013, errors.New("服务器错误")) // 3013 从表写入错误
		return
	}
	tx.Commit()
	// todo 删除缓存
	err = services.DeleteCacheDone(params.Order, uid)
	if err != nil {
		middleware.ResponseError(c, 3009, errors.New("服务器错误")) // 3009 删除key的问题
		return
	}
	middleware.ResponseSuccess(c, "记录成功")
}

// GetFormOutListDone godoc
// @Summary 获取我发出的订单已经填写好了的
// @Description 获取我发出的订单已经填写好了的
// @Tags 业务
// @ID /form/formOutDone
// @Accept  json
// @Produce  json
// @Param token query string true "通行证"
// @Success 200 {object} middleware.Response{data=dto.FormDoneInfo} "success"
// @Router /form/formOutDone [get]
func (form *FormController) GetFormOutListDone(c *gin.Context) {
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3001, err) // 3001 参数不正确
		return
	}
	// todo 通过缓存来找人
	uid, err := GetUidByContext(c)
	if err != nil {
		middleware.ResponseError(c, 3999, err) // 上下文获取UID错误
		return
	}
	// todo 针对数据库进行联合查询
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误")) // 10000 池子不通畅
		return
	}
	// todo 查询已经填写好了的
	formStruct := &dao.Form{}
	forms, err := formStruct.GetAllDoneInfo(c, tx, uid)
	if err != nil {
		middleware.ResponseError(c, 3010, errors.New("服务器错误")) // 3010 查询错误
		return
	}
	// todo 包装记录
	if forms == nil { // 不存在记录
		middleware.ResponseSuccess(c, dto.FormDoneInfo{Msg: "没查到相关记录", Total: 0, Info: nil})
		return
	}
	middleware.ResponseSuccess(c, PackageReturnFormInfo(forms, 1))
}

// GetFormOutListDoing godoc
// @Summary 获取我正在填写的表单
// @Description 获取我正在填写的表单
// @Tags 业务
// @ID /form/formDoing
// @Accept  json
// @Produce  json
// @Param token query string true "通行证"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /form/formDoing [get]
func (form *FormController) GetFormOutListDoing(c *gin.Context) {
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3001, err) // 3001 参数不正确
		return
	}
	// todo 缓存上下文查找
	uid, err := GetUidByContext(c)
	if err != nil {
		middleware.ResponseError(c, 3999, err) // 上下文获取UID错误
		return
	}
	// todo 这里的查询是有SQL优化1.索引 2.部分字段
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误")) // 10000 池子不通畅
		return
	}
	formStruct := &dao.Form{}
	forms, err := formStruct.GetAllDoing(c, tx, uid)
	// todo 根据订单号去找缓存
	forms, err = services.GetCacheInfo(forms)
	if err != nil {
		middleware.ResponseError(c, 3014, errors.New("服务器错误")) // 3014 缓存处理加工错误
		return
	}
	if forms == nil { // 不存在记录
		middleware.ResponseSuccess(c, dto.FormDoneInfo{Msg: "没查到相关记录", Total: 0, Info: nil})
		return
	}
	middleware.ResponseSuccess(c, PackageReturnFormInfo(forms, 0))
}

// ParasTokenAndGetUId 分析Token以及拿到UID(废弃)
func ParasTokenAndGetUId(c *gin.Context, token string) (*dao.Login, int, error) {
	openId, err := services.GetOpenIdFormToken(token)
	if err != nil {
		return nil, 3002, errors.New("token解析错误")
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		return nil, 10000, errors.New("服务器错误")
	}
	var loginInfo *dao.Login
	loginInfo, err = loginInfo.GetInfoByOpenId(c, tx, openId)
	return loginInfo, -1, nil
}

// PackageReturnFormInfo 包装返回信息
func PackageReturnFormInfo(forms []dao.Form, status int) dto.FormDoneInfo {
	var infoList []dto.FormDoneDetail
	for _, v := range forms {
		temp := dto.FormDoneDetail{}
		temp.FromId = v.RandomId
		temp.Create = v.CreatedAt
		temp.Ban = v.IsBan
		temp.Delete = v.IsDelete
		temp.Title = v.FormInfos.Title
		temp.Tips = v.FormInfos.Tip
		infoList = append(infoList, temp)
	}
	res := dto.FormDoneInfo{
		Msg:    "查询到相关数据",
		Total:  len(infoList),
		Info:   infoList,
		Status: status, // 为已经完成的
	}
	return res
}

// GetUidByContext 通过上下文来获取uid
func GetUidByContext(c *gin.Context) (string, error) {
	uid, ok := c.Get("uid")
	if !ok {
		return "", errors.New("获取UID错误")
	}
	strUid, ok := uid.(string)
	if !ok {
		return "", errors.New("服务器错误")
	}
	return strUid, nil
}

// ParasJson 对表单进行分析
func ParasJson(json string) (string, error) {
	buf := bytes.NewBuffer([]byte(json))
	jsonRes, err := simplejson.NewFromReader(buf)
	if err != nil {
		return "", errors.New("json解析失败")
	}
	title := jsonRes.Get("formTitle").MustString()
	if title == "" {
		title = "未知"
	}
	tips := jsonRes.Get("formTip").MustString()
	if tips == "" {
		tips = "未知"
	}
	return fmt.Sprintf("%v_%v_%v", title, tips, json), nil
}

// JudgeNumOneHas 创建订单判断
func JudgeNumOneHas(uid string) (bool, error) {
	do, err := common.RedisConfDo("ZCARD", uid)
	if err != nil {
		return false, err
	}
	if do == nil {
		return true, nil
	}
	num, ok := do.(int)
	if !ok {
		return false, errors.New("断言错误")
	}
	if num < 20 {
		return true, nil
	}
	return false, nil
}
