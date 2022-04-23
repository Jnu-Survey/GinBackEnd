package controller

import (
	"errors"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
	"wechatGin/services"
)

type LoginController struct {
}

func LoginRegister(group *gin.RouterGroup) {
	loginController := &LoginController{}
	group.POST("/dealLogin", loginController.GetTempToken)
}

// GetTempToken godoc
// @Summary 拿到通行证
// @Description 拿到通行证
// @Tags 登陆
// @ID /login/dealLogin
// @Accept  json
// @Produce  json
// @Param body body dto.LoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.LoginOutput} "success"
// @Router /login/dealLogin [post]
func (login *LoginController) GetTempToken(c *gin.Context) {
	// todo 拿到code
	params := &dto.LoginInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 1001, err) // 1001 参数不正确
		return
	}
	// todo 拿到Code后与微信服务器进行交互拿到openId/sessionId/UnionID(暂时拿到openId)
	openId, err := services.Code2Session(params.Code)
	if err != nil {
		middleware.ResponseError(c, 1002, err) // 1002 为Code2Session发生错误
		return
	}
	// todo 根据openId进行查询
	status := 0 // 根据状态往下走 默认是查询得到且不需要更新
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误")) // 10000 池子不通畅
		return
	}
	var loginInfo *dao.Login
	loginInfo, err = loginInfo.GetInfoByOpenId(c, tx, openId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = 1 // 状态为1即需要注册
		} else {
			middleware.ResponseError(c, 1003, err) // 1003 为查询数据库错误
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
			middleware.ResponseError(c, 1004, errors.New("注册失败")) // 1004 注册失败
			return
		}
	} else if status == 2 {
		loginInfo, err = loginInfo.UpdateStatus(c, tx, loginInfo)
		if err != nil {
			middleware.ResponseError(c, 1005, errors.New("更新失败")) // 1005 注册失败
			return
		}
	}
	// todo 打上redis的缓存然后进行返回新的通行证
	token := services.TokenOpenId(openId)
	idStr := strconv.Itoa(loginInfo.Id)
	err = services.MakeTokenCache(idStr, token)
	if err != nil {
		middleware.ResponseError(c, 1006, errors.New("服务器错误")) // 1006 缓存打不进了
		return
	}
	output := &dto.LoginOutput{Token: token, Msg: "获得新的Token, 有效时间为3天"}
	middleware.ResponseSuccess(c, output)
}
