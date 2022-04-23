package dto

import (
	"github.com/gin-gonic/gin"
	"wechatGin/public"
)

type FormRegisteredInput struct {
	Token string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"` // 通行证
}

func (params *FormRegisteredInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type FormRegisteredOutput struct {
	RandomId string `json:"randomId" form:"randomId" comment:"randomId"` // randomId
	Msg      string `json:"msg" form:"msg" comment:"msg"`                // msg
}

type FormUpdateInput struct {
	Token    string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"`        // 通行证
	Order    string `json:"order" form:"order" comment:"表单随机分配的ID" example:"" validate:"required"`  // 表单随机分配的ID
	FormJson string `json:"form" form:"form" comment:"表单json序列化后字段" example:"" validate:"required"` // 表单json序列化后放入该字段
}

func (params *FormUpdateInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type FormFinalInput struct {
	Token     string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"`                 // 通行证
	Order     string `json:"order" form:"order" comment:"表单随机分配的ID" example:"" validate:"required"`           // 表单随机分配的ID
	FormJson  string `json:"form" form:"form" comment:"表单json序列化" example:"" validate:"required"`             // 表单json序列化后放入该字段
	FormTitle string `json:"name" form:"name" comment:"标题提取出来后" example:"" validate:"required"`               // 最后提交的时候把标题提取出来
	FormTip   string `json:"description" form:"description" comment:"提示提取出来后" example:"" validate:"required"` // 最后提交的时候把提示提取出来
}

func (params *FormFinalInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type FormUpdateOutput struct {
	UpdateTime string `json:"update_time" form:"update_time" comment:"update_time"` // update_time
	Msg        string `json:"msg" form:"msg" comment:"msg"`                         // msg
}

type FormDoneDetail struct {
	FormId string
	Update string
	Ban    int
	Title  string
	Tips   string
}

type FormDoneInfo struct {
	Msg    string           `json:"msg" form:"msg" comment:"msg"`          // msg
	Total  int              `json:"total" form:"total" comment:"total"`    // total
	Info   []FormDoneDetail `json:"info" form:"info" comment:"info"`       // info
	Status int              `json:"status" form:"status" comment:"status"` // info
}

type EmailNoticeInput struct {
	Token string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"`              // 通行证
	Email string `json:"email" form:"email" comment:"联系方式" example:"" validate:"required,validaEmail"` // 联系方式
	Title string `json:"title" form:"title" comment:"主题" example:"" validate:"required"`               // 主题
	Info  string `json:"info" form:"info" comment:"信息" example:"" validate:"required"`                 // 信息
}

func (params *EmailNoticeInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}
