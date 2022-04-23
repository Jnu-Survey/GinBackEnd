package middleware

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strconv"
	"wechatGin/dao"
)

// JudgeFillMyself 填写别人的表单
func JudgeFillMyself() gin.HandlerFunc {
	return func(c *gin.Context) {
		// todo 拿到表单号
		temp := ""
		getOrder := c.Query("order")
		postOrder := c.DefaultPostForm("order", "")
		if getOrder != "" {
			temp = getOrder
		} else {
			temp = postOrder
		}
		// todo 查询表单号是谁发出的
		tx, err := lib.GetGormPool("default")
		if err != nil {
			ResponseError(c, 10000, errors.New("服务器错误"))
			c.Abort()
			return
		}
		var detail *dao.Form
		detail, err = detail.GetFormDetailByOrderId(c, tx, temp, false)
		if err != nil {
			ResponseError(c, 10008, err)
			c.Abort()
			return
		}
		// todo 判断
		nowUid, _ := getInfoByContext("uid", c)
		if nowUid == strconv.Itoa(detail.Uid) {
			ResponseError(c, 10009, errors.New("发布者与填写者不能为同一人"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func getInfoByContext(want string, c *gin.Context) (string, error) {
	uid, ok := c.Get(want)
	if !ok {
		return "", errors.New("获取UID错误")
	}
	strUid, ok := uid.(string)
	if !ok {
		return "", errors.New("服务器错误")
	}
	return strUid, nil
}
