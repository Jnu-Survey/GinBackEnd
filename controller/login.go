package controller

import (
	"errors"
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
	"wechatGin/public"
	"wechatGin/services"
)

type LoginController struct {
}

func LoginRegister(group *gin.RouterGroup) {
	loginController := &LoginController{}
	group.POST("/dealLogin", loginController.GetToken) // 拿到通行证
	//group.POST("/testLogin", loginController.GetTempToken) // 拿到临时通行证(用于压力测试)
}

func (login *LoginController) GetToken(c *gin.Context) {
	// todo 拿到code
	params := &dto.LoginInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 1001, err)
		return
	}
	// todo 拿到Code后与微信服务器进行交互拿到openId/sessionId/UnionID(暂时拿到openId)
	openId, err := services.Code2Session(params.Code)
	if err != nil {
		middleware.ResponseError(c, 1002, errors.New("与微信服务器交互错误"))
		return
	}
	// todo 根据openId进行查询
	status := 0 // 根据状态往下走 默认是查询得到且不需要更新
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误"))
		return
	}
	var loginInfo *dao.Login
	loginInfo, err = loginInfo.GetInfoByOpenId(c, tx, openId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = 1 // 状态为1即需要注册
		} else {
			middleware.ResponseError(c, 1003, errors.New("数据库错误"))
			return
		}
	}
	// todo 判断是否注册与是否需要更新
	if status != 1 {
		upDateTime := loginInfo.UpdatedAt
		if services.JudgeTime(upDateTime) {
			status = 2 // 状态为1即需要更新
		}
	}
	// todo 根据status进行不同的处理
	if status == 1 {
		temp := &dao.Login{
			AvatarUrl: params.AvatarUrl,
			City:      params.City,
			Country:   params.Country,
			Gender:    params.Gender,
			NickName:  params.NickName,
			Province:  params.Province,
			OpenId:    openId,
		}
		loginInfo, err = loginInfo.RegisterOne(c, tx, temp)
		if err != nil {
			middleware.ResponseError(c, 1004, errors.New("注册失败"))
			return
		}
	} else if status == 2 {
		loginInfo, err = loginInfo.UpdateStatus(c, tx, loginInfo)
		if err != nil {
			middleware.ResponseError(c, 1005, errors.New("更新失败"))
			return
		}
	}
	// todo 打上redis的缓存然后进行返回新的通行证
	token := public.Base64UrlEncoding([]byte(services.TokenOpenId(openId)))
	info := services.PackageInfo(loginInfo)
	err = services.MakeTokenCache(info, token)
	if err != nil {
		middleware.ResponseError(c, 1006, errors.New("服务器错误"))
		return
	}
	output := &dto.LoginOutput{Token: token, Msg: "获得新的Token, 有效时间为3天"}
	middleware.ResponseSuccess(c, output)
}

func (login *LoginController) GetTempToken(c *gin.Context) {
	// todo 拿到code (推荐根据时间戳作为code)
	params := &dto.LoginTempInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 1001, err)
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误"))
		return
	}
	// todo 对时间戳进行处理
	tempOpenId := public.HashSHA256Encoding(fmt.Sprintf("%v_%v", time.Now().UnixNano(), params.TempCode))[2:10]
	// todo 直接注册
	var loginInfo *dao.Login
	temp := &dao.Login{
		AvatarUrl: "https://avatars.githubusercontent.com/u/98681454?v=4",
		NickName:  "测试用户",
		OpenId:    tempOpenId,
	}
	loginInfo, err = loginInfo.RegisterOne(c, tx, temp)
	if err != nil {
		middleware.ResponseError(c, 1004, errors.New("注册失败"))
		return
	}
	// todo 打上redis的缓存然后进行返回新的通行证
	token := public.Base64UrlEncoding([]byte(services.TokenOpenId(tempOpenId)))
	info := services.PackageInfo(loginInfo)
	err = services.MakeTokenCache(info, token)
	if err != nil {
		middleware.ResponseError(c, 1006, errors.New("服务器错误"))
		return
	}
	output := &dto.LoginOutput{Token: token, Msg: "获得新的Token, 有效时间为3天"}
	middleware.ResponseSuccess(c, output)
}
