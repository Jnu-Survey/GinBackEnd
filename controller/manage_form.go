package controller

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"
	"wechatGin/common"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
	"wechatGin/public"
	"wechatGin/rabbitmq"
	"wechatGin/services"
)

type ManageController struct {
}

func ManageRegister(group *gin.RouterGroup) {
	manageController := &ManageController{}
	group.GET("/switch", manageController.Switch)                     // 是否开启与关闭
	group.GET("/shareTemplate", manageController.ShareTemplate)       // 发表模版
	group.GET("/getShareTemplate", manageController.GetShareTemplate) // 拿到模版信息
	group.GET("/copyForm", manageController.CopyForm)                 // 复制表单
	group.GET("/deleteForm", manageController.DeleteForm)             // 删除表单
	group.GET("/shareCode", manageController.ShareCode)               // 分享表单
	group.POST("/commit", manageController.CommitDataBySelf)          // 自己为自己表单添加数据
	group.GET("/getAllDoneInfo", manageController.GetAllDoneInfo)     // 查看数据
	group.GET("/getAnalyzeData", manageController.GetAnalyzeData)     // 查看报表
	group.GET("/doInvalid", manageController.DoInvalid)               // 设置无效
}

type manageInit struct {
	code     int
	uid      string
	tx       *gorm.DB
	formInfo *dao.Form
	err      error
	fromJson string
	order    string
	fromUid  string
}

type CodeErr struct {
	code int
	err  error
}

func (manage *ManageController) Switch(c *gin.Context) {
	// todo 初始化
	params := &dto.MyFormDetailInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	// todo 判断状态以及修改状态
	notice := ""
	if manageBag.formInfo.IsBan == 0 {
		manageBag.formInfo.IsBan = 1
		notice = "修改状态为：拒绝接受新的提交"
	} else {
		manageBag.formInfo.IsBan = 0
		notice = "修改状态为：恢复接受新的提交"
	}
	// todo 回写数据
	err := manageBag.formInfo.ChangeBanStatus(c, manageBag.tx, manageBag.formInfo)
	if err != nil {
		middleware.ResponseError(c, 8004, errors.New("修改状态失败"))
		return
	}
	middleware.ResponseSuccess(c, notice)
}

func (manage *ManageController) ShareTemplate(c *gin.Context) {
	// todo 初始化
	params := &dto.MyFormDetailInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	newTempShare := public.HashSHA256Encoding(manageBag.formInfo.RandomId + common.TempTokenKey)[:8] // 订单号确定了分享号就确定了
	// todo 判断是不是已经分享了，如果已经存在记录了则直接返回分享码
	shareInfo := &dao.ShareInfo{}
	info, err := shareInfo.IsExistShareInfo(c, manageBag.tx, newTempShare)
	if err != nil {
		middleware.ResponseError(c, 8005, err)
		return
	}
	if info != "" { // 如果已经存在了那么直接返回
		middleware.ResponseSuccess(c, info)
		return
	}
	shareInfo.ShareId = newTempShare
	shareInfo.Out = manageBag.formInfo.Id
	shareInfo.Parent = manageBag.formInfo.Uid
	err = shareInfo.MakeRecord(c, manageBag.tx, shareInfo)
	if err != nil {
		middleware.ResponseError(c, 8006, errors.New("服务器错误"))
		return
	}
	middleware.ResponseSuccess(c, newTempShare)
}

func (manage *ManageController) GetShareTemplate(c *gin.Context) {
	// todo 初始化
	params := &dto.MyFormDetailInput{}
	err := params.BindValidParam(c)
	if err != nil {
		middleware.ResponseError(c, 8001, err)
		return
	}
	// todo 拿到数据库链接
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误"))
		return
	}
	shareInfo := &dao.ShareInfo{}
	info, err := shareInfo.GetParentOrderJsonInfo(c, tx, params.Order)
	if err != nil {
		middleware.ResponseError(c, 8007, err)
		return
	}
	middleware.ResponseSuccess(c, info)
}

