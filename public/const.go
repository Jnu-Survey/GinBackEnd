package public

const (
	ValidatorKey     = "ValidatorKey"
	TranslatorKey    = "TranslatorKey"
	RedisFlowDayKey  = "flow_day_count"
	RedisFlowHourKey = "flow_hour_count"
	FlowTotal        = "flow_total" // 全站
	FlowCountLimit   = 2000         // 限制qps
	CreatingNum      = 10           // 每个用户正在创建的数量
	SplitSymbol      = "%_%"        // 切割信息字符串
)
