package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/eddieivan01/nic"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"wechatGin/common"
	"wechatGin/dto"
	"wechatGin/public"
)

// GetAccessToken https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/access-token/auth.getAccessToken.html
func GetAccessToken() (string, error) {
	aimUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%v&secret=%v", common.Appid, common.Secret)
	resp, err := nic.Get(aimUrl, nil)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer([]byte(resp.Text))
	jsonRes, err := simplejson.NewFromReader(buf)
	accessToken := jsonRes.Get("access_token").MustString()
	errCode := jsonRes.Get("errcode").MustInt()
	if errCode != 0 {
		errMsg := jsonRes.Get("errmsg").MustString()
		return "", errors.New(errMsg)
	}
	return accessToken, nil
}

// GetCodeInfo 处理二维码分享
func GetCodeInfo(path, order string) (string, error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		return "", errors.New("获取Access_Token错误")
	}
	aimUrl := "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=" + accessToken
	rawInfo := fmt.Sprintf(`{"page": "%v","scene": "%v","check_path": false, "env_version": "develop"}`, path, order)
	resp, err := nic.Post(aimUrl, nic.H{
		Raw: rawInfo,
	})
	if err != nil {
		return "", errors.New("生成二维码错误")
	}
	if len([]byte(resp.Text)) > 500 { // 如果大于500了那么不是Json（真的是莫名其妙的这个接口）
		base64Image := "data:image/png;base64," + public.Base64Encoding([]byte(resp.Text))
		return base64Image, nil
	}
	buf := bytes.NewBuffer([]byte(resp.Text))
	jsonRes, err := simplejson.NewFromReader(buf)
	if err != nil {
		return "", errors.New("解析错误")
	}
	errCode := jsonRes.Get("errcode").MustInt()
	errMsg := jsonRes.Get("errmsg").MustString()
	return "", errors.New(strconv.Itoa(errCode) + ":" + errMsg)
}

// HandleHeader 处理Excel的头部
func HandleHeader(jsonStr string) (dto.ExcelHeader, error) { // 基本的格式是 "序列" + 分析的字段 + "提交时间" + "提交人"
	var headerStruct dto.ExcelHeader
	// todo 先对压缩的Json进行解析
	decompress, err := public.JsonDeTool(jsonStr)
	if err != nil {
		return headerStruct, err
	}
	// todo 开始拼凑
	headerStruct.HeaderName = append(headerStruct.HeaderName, "序列")
	buf := bytes.NewBuffer(decompress)
	jsonRes, err := simplejson.NewFromReader(buf)
	if err != nil {
		return headerStruct, errors.New("解析错误")
	}
	filed := jsonRes.Get("fields").MustArray()
	for k, _ := range filed {
		current := filed[k].(map[string]interface{})
		typeInfo, title := current["type"], current["title"]
		if typeValue, ok := typeInfo.(string); ok {
			keyField := strings.Split(typeValue, "-")[0]
			if _, isIn := public.FormMapping[keyField]; !isIn { // 不在有效字段里面的话(比如排除描述字段)
				continue
			}
			headerStruct.HeaderField = append(headerStruct.HeaderField, typeValue)
			if titleValue, ok := title.(string); ok {
				headerStruct.HeaderName = append(headerStruct.HeaderName, titleValue)
			} else {
				headerStruct.HeaderName = append(headerStruct.HeaderName, "未知字段")
			}
		} else {
			continue
		}
	}
	// todo 补上末尾
	headerStruct.HeaderName = append(headerStruct.HeaderName, "提交时间")
	headerStruct.HeaderName = append(headerStruct.HeaderName, "提交人")
	return headerStruct, nil
}

// HandleJsonBackInfo 处理mongo返回回来的json
func HandleJsonBackInfo(jsonStr string, header dto.ExcelHeader, deleteMap map[string]bool) ([]dto.JsonOne, error) {
	buf := bytes.NewBuffer([]byte(jsonStr))
	jsonRes, err := simplejson.NewFromReader(buf)
	if err != nil {
		return nil, err
	}
	jsonArray := jsonRes.MustArray()
	var resAns []dto.JsonOne
	for k, _ := range jsonArray {
		eachAns := dto.JsonOne{}
		eachByte, err := json.Marshal(jsonArray[k])
		eachBuf := bytes.NewBuffer(eachByte)
		eachJsonRes, err := simplejson.NewFromReader(eachBuf)
		if err != nil {
			continue
		}
		updateTime := eachJsonRes.Get("update_time").MustString()
		nickName := eachJsonRes.Get("nick_name").MustString()
		fromUid := eachJsonRes.Get("from_uid").MustString()
		if _, ok := deleteMap[fromUid]; ok { // 存在里面
			continue
		}
		eachJson := eachJsonRes.Get("fields")
		eachAns.Info = append(eachAns.Info, strconv.Itoa(k))
		for _, v := range header.HeaderField {
			curInfo := handleValue(eachJson.Get(v).Interface())
			eachAns.Info = append(eachAns.Info, curInfo)
			if "未解析到" == curInfo {
				eachAns.Flag = 1
			}
		}
		eachAns.Info = append(eachAns.Info, updateTime)
		eachAns.Info = append(eachAns.Info, nickName)
		resAns = append(resAns, eachAns)
	}
	return resAns, nil
}

func handleValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		var temp []string
		for _, v := range v {
			temp = append(temp, fmt.Sprintf("%v", v))
		}
		res := strings.Join(temp, "|")
		return res
	default:
		return "未解析到"
	}
}

// GetNeedParse 找出哪些需要分析的字段
func GetNeedParse(jsonStr string) (dto.ExcelHeader, error) {
	var ans dto.ExcelHeader
	// todo 先对压缩的Json进行解析
	decompress, err := public.JsonDeTool(jsonStr)
	if err != nil {
		return ans, err
	}
	// todo 开始分析
	buf := bytes.NewBuffer(decompress)
	jsonRes, err := simplejson.NewFromReader(buf)
	if err != nil {
		return ans, errors.New("解析错误")
	}
	filed := jsonRes.Get("fields").MustArray()
	for k, _ := range filed {
		current := filed[k].(map[string]interface{})
		typeInfo, title := current["type"], current["title"]
		if typeValue, ok := typeInfo.(string); ok {
			keyField := strings.Split(typeValue, "-")[0]
			if _, isIn := public.WantParse[keyField]; !isIn {
				continue
			}
			ans.HeaderField = append(ans.HeaderField, typeValue)
			if titleValue, ok := title.(string); ok {
				ans.HeaderName = append(ans.HeaderName, titleValue)
			} else {
				ans.HeaderName = append(ans.HeaderName, "未知字段")
			}
		} else {
			continue
		}
	}
	return ans, nil
}

var globalInfo map[string]map[string]int

// DoParseInfo 分析报表
func DoParseInfo(jsonStr string, want dto.ExcelHeader, deleteMap map[string]bool) ([]dto.ParseEach, error) {
	var resAns []dto.ParseEach
	// todo 判断是不是为空
	if len(want.HeaderField) == 0 {
		return resAns, errors.New("结果为空")
	}
	// todo 创建对应的map
	globalInfo = make(map[string]map[string]int)
	for _, v := range want.HeaderField {
		globalInfo[v] = make(map[string]int)
	}
	filedToTitleMap := make(map[string]string)
	for k, _ := range want.HeaderField {
		filedToTitleMap[want.HeaderField[k]] = want.HeaderName[k]
	}
	// todo 解析Json
	buf := bytes.NewBuffer([]byte(jsonStr))
	jsonRes, err := simplejson.NewFromReader(buf)
	if err != nil {
		return nil, err
	}
	jsonArray := jsonRes.MustArray()
	// todo 遍历拿数据
	for k, _ := range jsonArray {
		eachByte, err := json.Marshal(jsonArray[k])
		eachBuf := bytes.NewBuffer(eachByte)
		eachJsonRes, err := simplejson.NewFromReader(eachBuf)
		if err != nil {
			continue
		}
		fromUid := eachJsonRes.Get("from_uid").MustString()
		if _, ok := deleteMap[fromUid]; ok { // 存在里面
			continue
		}
		fieldInfo := eachJsonRes.Get("fields")
		for _, v := range want.HeaderField {
			makeParseData(v, fieldInfo.Get(v).Interface())
		}
	}
	// todo 拼装结果
	for k, v := range globalInfo {
		eachAns := dto.ParseEach{}
		eachAns.Kind = k
		eachAns.Title = filedToTitleMap[k]
		for key, value := range v {
			temp := dto.EachData{
				Name: key,
				Num:  value,
			}
			eachAns.Data = append(eachAns.Data, temp)
		}
		resAns = append(resAns, eachAns)
	}
	return resAns, nil
}

func makeParseData(filedName string, value interface{}) {
	switch v := value.(type) {
	case string:
		globalInfo[filedName][v]++
	case []interface{}:
		if len(v) != 0 && (filedName == "geo_location" || filedName == "detail_location") {
			valueStr, _ := v[0].(string)
			globalInfo[filedName][valueStr]++
		} else if len(v) != 0 && filedName == "multiple_choice" {
			for _, valueInterface := range v {
				valueStr, _ := valueInterface.(string)
				globalInfo[filedName][valueStr]++
			}
		}
	default:
	}
}
