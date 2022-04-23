package controller

import (
	"encoding/json"
	"errors"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
	"wechatGin/services"
)

type HomeController struct {
}

func HomeRegister(group *gin.RouterGroup) {
	homeController := &HomeController{}
	group.GET("/swiperItem", homeController.SwiperItem)
	group.GET("/buttonTitle", homeController.ButtonTitle)
}

// SwiperItem godoc
// @Summary 获取轮播图信息
// @Description 获取轮播图信息
// @Tags 首页
// @ID /home/swiperItem
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=[]dto.HomeOutput} "success"
// @Router /home/swiperItem [get]
func (home *HomeController) SwiperItem(c *gin.Context) {
	// 因为是首页的话那么不需要Get携带参数
	// todo 先去缓存中查找如果没有再去数据库查找然后打上缓存（有效去为1天吧）
	cache, err := services.GetHomeCache("home")
	if err != nil {
		middleware.ResponseError(c, 2001, errors.New("服务器错误")) // 2001 缓存出问题了
		return
	}
	var res []dao.Home
	if cache == "" { // 如果不存在的话
		tx, err := lib.GetGormPool("default")
		if err != nil {
			middleware.ResponseError(c, 10000, errors.New("服务器错误")) // 10000 池子不通畅
			return
		}
		var homeTemp *dao.Home
		res, err = homeTemp.GetAllInfo(c, tx)
		if err != nil {
			middleware.ResponseError(c, 2002, errors.New("查询数据库错误")) // 2002 要么错误/要么没有
			return
		}
		toJson, err := json.Marshal(res)
		if err != nil {
			middleware.ResponseError(c, 2003, errors.New("服务器错误")) // 2003 序列化错误
			return
		}
		err = services.MakeHomeCache("home", string(toJson))
		if err != nil {
			middleware.ResponseError(c, 2004, errors.New("服务器错误")) // 2004 缓存打不上了
			return
		}
	} else {
		err = json.Unmarshal([]byte(cache), &res)
		if err != nil {
			middleware.ResponseError(c, 2005, errors.New("服务器错误")) // 2004 缓存打不上了
			return
		}
	}
	// todo 组装返回的
	var resList []dto.HomeOutput
	for _, v := range res {
		temp := dto.HomeOutput{
			BackgroundColor: v.Color,
			Title:           v.Title,
			Img:             v.Img,
		}
		resList = append(resList, temp)
	}
	middleware.ResponseSuccess(c, resList)
}

// ButtonTitle godoc
// @Summary 按钮文字
// @Description 按钮文字
// @Tags 首页
// @ID /home/buttonTitle
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /home/buttonTitle [get]
func (home *HomeController) ButtonTitle(c *gin.Context) {
	middleware.ResponseSuccess(c, "快速开始 创建表单")
}
