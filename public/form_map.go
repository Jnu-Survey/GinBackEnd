package public

var FormMapping = map[string]string{
	"name":             "姓名",
	"email":            "邮箱",
	"phone":            "电话",
	"address":          "地址",
	"single_choice":    "单项选择",
	"multiple_choice":  "多项选择",
	"single_line_text": "单行文字",
	"paragraph_text":   "多行文字",
	"drop_down":        "单选下拉框",
	"date":             "日期",
	"geo_location":     "地理位置",
	"detail_location":  "详细位置",
	"time":             "时间",
	"upload_file":      "上传文件",
}

var WantParse = map[string]bool{
	"single_choice":   true,
	"multiple_choice": true,
	"drop_down":       true,
	"geo_location":    true,
}
