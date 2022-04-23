package router

import (
	"github.com/gin-gonic/gin"
	"wechatGin/controller"
	"wechatGin/middleware"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.Default()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 指定路由
	// todo 首页
	homeRouter := router.Group("/home")
	homeRouter.Use(
		middleware.FlowCountMiddleware(),   // 访问记录首页就好了
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	{
		controller.HomeRegister(homeRouter)
	}

	// todo 登陆事项
	loginRouter := router.Group("/login")
	loginRouter.Use(
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	// todo 注册到子路由中
	{
		controller.LoginRegister(loginRouter)
	}

	// todo 表单
	formRouter := router.Group("/form")
	formRouter.Use(
		middleware.VerifyToken(),           // 验证Token
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	{
		controller.FormRegister(formRouter)
	}

	// todo 个人中心
	personRouter := router.Group("/person")
	personRouter.Use(
		middleware.VerifyToken(),           // 验证Token
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	{
		controller.PersonRegister(personRouter)
	}

	// todo 消费者
	orderRouter := router.Group("/order")
	orderRouter.Use(
		middleware.VerifyToken(),           // 验证Token
		middleware.JudgeFillMyself(),       // 判断是不是自己
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	{
		controller.OrderRegister(orderRouter)
	}

	// todo 我的表单啊
	myFormRouter := router.Group("/my_form")
	myFormRouter.Use(
		middleware.VerifyToken(),           // 验证Token
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	{
		controller.MyFormRegister(myFormRouter)
	}

	// todo 填表记录
	formToRouter := router.Group("/turn")
	formToRouter.Use(
		middleware.VerifyToken(),           // 验证Token
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	{
		controller.FromToRegister(formToRouter)
	}

	// todo 填表记录
	manageRouter := router.Group("/manage")
	manageRouter.Use(
		middleware.VerifyToken(),           // 验证Token
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	{
		controller.ManageRegister(manageRouter)
	}

	return router
}
