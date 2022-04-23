package controller

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
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
	group.GET("/getFormId", formController.GetFormId)    // 获取创建订单ID编号
	group.POST("/tempUpdate", formController.TempUpdate) // 更新记录缓存
	group.POST("/formDone", formController.FormDone)     // 完成表单提交
}

func (form *FormController) GetFormId(c *gin.Context) {
	// 进行了通行证与数量的鉴权
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3001, err)
		return
	}
	// todo 通过缓存来找人
	uid, err := GetInfoByContext("uid", c)
	if err != nil {
		middleware.ResponseError(c, 3999, errors.New("token解析错误"))
		return
	}
	// todo 添加判断数量是否过关
	flag, err := judgeNumOneHas(uid)
	if err != nil {
		middleware.ResponseError(c, 3002, errors.New("服务器错误"))
		return
	}
	if !flag {
		middleware.ResponseError(c, 3003, errors.New("正在创建表单超过规定数量"))
		return
	}
	// todo 此时生成随机ID 解析Token找到对应的人
	randomToken := services.RandomFormId(params.Token)
	// todo 把uid与随机表单号写在有序集合缓存里面
	err = services.MakeFormCache(uid, randomToken)
	if err != nil {
		middleware.ResponseError(c, 3004, errors.New("服务器错误"))
		return
	}
	// todo 将写入数据库的操作写入消息队列
	intUId, _ := strconv.Atoi(uid)
	err = services.PackInfo2Queue(randomToken, intUId)
	if err != nil {
		middleware.ResponseError(c, 3005, errors.New("消息队列发生错误"))
		return
	}
	// todo 返回随机ID
	res := dto.FormRegisteredOutput{
		RandomId: randomToken,
		Msg:      "创建表单成功，已经加入消息队列",
	}
	middleware.ResponseSuccess(c, res)
}

func (form *FormController) TempUpdate(c *gin.Context) {
	params := &dto.FormUpdateInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3001, err) // 3001 参数不正确
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 3000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	tempForm := &dao.Form{}
	err = tempForm.IsDoingAndJudge(c, base.tx, params.Order, base.uid, tempForm)
	if err != nil {
		middleware.ResponseError(c, 3006, err)
		return
	}
	// todo 对表单进行分析
	resInfo, err := ParasJson(params.FormJson)
	if err != nil {
		middleware.ResponseError(c, 3007, errors.New("表单分析错误"))
		return
	}
	// todo 订单号与压缩后的信息建立缓存
	err = services.MakeFormInfoCache(params.Order+"_"+base.uid, public.JsonCoTool(resInfo)) // fixed 修改Bug建立缓存是需要添加上是谁的
	if err != nil {
		middleware.ResponseError(c, 3008, errors.New("更新记录发生错误"))
		return
	}
	returnInfo := dto.FormUpdateOutput{
		UpdateTime: strconv.Itoa(int(time.Now().Unix())),
		Msg:        "本次缓存更新成功",
	}
	middleware.ResponseSuccess(c, returnInfo)
}

func (form *FormController) FormDone(c *gin.Context) {
	// 说一下自己的思考：这里没有使用消息队列写入数据库这是因为在创建的时候
	// 已经进行了削峰的处理，这里只有当事人才能进行唯一的一次提交，可以忽略写的影响
	params := &dto.FormFinalInput{} // XXXX 等待添加验证md5
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3001, err) // 3001 参数不正确
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 3000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	// todo 直接保存到Mysql因为不需要分析只需要给前端进行渲染就好了
	// 事务：判断主表是不是存在；是不是能够修改；修改状态；从表加入数据
	res := public.JsonCoTool(params.FormJson)
	// todo 判断状态
	tempForm := &dao.Form{}
	tempForm, err = tempForm.IsExistAndJudgeStatus(c, base.tx, params.Order, base.uid, tempForm)
	if err != nil {
		middleware.ResponseError(c, 3009, err) // 3009 要么找不到要么不能修改
		return
	}
	// todo 开启事务
	base.tx = base.tx.Begin() // fixed 修复Bug指针丢失的问题
	err = tempForm.RewriteStatus(c, base.tx, tempForm)
	if err != nil {
		base.tx.Rollback()
		middleware.ResponseError(c, 3010, errors.New("服务器错误"))
		return
	}
	formInfo := &dao.FormInfo{FormJson: string(res), Title: params.FormTitle, Tip: params.FormTip, Out: tempForm.Id}
	formInfo, err = formInfo.RecordJson(c, base.tx, formInfo)
	if err != nil {
		base.tx.Rollback()
		middleware.ResponseError(c, 3011, errors.New("服务器错误"))
		return
	}
	base.tx.Commit()
	// todo 删除缓存
	services.DeleteCacheDone(params.Order+"_"+base.uid, params.Order, base.uid) // fixed 要删除正确的缓存
	middleware.ResponseSuccess(c, "记录成功")
}

// PackageReturnFormInfo 包装返回信息
func PackageReturnFormInfo(forms []dao.Form, status int) dto.FormDoneInfo {
	var infoList []dto.FormDoneDetail
	for _, v := range forms {
		if v.IsDelete == 1 {
			continue
		}
		temp := dto.FormDoneDetail{}
		temp.FormId = v.RandomId
		temp.Update = v.UpdatedAt.Format("2006-01-02 15:04") // 重新格式化时间
		temp.Ban = v.IsBan
		temp.Title = HandleTextLength(v.FormInfos.Title, 10)
		temp.Tips = HandleTextLength(v.FormInfos.Tip, 16)
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

// GetInfoByContext 通过上下文来获取
func GetInfoByContext(want string, c *gin.Context) (string, error) {
	uid, ok := c.Get(want)
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
	title := jsonRes.Get("name").MustString()
	if title == "" {
		title = "未知"
	}
	tips := jsonRes.Get("description").MustString()
	if tips == "" {
		tips = "未知"
	}
	TimeLocation, _ := time.LoadLocation("Asia/Shanghai")
	nowTime := time.Now().In(TimeLocation).Unix()
	return fmt.Sprintf("%v%v%v%v%v%v%v", nowTime, public.SplitSymbol, title, public.SplitSymbol, tips, public.SplitSymbol, json), nil
}

// 创建订单判断
func judgeNumOneHas(uid string) (bool, error) {
	do, err := common.RedisConfDo("ZCARD", uid+"_from")
	if err != nil {
		return false, err
	}
	num, ok := do.(int64)
	if !ok {
		return false, errors.New("断言错误")
	}
	if num < public.CreatingNum {
		return true, nil
	}
	return false, nil
}
