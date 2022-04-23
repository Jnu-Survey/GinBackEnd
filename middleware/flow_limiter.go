package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"wechatGin/common"
	"wechatGin/public"
)

// FlowLimiterMiddleware 限流器
func FlowLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// todo 针对服务进行限流
		qps := public.FlowCountLimit
		serviceLimiter, err := common.FlowLimiterHandler.GetLimiter(public.FlowTotal, float64(qps))
		if err != nil {
			ResponseError(c, 10006, err)
			c.Abort()
			return
		}
		if !serviceLimiter.Allow() {
			ResponseError(c, 10008, errors.New("保护并限流中"))
			c.Abort()
			return
		}
		c.Next()
	}
}
