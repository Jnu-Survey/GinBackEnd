package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"log"
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

// RabbitMQ rabbitMQ结构体
type RabbitMQ struct {
	conn      *amqp.Connection // 链接
	channel   *amqp.Channel    // 通道
	QueueName string           //队列名称
	Exchange  string           //交换机名称
	Key       string           //bind Key 名称
	Mqurl     string           //连接信息
}

// NewRabbitMQ 创建结构体实例
func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	return &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, Mqurl: public.RabbitMQURL}
}

// Destroy 断开 channel 和 connection
func (r *RabbitMQ) Destroy() {
	r.channel.Close() // 断开 channel
	r.conn.Close()    // 断开 conn
}

// 错误处理函数
func (r *RabbitMQ) failOnErr(err error) {
	path := Path("err.log")
	times := time.Now().Unix()
	res := fmt.Sprintf("%v_%v", times, err)
	file, _ := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0666)
	defer file.Close()
	wr := &SyncWriter{sync.Mutex{}, file} // 添加上锁
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(content string) { // fork子协程去写
		fmt.Fprintln(wr, content)
		wg.Done()
	}(res)
	wg.Wait()
}

// NewRabbitMQSimple 创建简单模式下RabbitMQ实例
// 在Simple模式下唯一不同的是 queueName
func NewRabbitMQSimple(queueName string) (*RabbitMQ, error) {
	// todo 创建RabbitMQ实例
	rabbitmq := NewRabbitMQ(queueName, "", "")
	var err error
	// todo 补上conn与channel
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl) // 获取connection
	if err != nil {
		return nil, err
	}
	rabbitmq.channel, err = rabbitmq.conn.Channel() // 获取channel
	if err != nil {
		return nil, err
	}
	return rabbitmq, nil
}

// PublishSimple 简单模式下队列生产
func (r *RabbitMQ) PublishSimple(message string) error {
	// todo 申请队列，如果队列不存在会自动创建，存在则跳过创建
	_, err := r.channel.QueueDeclare(
		r.QueueName, // 首先放入名称
		false,       //是否持久化
		false,       //是否自动删除
		false,       //是否具有排他性
		false,       //是否阻塞处理
		nil,         //额外的属性
	)
	if err != nil {
		return err
	}
	//todo 调用channel 发送消息到队列中
	err = r.channel.Publish(
		r.Exchange, // 此处为空
		r.QueueName,
		false, //如果为true，根据自身exchange类型和routeKey规则；无法找到符合条件的队列会把消息返还给发送者
		false, //如果为true，当exchange发送消息到队列后发现队列上没有消费者，则会把消息返还给发送者
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		return err
	}
	return nil
}

// ConsumeSimple 简单模式下消费者
func (r *RabbitMQ) ConsumeSimple(numbers int, dealFuncList []DealFunc, handle func(msg string, dealList ...DealFunc) (curErr error)) {
	//todo 申请队列，如果队列不存在会自动创建，存在则跳过创建
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		false, //是否持久化
		false, //是否自动删除
		false, //是否具有排他性
		false, //是否阻塞处理
		nil,   //额外的属性
	)
	if err != nil {
		r.failOnErr(err)
	}
	//todo 接收消息
	msg, err := r.channel.Consume(
		q.Name, // queue
		"",     //用来区分多个消费者 此处不区分
		true,   //是否自动应答
		false,  //是否独有
		false,  //设置为true，表示不能将同一个Connection中生产者发送的消息传递给这个Connection中的消费者
		false,  // 是否阻塞处理
		nil,    // 额外的属性
	)
	if err != nil {
		r.failOnErr(err)
	}
	//todo 启用协程处理消息
	forever := make(chan bool)
	for i := 0; i < numbers; i++ {
		go func(i int) {
			for d := range msg {
				err = handle(string(d.Body), dealFuncList...)
				if err != nil {
					r.failOnErr(err)
				} else {
					fmt.Println("处理消息成功")
				}
			}
		}(i)
	}
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