func (manage *ManageController) CopyForm(c *gin.Context) {
	// todo 初始化  	// XXXX 后期可以加一个冷却时间
	params := &dto.MyFormDetailInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	// todo 处理新的order
	newOrderId := public.HashSHA256Encoding(manageBag.formInfo.RandomId + strconv.Itoa(int(time.Now().Unix())))[:16]
	// todo 开启事务
	manageBag.tx = manageBag.tx.Begin()
	var newForm = &dao.Form{
		RandomId: newOrderId,
		Uid:      manageBag.formInfo.Uid,
		Status:   1,
	}
	newForm, err := newForm.AddForm2Uid(manageBag.tx, newForm)
	if err != nil {
		manageBag.tx.Rollback()
		middleware.ResponseError(c, 8008, errors.New("服务器错误"))
		return
	}
	newFormInfo := &dao.FormInfo{FormJson: manageBag.formInfo.FormInfos.FormJson, Title: "[新]-" + manageBag.formInfo.FormInfos.Title, Tip: manageBag.formInfo.FormInfos.Tip, Out: newForm.Id}
	newFormInfo, err = newFormInfo.RecordJson(c, manageBag.tx, newFormInfo)
	if err != nil {
		manageBag.tx.Rollback()
		middleware.ResponseError(c, 8009, errors.New("服务器错误"))
		return
	}
	manageBag.tx.Commit()
	middleware.ResponseSuccess(c, "复制成功，请在我的表单中刷新")
}

func (manage *ManageController) DeleteForm(c *gin.Context) {
	// todo 初始化
	params := &dto.MyFormDetailInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	// todo 判断状态然后进行修改
	if manageBag.formInfo.IsDelete == 1 {
		middleware.ResponseSuccess(c, "该表单已经删除")
	}
	manageBag.formInfo.IsDelete = 1
	err := manageBag.formInfo.ChangeBanStatus(c, manageBag.tx, manageBag.formInfo)
	if err != nil {
		middleware.ResponseError(c, 8010, errors.New("服务器错误"))
		return
	}
	middleware.ResponseSuccess(c, "删除成功")
}

func (manage *ManageController) ShareCode(c *gin.Context) {
	// todo 初始化
	params := &dto.MyFormDetailInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	base64Image, err := services.GetCodeInfo("pages/show/show", manageBag.formInfo.RandomId)
	if err != nil {
		middleware.ResponseError(c, 8011, err)
		return
	}
	res := dto.CodeOutput{
		Path:        "pages/show/show",
		Params:      manageBag.formInfo.RandomId,
		Base64Image: base64Image,
	}
	middleware.ResponseSuccess(c, res)
}

func (manage *ManageController) CommitDataBySelf(c *gin.Context) {
	// todo 初始化
	params := &dto.FormUpdateInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	nickName, err := GetInfoByContext("nickName", c)
	if err != nil {
		middleware.ResponseError(c, 8998, err)
		return
	}
	codeErrStruct := insertMongoAndMysql(manageBag.tx, manageBag.fromJson, manageBag.uid, manageBag.order, nickName, 8012)
	if codeErrStruct.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(codeErrStruct.code), err)
		return
	}
	middleware.ResponseSuccess(c, "添加数据成功")
}

func (manage *ManageController) GetAllDoneInfo(c *gin.Context) {
	params := &dto.MyFormDetailInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	// todo 判断数量
	codeErr := judgeExistData(c, manageBag)
	if codeErr.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(codeErr.code), codeErr.err)
		return
	}
	// todo 分析excel头部
	headerStruct, err := services.HandleHeader(manageBag.formInfo.FormInfos.FormJson)
	if err != nil {
		middleware.ResponseError(c, 8019, errors.New("表格头部解析错误"))
		return
	}
	// todo 根据表单对mongo数据库进行查询
	jsonStr, codeErr := getMongoData(manageBag)
	if codeErr.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(codeErr.code), codeErr.err)
		return
	}
	// todo 拿到无效的列表
	commitInfos := &dao.Commit{}
	uidMap, err := commitInfos.GetValidFromUId(manageBag.tx, manageBag.uid, manageBag.order)
	if err != nil {
		middleware.ResponseError(c, 8024, err)
		return
	}
	// todo 对返回回来的每条数据进行分析
	ansInfo, err := services.HandleJsonBackInfo(jsonStr, headerStruct, uidMap)
	if err != nil {
		return
	}
	// todo 返回结果
	var resStruct dto.ParseFormOutput
	resStruct.Header = headerStruct.HeaderName
	resStruct.Body = ansInfo
	middleware.ResponseSuccess(c, resStruct)
}

