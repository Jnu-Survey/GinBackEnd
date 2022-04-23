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
	group.GET("/swiperItem", homeController.SwiperItem) // 获取轮播图信息
}

func (home *HomeController) SwiperItem(c *gin.Context) {
	// todo 先去缓存中查找如果没有再去数据库查找然后打上缓存（有效去为1天吧）
	cache, err := services.GetHomeCache("home")
	if err != nil {
		middleware.ResponseError(c, 2001, errors.New("服务器错误"))
		return
	}
	var res []dao.Home
	var flag = true
	if cache != "" {
		err = json.Unmarshal([]byte(cache), &res)
		if err != nil {
			flag = false
		}
	}
	if cache == "" || !flag { // 如果不存在的话
		tx, err := lib.GetGormPool("default")
		if err != nil {
			middleware.ResponseError(c, 10000, errors.New("服务器错误"))
			return
		}
		var homeTemp *dao.Home
		res, err = homeTemp.GetAllInfo(c, tx)
		if err != nil {
			middleware.ResponseError(c, 2002, errors.New("查询数据库错误"))
			return
		}
		toJson, err := json.Marshal(res)
		if err != nil {
			middleware.ResponseError(c, 2003, errors.New("服务器错误"))
			return
		}
		services.MakeHomeCache("home", string(toJson))
	}
	// todo 组装返回的
	var swapping []dto.HomeOutput
	var button []dto.HomeOutput
	for k, _ := range res {
		temp := dto.HomeOutput{
			Jump: res[k].Jump,
			Img:  res[k].Img,
		}
		if k < len(res)-2 { // 这个是轮播图
			swapping = append(swapping, temp)
		} else {
			button = append(button, temp)
		}
	}
	final := dto.HomePartOutput{
		Swapping: swapping,
		Button:   button,
	}
	middleware.ResponseSuccess(c, final)
}
