package dto

import (
	"github.com/gin-gonic/gin"
	"wechatGin/public"
)

type FormOrderInput struct {
	Token string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"`  // 通行证
	Order string `json:"order" form:"order" comment:"表单编号" example:"" validate:"required"` // 表单编号
}

func (params *FormOrderInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type FormOrderOutput struct {
	Flag    int    `json:"flag" form:"flag" comment:"flag"`             // flag
	JsonMsg string `json:"json_msg" form:"json_msg" comment:"json_msg"` // json_msg
}
