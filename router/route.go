package router

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"wechatGin/controller"
	"wechatGin/docs"
	"wechatGin/middleware"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	docs.SwaggerInfo.Title = lib.GetStringConf("base.swagger.title")
	docs.SwaggerInfo.Description = lib.GetStringConf("base.swagger.desc")
	docs.SwaggerInfo.Version = "2.0"
	docs.SwaggerInfo.Host = lib.GetStringConf("base.swagger.host")
	docs.SwaggerInfo.BasePath = lib.GetStringConf("base.swagger.base_path")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	router := gin.Default()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 指定路由
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

	// todo 首页
	homeRouter := router.Group("/home")
	homeRouter.Use(
		middleware.RecoveryMiddleware(),    // 捕获所有panic，并且返回错误信息
		middleware.RequestLog(),            // 请求输出日志,经过这个接口的都会记录到日志中
		middleware.TranslationMiddleware(), // 翻译
	)
	{
		controller.HomeRegister(homeRouter)
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

	return router
}
