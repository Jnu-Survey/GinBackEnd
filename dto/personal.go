package dto

import (
	"github.com/gin-gonic/gin"
	"wechatGin/public"
)

type BaseInfoOutput struct {
	ImgUrl   string `json:"img_url" form:"img_url" comment:"img_url"`       // img_url
	NickName string `json:"nick_name" form:"nick_name" comment:"nick_name"` // msg
	Identity string `json:"identity" form:"identity" comment:"identity"`    // identity
}

type DoingOutput struct {
	Id    string `json:"id" form:"id" comment:"id"`          // id
	Title string `json:"title" form:"title" comment:"title"` // title
	Tip   string `json:"tip" form:"tip" comment:"tip"`       // tip
	Time  string `json:"time" form:"time" comment:"time"`    // time
}

type QiNiuCloudInput struct {
	Token    string `json:"token" form:"token" comment:"通行证" example:"" validate:"required"`          // 通行证
	FileName string `json:"file_name" form:"file_name" comment:"文件名字" example:"" validate:"required"` // 文件名字
}

func (params *QiNiuCloudInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type QiNiuCloudTokenOutput struct {
	Token    string `json:"token" form:"token" comment:"token"`             // token
	FileName string `json:"file_name" form:"file_name" comment:"file_name"` // file_name
}

type CountDataOutput struct {
	IDoHowMany  int `json:"i_do_how_many" form:"i_do_how_many" comment:"i_do_how_many"`
	IGetHowMany int `json:"i_get_how_many" form:"i_get_how_many" comment:"i_get_how_many"`
}
