package controller

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
	"wechatGin/public"
	"wechatGin/services"
)

type MyFormController struct { // 我的表单
}

func MyFormRegister(group *gin.RouterGroup) {
	myHomeController := &MyFormController{}
	group.GET("/formDoing", myHomeController.GetFormOutListDoing) // 获取我正在制作的表单
	group.GET("/formDone", myHomeController.GetFormOutListDone)   // 获取我制作完成的表单
	group.GET("/detailDone", myHomeController.GetDetailDone)      // 获取指定制作好了的表单信息
	group.GET("/detailDoing", myHomeController.GetDetailDoing)    // 获取指定正在制作的表单信息
}

type BaseInfo struct {
	uid string
	tx  *gorm.DB
}

func (myForm *MyFormController) GetFormOutListDoing(c *gin.Context) {
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 6001, err)
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 6000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	formStruct := &dao.Form{}
	forms, err := formStruct.GetAllDoing(c, base.tx, base.uid)
	// todo 根据订单号去找缓存
	forms, err = services.GetCacheInfo(base.uid, forms)
	if err != nil {
		middleware.ResponseError(c, 6002, errors.New("服务器错误"))
		return
	}
	if forms == nil {
		middleware.ResponseSuccess(c, dto.FormDoneInfo{Msg: "没查到相关记录", Total: 0, Info: nil})
		return
	}
	middleware.ResponseSuccess(c, PackageReturnFormInfo(forms, 0))
}

func (myForm *MyFormController) GetFormOutListDone(c *gin.Context) {
	params := &dto.MyFormInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 6001, err) // 6001 参数不正确
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 6000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	// todo 查询已经填写好了的
	formStruct := &dao.Form{}
	forms, err := formStruct.GetAllDoneInfo(c, base.tx, base.uid, params.Page)
	if err != nil {
		middleware.ResponseError(c, 6003, errors.New("数据库错误"))
		return
	}
	// todo 包装记录
	if forms == nil { // 不存在记录
		middleware.ResponseSuccess(c, dto.FormDoneInfo{Msg: "没查到相关记录", Total: 0, Info: nil})
		return
	}
	middleware.ResponseSuccess(c, PackageReturnFormInfo(forms, 1))
}

func (myForm *MyFormController) GetDetailDone(c *gin.Context) {
	params := &dto.MyFormDetailInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 6001, err) // 6001 参数不正确
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 6000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	formStruct := &dao.Form{}
	formStruct, err = formStruct.GetDetailDone(c, base.tx, params.Order, base.uid, formStruct)
	if err != nil {
		middleware.ResponseError(c, 6004, err)
		return
	}
	decompress, err := public.JsonDeTool(formStruct.FormInfos.FormJson)
	if err != nil {
		middleware.ResponseError(c, 6005, err)
		return
	}
	middleware.ResponseSuccess(c, string(decompress))
}

func (myForm *MyFormController) GetDetailDoing(c *gin.Context) {
	params := &dto.MyFormDetailInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 6001, err) // 6001 参数不正确
		return
	}
	// todo 上下文查找
	uid, err := GetInfoByContext("uid", c)
	if err != nil {
		middleware.ResponseError(c, 6999, err) // 上下文获取UID错误
		return
	}
	// todo 拿到详细信息
	forms := make([]dao.Form, 1)
	for k, _ := range forms {
		forms[k].RandomId = params.Order
	}
	cacheInfo, err := services.GetCacheInfo(uid, forms)
	if err != nil {
		middleware.ResponseError(c, 6006, errors.New("服务器错误"))
		return
	}
	if len(cacheInfo) != 1 {
		middleware.ResponseError(c, 6007, errors.New("渲染错误"))
		return
	}
	temp := cacheInfo[0]
	middleware.ResponseSuccess(c, temp.FormInfos.FormJson)
}

func GetUidAndDataBaseConnection(c *gin.Context, start int) (BaseInfo, int, error) {
	var initInfo BaseInfo
	uid, err := GetInfoByContext("uid", c)
	if err != nil {
		return initInfo, start + 999, errors.New("token解析错误")
	}
	initInfo.uid = uid
	tx, err := lib.GetGormPool("default")
	if err != nil {
		return initInfo, 10000, errors.New("服务器错误")
	}
	initInfo.tx = tx
	return initInfo, -1, nil
}
