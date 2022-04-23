package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strconv"
	"unicode/utf8"
	"wechatGin/common"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
)

type FromToController struct {
}

func FromToRegister(group *gin.RouterGroup) {
	fromToController := &FromToController{}
	group.GET("/fromMyselfDoing", fromToController.FromMyselfDoing) // 我正在填的表
	group.GET("/fromMyselfDone", fromToController.FromMyselfDone)   // 我已经填写好了的表
	group.GET("/toMyselfDone", fromToController.ToMyselfDone)       // 别人填写我的表格
}

func (fromTo *FromToController) FromMyselfDoing(c *gin.Context) {
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 7001, err) // 7001 参数不正确
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 7000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	commitInfos := &dao.Commit{}
	forms, err := commitInfos.GetAllDoingOrDone(c, base.tx, base.uid, 1, 0)
	if err = params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 7002, errors.New("数据库错误"))
		return
	}
	if forms == nil {
		middleware.ResponseSuccess(c, dto.FromDoingOutput{})
		return
	}
	// todo 加工信息
	res := handleInfos(forms, 0) // to
	middleware.ResponseSuccess(c, dto.FromDoingOutput{
		Infos: res,
	})
}

func (fromTo *FromToController) FromMyselfDone(c *gin.Context) {
	params := &dto.MyFormInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 7001, err)
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 7000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	// todo 查询已经填写好了的
	commitInfos := &dao.Commit{}
	forms, err := commitInfos.GetAllDoingOrDone(c, base.tx, base.uid, params.Page, 1)
	if forms == nil {
		middleware.ResponseSuccess(c, dto.FromDoingOutput{})
		return
	}
	res := handleInfos(forms, 0) // to
	middleware.ResponseSuccess(c, dto.FromDoingOutput{
		Infos: res,
	})
}

func (fromTo *FromToController) ToMyselfDone(c *gin.Context) {
	params := &dto.MyFormInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 7001, err)
		return
	}
	base, code, err := GetUidAndDataBaseConnection(c, 7000)
	if err != nil {
		middleware.ResponseError(c, middleware.ResponseCode(code), err)
		return
	}
	// todo 查询已经填写好了的
	commitInfos := &dao.Commit{}
	forms, err := commitInfos.GetAllFillFormForMe(c, base.tx, base.uid, params.Page)
	currentInfos := handleInfos(forms, 1) // 包装下
	marshal, err := json.Marshal(currentInfos)
	if err != nil {
		middleware.ResponseError(c, 7003, errors.New("服务器错误"))
		return
	}
	// todo 升级成WebSocket
	common.WebsocketService.GetPushNews(c, base.uid)
	common.WebsocketService.PushInfo(base.uid, string(marshal))
}

// 进行包装
func handleInfos(forms []dao.Commit, flag int) []dto.EachFromDoing {
	var res []dto.EachFromDoing
	for k, _ := range forms {
		if forms[k].OrderInfos.IsDelete == 1 {
			continue
		}
		each := dto.EachFromDoing{}
		if flag == 0 {
			each.AvatarUrl = forms[k].ToUserInfo.AvatarUrl
			each.Nickname = forms[k].ToUserInfo.NickName
			each.AimUid = strconv.Itoa(forms[k].ToUid)
		} else {
			each.AvatarUrl = forms[k].FromUserInfo.AvatarUrl
			each.Nickname = forms[k].FromUserInfo.NickName
			each.AimUid = strconv.Itoa(forms[k].FromUid)
		}
		each.OrderId = forms[k].OrderId
		each.UpdateTime = forms[k].UpdatedAt.Format("2006-01-02 15:04") // 重新格式化时间
		each.Title = HandleTextLength(forms[k].OrderInfos.FormInfos.Title, 10)
		each.Tips = HandleTextLength(forms[k].OrderInfos.FormInfos.Tip, 16)
		each.IsBan = handleIsBan(forms[k].OrderInfos.IsBan)
		each.IsDelete = handleIsDelete(forms[k].IsDelete)
		res = append(res, each)
	}
	return res
}

// HandleIsBan 处理Ban的状态
func handleIsBan(flag int) string {
	if flag == 0 {
		return "正在收集"
	} else {
		return "停止收集"
	}
}

func handleIsDelete(flag int) string {
	if flag == 0 {
		return "正常状态"
	} else {
		return "无效状态"
	}
}

// HandleTextLength 处理字长
func HandleTextLength(str string, length int) string {
	res := []rune(str)
	now := utf8.RuneCountInString(str)
	if now > length {
		res = res[:length]
		return string(res) + "..."
	}
	return str
}
