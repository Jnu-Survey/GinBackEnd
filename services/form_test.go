package services

import (
	"bytes"
	"fmt"
	"github.com/bitly/go-simplejson"
	"testing"
	"wechatGin/public"
)

func TestFrom(t *testing.T) {
	res := `{"world":"表单标题","formTip":"表单描述","itemList":[{"type":"单行文字","itemTitle":"题目", "itemTip":"题目提示","must":true}]}`
	buf := bytes.NewBuffer([]byte(res))
	jsonRes, err := simplejson.NewFromReader(buf)
	if err != nil {
		fmt.Println(err)
	}
	title := jsonRes.Get("formTitle").MustString()
	if title == "" {
		fmt.Println("1")
	}
	tips := jsonRes.Get("formTip").MustString()
	fmt.Println(title)
	fmt.Println(tips)
}

func TestCom(t *testing.T) {
	res := `KLUv/QQAZQQAhAfmnKrnn6Vf6KGo5Y2V5o+P6L+wX3sid29ybGQiOiKgh+mimCIsImZvcm1UaXAiLCJpdGVtTGlzdCI6W3sidHlwZSI6IuWNleihjOaWh+Wtl1RpdGxlIjoi6aKY55uuIiwgcOaPkOekuiIsIm11c3QiOnRydWV9XX0GADeLmDOvGZU/VQN4wqq6Mw//9YVJ`
	r, _ := public.JsonDecompress(public.Base64Decoding(res))
	fmt.Println(string(r))
}
