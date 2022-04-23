package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"wechatGin/common"
	"wechatGin/public"
)

// FlowCountMiddleware 流量统计
func FlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// todo 拿到计数器
		totalCounter, err := common.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			ResponseError(c, 10006, errors.New("redis错误"))
			c.Abort()
			return
		}
		totalCounter.Increase() // +1
		c.Next()
	}
}
