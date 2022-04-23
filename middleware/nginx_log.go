package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"sync"
	"time"
	"wechatGin/public"
)

type SyncWriter struct {
	m      sync.Mutex
	Writer io.Writer
}

func (w *SyncWriter) Write(b []byte) (n int, err error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.Writer.Write(b)
}

// NginxLogMiddleware 为记录原始IP
func NginxLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// fixed 暂时将bodySize设置为了 -1
		// todo 拿到目录路径
		nginxPath := public.Path("access.log")
		// todo 拿到将要合成的信息
		ip := c.ClientIP()
		utcTimeStr := time.Now().Format("02/Jan/2006:15:04:05")
		url := c.Request.URL
		method := c.Request.Method
		status := c.Writer.Status()
		http := c.Request.Proto
		header := c.Request.UserAgent()
		// todo 写入标准格式
		res := fmt.Sprintf("%v - - [%v +8000] \"%v %v %v\" %v %v \"-\" \"%v\"", ip, utcTimeStr, method, url, http, status, -1, header)
		// todo 写入文件
		file, _ := os.OpenFile(nginxPath, os.O_WRONLY|os.O_APPEND, 0666)
		defer file.Close()
		wr := &SyncWriter{sync.Mutex{}, file}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func(content string) {
			fmt.Fprintln(wr, content)
			wg.Done()
		}(res)
		wg.Wait()
		c.Next()
	}
}