func (manage *ManageController) GetAnalyzeData(c *gin.Context) {
	params := &dto.MyFormDetailInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	// todo 判断数量
	codeErr := judgeExistData(c, manageBag)
	if codeErr.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(codeErr.code), codeErr.err)
		return
	}
	// todo 分析创建表单的json找出哪些需要分析字段
	needHeader, err := services.GetNeedParse(manageBag.formInfo.FormInfos.FormJson)
	if err != nil {
		middleware.ResponseError(c, 8021, errors.New("解析错误"))
		return
	}
	// todo 如果为空那么就直接溜了
	if len(needHeader.HeaderField) == 0 || len(needHeader.HeaderField) == 0 {
		middleware.ResponseSuccess(c, []dto.ParseEach{})
		return
	}
	// todo 拿到无效的列表
	commitInfos := &dao.Commit{}
	uidMap, err := commitInfos.GetValidFromUId(manageBag.tx, manageBag.uid, manageBag.order)
	if err != nil {
		middleware.ResponseError(c, 8024, err)
		return
	}
	// todo 查询mongo
	jsonStr, codeErr := getMongoData(manageBag)
	if codeErr.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(codeErr.code), codeErr.err)
		return
	}
	// todo 进行分析并返回
	data, err := services.DoParseInfo(jsonStr, needHeader, uidMap)
	if err != nil {
		middleware.ResponseError(c, 8022, errors.New("分析错误"))
		return
	}
	middleware.ResponseSuccess(c, data)
}

func (manage *ManageController) DoInvalid(c *gin.Context) {
	params := &dto.DoInvalidInput{}
	manageBag := initFunc(c, params)
	if manageBag.err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(manageBag.code), manageBag.err)
		return
	}
	commitInfo := &dao.Commit{}
	err := commitInfo.HandleValid(manageBag.tx, manageBag.fromUid, manageBag.uid, manageBag.order)
	if err != nil {
		middleware.ResponseError(c, 8023, err)
		return
	}
	middleware.ResponseSuccess(c, "设置无效操作成功")
}

// 初始化
func initFunc(c *gin.Context, params interface{}) *manageInit {
	manageStruct := &manageInit{}
	// todo 检查参数
	var tempErr error
	switch v := params.(type) {
	case *dto.MyFormDetailInput:
		tempErr = v.BindValidParam(c)
		manageStruct.order = v.Order
	case *dto.FormUpdateInput:
		tempErr = v.BindValidParam(c)
		manageStruct.order = v.Order
		manageStruct.fromJson = v.FormJson
	case *dto.DoInvalidInput:
		tempErr = v.BindValidParam(c)
		manageStruct.order = v.Order
		manageStruct.fromUid = v.FromUid
	default:
		tempErr = errors.New("格式不对")
	}
	if tempErr != nil {
		manageStruct.code = 8001
		manageStruct.err = tempErr
		return manageStruct
	}
	// todo 上下文查找
	uid, err := GetInfoByContext("uid", c)
	if err != nil {
		manageStruct.code = 8999
		manageStruct.err = errors.New("token解析错误")
		return manageStruct
	}
	manageStruct.uid = uid
	// todo 拿到数据库链接
	tx, err := lib.GetGormPool("default")
	if err != nil {
		manageStruct.code = 10000
		manageStruct.err = errors.New("服务器错误")
		return manageStruct
	}
	manageStruct.tx = tx
	// todo 拿到订单的基础信息
	formStruct := &dao.Form{}
	formStruct, err = formStruct.GetFormDetailByOrderId(c, manageStruct.tx, manageStruct.order, true)
	if err != nil {
		manageStruct.code = 8002
		manageStruct.err = errors.New("数据库查询错误")
		return manageStruct
	}
	manageStruct.formInfo = formStruct
	// todo 判断是不是本人
	if strconv.Itoa(manageStruct.formInfo.Uid) != manageStruct.uid {
		manageStruct.code = 8003
		manageStruct.err = errors.New("非法用户")
		return manageStruct
	}
	return manageStruct
}

