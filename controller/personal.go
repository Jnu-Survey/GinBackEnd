package controller

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
)

type PersonController struct {
}

func PersonRegister(group *gin.RouterGroup) {
	personalController := &PersonController{}
	group.GET("/getBaseInfo", personalController.GetBaseInfo)
}

// GetBaseInfo godoc
// @Summary 个人信息
// @Description 个人信息
// @Tags 个人中心
// @ID /person/getBaseInfo
// @Accept  json
// @Produce  json
// @Param token query string true "通行证"
// @Success 200 {object} middleware.Response{data=[]dto.BaseInfoOutput} "success"
// @Router /person/getBaseInfo [get]
func (person *PersonController) GetBaseInfo(c *gin.Context) {
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4001, err) // 4001 参数不正确
		return
	}
	// todo 通过缓存来找人
	uid, err := GetUidByContext(c)
	if err != nil {
		middleware.ResponseError(c, 4999, err) // 上下文获取UID错误
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误")) // 10000 池子不通畅
		return
	}
	var baseInfo *dao.Login
	info, err := baseInfo.GetUidBaseInfo(c, tx, uid)
	if err != nil {
		middleware.ResponseError(c, 4002, errors.New("服务器错误")) // 4002 数据库错误
		return
	}
	resInfo := dto.BaseInfoOutput{
		ImgUrl:   info.AvatarUrl,
		NickName: info.NickName,
	}
	middleware.ResponseSuccess(c, resInfo)
}
