package dto

import (
	"wechatGin/public"

	"github.com/gin-gonic/gin"
)

// ------------------ 登陆 ------------------

type LoginInput struct {
	Code      string `json:"code" form:"code" comment:"登录凭证code" example:"" validate:"required"`       // 登录凭证code
	AvatarUrl string `json:"avatarUrl" form:"avatarUrl" comment:"头像地址" example:"" validate:"required"` // 头像地址
	City      string `json:"city" form:"city" comment:"城市" example:"" validate:""`                     // 城市
	Country   string `json:"country" form:"country" comment:"国家" example:"" validate:""`               // 国家
	Gender    int    `json:"gender" form:"gender" comment:"性别" example:"0" validate:""`                // 性别
	NickName  string `json:"nickName" form:"nickName" comment:"微信用户" example:"" validate:"required"`   // 微信用户
	Province  string `json:"province" form:"province" comment:"省会" example:"" validate:""`             // 省会
}

func (params *LoginInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type LoginOutput struct {
	Token string `json:"token" form:"token" comment:"token"` // token
	Msg   string `json:"msg" form:"msg" comment:"msg"`       // msg
}

type LoginTempInput struct {
	TempCode string `json:"temp_code" form:"temp_code" comment:"临时登录凭证temp_code" example:"" validate:"required"` // 临时登录凭证temp_code
}

func (params *LoginTempInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}
