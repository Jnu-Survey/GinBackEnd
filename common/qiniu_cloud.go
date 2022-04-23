package common

import (
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"strings"
	"wechatGin/public"
)

func JudgeFileName(fileName string) ([]string, bool) {
	strArray := strings.Split(fileName, ".")
	if len(strArray) != 2 {
		return []string{}, false
	}
	return strArray, true
}

func HandleFileName(fileName, uid string) string {
	info := fmt.Sprintf("%v_%v_%v", uid, TempTokenKey, fileName)
	return public.HashSHA256Encoding(info)[:8]
}

// GetQiNiuCloudUpToken 给前端返回七牛云上传Token
func GetQiNiuCloudUpToken(fileName string) string {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", bucketName, fileName),
	}
	putPolicy.Expires = 1800 // 半小时内有效
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)
	return upToken
}
