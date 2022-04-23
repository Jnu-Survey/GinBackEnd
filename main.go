package main

import (
	"github.com/e421083458/golang_common/lib"
	"os"
	"os/signal"
	"syscall"
	"wechatGin/router"
)

func main() {
	lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
	defer lib.Destroy()

	router.HttpServerRun()
	redisTimer := router.NewRedisTimer()
	redisTimer.RedisTimerStart()

	// 进入阻塞
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 退出
	redisTimer.RedisTimerStop()
	router.HttpServerStop()
}
