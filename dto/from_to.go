package dto

type EachFromDoing struct {
	Nickname   string `json:"nickname" form:"nickname" comment:"nickname"`          // 对方昵称
	AvatarUrl  string `json:"avatar_url" form:"avatar_url" comment:"avatar_url"`    // 对方头像地址
	UpdateTime string `json:"update_time" form:"update_time" comment:"update_time"` // 首次进入时间
	Title      string `json:"title" form:"title" comment:"title"`                   // 表单标题
	Tips       string `json:"tips" form:"tips" comment:"tips"`                      // 表单Tips
	IsBan      string `json:"is_ban" form:"is_ban" comment:"is_ban"`                // 表单状态
	OrderId    string `json:"order_id" form:"order_id" comment:"order_id"`          // 订单号
	IsDelete   string `json:"is_delete" form:"is_delete" comment:"is_delete"`       // 是否被删除无效了
	AimUid     string `json:"aim_uid" form:"aim_uid" comment:"aim_uid"`             // 对方的UId
}

type FromDoingOutput struct {
	Infos []EachFromDoing `json:"infos" form:"infos" comment:"infos"` // infos
}
