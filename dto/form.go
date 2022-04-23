package dto

import (
	"github.com/gin-gonic/gin"
	"time"
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
	Token    string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"`           // 通行证
	Order    string `json:"order" form:"order" comment:"表单随机分配的ID" example:"" validate:"required"`     // 表单随机分配的ID
	FormJson string `json:"form" form:"form" comment:"表单json序列化后放入该字段" example:"" validate:"required"` // 表单json序列化后放入该字段
}

func (params *FormUpdateInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type FormFinalInput struct {
	Token     string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"`            // 通行证
	Order     string `json:"order" form:"order" comment:"表单随机分配的ID" example:"" validate:"required"`      // 表单随机分配的ID
	FormJson  string `json:"form" form:"form" comment:"表单json序列化后放入该字段" example:"" validate:"required"`  // 表单json序列化后放入该字段
	FormTitle string `json:"title" form:"title" comment:"最后提交的时候把标题提取出来" example:"" validate:"required"` // 最后提交的时候把标题提取出来
	FormTip   string `json:"tip" form:"tip" comment:"最后提交的时候把提示提取出来" example:"" validate:"required"`     // 最后提交的时候把提示提取出来
}

func (params *FormFinalInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type FormUpdateOutput struct {
	UpdateTime string `json:"update_time" form:"update_time" comment:"update_time"` // update_time
	Msg        string `json:"msg" form:"msg" comment:"msg"`                         // msg
}

type FormDoneDetail struct {
	FromId string
	Create time.Time
	Delete int
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
