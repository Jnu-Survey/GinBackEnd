package common

const (
	Appid  = "" // 小程序 appId
	Secret = "" // 小程序 appSecret
)

// ----------- 项目密钥 -----------

const (
	// TempTokenKey 需要16位
	TempTokenKey = ""
)

// TempTokenIv 项目偏移量 16位
var TempTokenIv = []byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}

const (
	// RabbitDsn 这个是Rabbit去连接数据库
	RabbitDsn = "root:root@tcp(127.0.0.1:6000)/jnuwechat?charset=utf8&parseTime=true&loc=Asia%2FChongqing"
	// RabbitMQURL 这个是连接服务器中的rabbit
	RabbitMQURL = "amqp://admin:admin@127.0.0.1:5672/"
	// RabbitMysqlConsumeNum Mysql消费大小设置
	RabbitMysqlConsumeNum = 8
)

const (
	// MongoDataSource 连接Mongo
	MongoDataSource    = "mongodb://admin123:admin123@localhost:27017/?authSource=wechat"
	MongoName          = "wechat"
	MongoCollection    = "info"
	MongoMaxCollection = 10
)

const (
	EmailUser     = ""                // 管理员邮箱
	EmailPassword = ""                // 去第三方SMTP申请的密钥
	EmailHost     = "smtp.qq.com:465" // 默认是QQ
	AdminEmail    = ""                // 管理员邮箱
)

const (
	accessKey  = "" // 七牛云 accessKey
	secretKey  = "" // 七牛云 secretKey
	bucketName = "" // 存在哪个桶里面的
)
