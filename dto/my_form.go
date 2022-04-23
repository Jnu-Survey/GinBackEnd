package dto

import (
	"github.com/gin-gonic/gin"
	"wechatGin/public"
)

type MyFormInput struct {
	Token string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"` // 通行证
	Page  int    `json:"page" form:"page" comment:"页数" example:"" validate:"required"`    // 页数
}

func (params *MyFormInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type MyFormDetailInput struct {
	Token string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"` // 通行证
	Order string `json:"order" form:"order" comment:"表单号" example:"" validate:"required"` // 表单号
}

func (params *MyFormDetailInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type DoInvalidInput struct {
	Token   string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"`         // 通行证
	Order   string `json:"order" form:"order" comment:"表单号" example:"" validate:"required"`         // 表单号
	FromUid string `json:"from_uid" form:"from_uid" comment:"填写者ID" example:"" validate:"required"` // 填写者ID
}

func (params *DoInvalidInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}
