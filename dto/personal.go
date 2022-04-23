package dto

type BaseInfoOutput struct {
	ImgUrl   string `json:"img_url" form:"img_url" comment:"img_url"`       // img_url
	NickName string `json:"nick_name" form:"nick_name" comment:"nick_name"` // msg
}