// 自己为自己添加记录
func insertMongoAndMysql(tx *gorm.DB, json, uid, order, nickName string, code int) *CodeErr {
	codeErrStruct := &CodeErr{}
	tempCommit := &dao.Commit{}
	// todo 开启事务
	tx = tx.Begin()
	// todo 加入Mongo记录去拿ID
	mongoConnect, err := common.NewMongoDbPool()
	if err != nil {
		codeErrStruct.code = code
		codeErrStruct.err = errors.New("数据库错误")
		return codeErrStruct
	}
	json, err = rabbitmq.HandleJsonInfo(json, order, nickName, uid)
	if err != nil {
		codeErrStruct.code = code + 1
		codeErrStruct.err = errors.New("处理字段错误")
		return codeErrStruct
	}
	dbStr, err := mongoConnect.InsertToDb(json)
	if err != nil {
		codeErrStruct.code = code + 2
		codeErrStruct.err = errors.New("数据库错误")
		return codeErrStruct
	}
	// todo 加入Mysql记录
	uidInt, _ := strconv.Atoi(uid)
	tempCommit = &dao.Commit{
		FromUid: uidInt,
		ToUid:   uidInt,
		OrderId: order,
		Status:  1,
		HexId:   dbStr,
	}
	tempCommit, err = tempCommit.AddFrom2To(tx, tempCommit)
	if err != nil {
		tx.Rollback()
		codeErrStruct.code = code + 3
		codeErrStruct.err = errors.New("数据库错误")
		return codeErrStruct
	}
	formInfo := &dao.CommitInfo{FormJson: public.JsonCoTool(json), Out: tempCommit.Id}
	formInfo, err = formInfo.RecordJson(tx, formInfo)
	if err != nil {
		tx.Rollback()
		codeErrStruct.code = code + 4
		codeErrStruct.err = errors.New("数据库错误")
		return codeErrStruct
	}
	tx.Commit()
	return codeErrStruct
}

func judgeExistData(c *gin.Context, manageBag *manageInit) *CodeErr {
	codeErrStruct := &CodeErr{}
	// todo 判断数量
	commit := &dao.Commit{}
	commitInfos, err := commit.GetInfoToMeByOrder(c, manageBag.tx, manageBag.formInfo.RandomId, manageBag.uid)
	if err != nil {
		codeErrStruct.code = 8017
		codeErrStruct.err = errors.New("查询错误")
		return codeErrStruct
	}
	if len(commitInfos) == 0 {
		codeErrStruct.code = 8018
		codeErrStruct.err = errors.New("不存在填表记录")
		return codeErrStruct
	}
	return codeErrStruct
}

func getMongoData(manageBag *manageInit) (string, *CodeErr) {
	codeErrStruct := &CodeErr{}
	mongoConnect, err := common.NewMongoDbPool()
	if err != nil {
		codeErrStruct.code = 8020
		codeErrStruct.err = errors.New("服务器错误")
		return "", codeErrStruct
	}
	jsonStr, err := mongoConnect.FindInfoByField("order_id_key", manageBag.formInfo.RandomId)
	if err != nil {
		codeErrStruct.code = 8021
		codeErrStruct.err = errors.New("服务器错误")
		return "", codeErrStruct
	}
	return jsonStr, codeErrStruct
}
