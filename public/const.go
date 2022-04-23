package public

const (
	ValidatorKey     = "ValidatorKey"
	TranslatorKey    = "TranslatorKey"
	RedisFlowDayKey  = "flow_day_count"
	RedisFlowHourKey = "flow_hour_count"
)

const (
	Appid  = "" // 小程序 appId
	Secret = "" // 小程序 appSecret
)

// ----------- 密钥 -----------

const (
	TempTokenKey = ""
	RabbitDsn    = "root:root@tcp(127.0.0.1:3306)/jnuwechat?charset=utf8&parseTime=true&loc=Asia%2FChongqing"
	RabbitMQURL  = "amqp://admin:admin@127.0.0.1:5672/"
)

var TempTokenIv = []byte{}
