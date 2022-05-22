package common

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 定义websocket升级
var upGrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 5 * time.Second,
	CheckOrigin: func(r *http.Request) bool { // 取消ws跨域校验
		return true
	},
}

var WebsocketService *WebSocket // 对外暴露

type WebSocket struct {
	news   map[string]chan interface{} // 消息通道
	client map[string]*websocket.Conn  // websocket客户端链接池
	mux    sync.Mutex                  // 互斥锁，防止程序对统一资源同时进行读写
	init   sync.Once                   // 一次加载到内存中
}

// NewOneWebSocketStruct 可以放在最初运行时候开辟新环境进行初始化
func NewOneWebSocketStruct() *WebSocket {
	one := &WebSocket{
		news:   make(map[string]chan interface{}),
		client: make(map[string]*websocket.Conn),
		mux:    sync.Mutex{},
		init:   sync.Once{},
	}
	return one
}

func init() {
	WebsocketService = NewOneWebSocketStruct()
}

// --------------- 以下对结构体的方法进行定义 ---------------

// AddClient 将客户端添加到客户端链接池
func (w *WebSocket) AddClient(id string, conn *websocket.Conn) {
	w.mux.Lock()
	defer w.mux.Unlock()
	w.client[id] = conn // map的并发加锁
}

// GetClient 获取指定客户端链接
func (w *WebSocket) GetClient(id string) (conn *websocket.Conn, exist bool) {
	w.mux.Lock()
	defer w.mux.Unlock()
	conn, exist = w.client[id]
	return
}

// DeleteClient 删除客户端链接
func (w *WebSocket) DeleteClient(id string) {
	w.mux.Lock()
	defer w.mux.Unlock()
	delete(w.client, id)
}

// AddNewsChannel 添加用户消息通道
func (w *WebSocket) AddNewsChannel(id string, m chan interface{}) {
	w.mux.Lock()
	defer w.mux.Unlock()
	w.news[id] = m
}

// GetNewsChannel 获取指定用户消息通道
func (w *WebSocket) GetNewsChannel(id string) (m chan interface{}, exist bool) {
	w.mux.Lock()
	defer w.mux.Unlock()
	m, exist = w.news[id]
	return
}

// DeleteNewsChannel 删除指定消息通道
func (w *WebSocket) DeleteNewsChannel(id string) {
	w.mux.Lock()
	defer w.mux.Unlock()
	if m, ok := w.news[id]; ok {
		close(m)
		delete(w.news, id)
	}
}

// DeleteClientAndChannel 销毁客户端与管道
func (w *WebSocket) DeleteClientAndChannel(id string) {
	// todo 关闭websocket链接
	conn, exist := w.GetClient(id)
	if exist {
		conn.Close()
		w.DeleteClient(id)
	}
	// todo 关闭其消息通道
	_, exist = w.GetNewsChannel(id)
	if exist {
		w.DeleteNewsChannel(id)
	}
}

// WsHandler 处理ws请求
func (w *WebSocket) WsHandler(writer http.ResponseWriter, r *http.Request, id string) {
	var conn *websocket.Conn
	var exist bool
	// todo 创建一个定时器用于服务端心跳
	pingTicker := time.NewTicker(time.Second * 10)
	// todo 将连接进行升级
	conn, err := upGrader.Upgrade(writer, r, nil)
	if err != nil {
		return
	}
	// todo 把与客户端的链接添加到客户端链接池中
	w.AddClient(id, conn)
	// todo 获取该客户端的消息通道
	m, exist := w.GetNewsChannel(id)
	if !exist { // 如果没有
		m = make(chan interface{})
		w.AddNewsChannel(id, m)
	}
	// todo 设置客户端关闭ws链接回调函数
	conn.SetCloseHandler(func(code int, text string) error {
		w.DeleteClientAndChannel(id)
		return nil
	})
	go func() {
		for {
			select {
			case content, _ := <-m: // 从通道接收消息，然后推送给前端
				err = conn.WriteJSON(content)
				if err != nil {
					conn.Close()
					w.DeleteClientAndChannel(id)
					return
				}
			case <-pingTicker.C: // 服务端心跳
				err = conn.WriteJSON("")
				if err != nil {
					conn.Close()
					w.DeleteClientAndChannel(id)
					return
				}
			}
		}
	}()
}

// GetPushNews websocket客户端连接
func (w *WebSocket) GetPushNews(c *gin.Context, id string) {
	// todo 升级成websocket长链接
	w.WsHandler(c.Writer, c.Request, id)
}

// PushInfo 向指定的管道推送消息
func (w *WebSocket) PushInfo(id, msg string) {
	// 检查下管道是否存在
	oneChan, isExist := w.GetNewsChannel(id)
	if !isExist {
		return
	}
	oneChan <- msg // 存在则往里面送东西
}
