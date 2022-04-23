package common

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// SendMailUsingTLS 发送邮件
func SendMailUsingTLS(from string, to string, msg []byte) (err error) {
	auth := smtp.PlainAuth("", EmailUser, EmailPassword, strings.Split(EmailHost, ":")[0])
	conn, err := tls.Dial("tcp", EmailHost, nil)
	if err != nil {
		return err
	}
	host, _, _ := net.SplitHostPort(EmailHost)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	// todo 验证是否正确
	if ok, _ := client.Extension("AUTH"); ok {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}
	// todo 向服务器发出Mail命令
	if err = client.Mail(from); err != nil {
		return err
	}
	// todo 分开地址并发送命令 xxx@qq.com;xxx@qq.com
	tos := strings.Split(to, ";")
	for _, addr := range tos {
		if err = client.Rcpt(addr); err != nil { // 向服务器发出Rcpt命令
			return err
		}
	}
	// todo 写入数据
	write, err := client.Data()
	if err != nil {
		return err
	}
	_, err = write.Write(msg)
	if err != nil {
		return err
	}
	err = write.Close()
	if err != nil {
		return err
	}
	err = client.Close() // 手动关闭
	if err != nil {
		return err
	}
	// todo 返回空 手动去结束
	return nil
}

func WaterBody(from, to string, resInfo []string) string {
	header := make(map[string]string)
	header["From"] = "问卷调查意见" + "<" + from + ">"
	header["To"] = to
	header["Subject"] = "问卷调查意见"
	header["Content-Type"] = "text/html;charset=UTF-8"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s:%s\r\n", k, v)
	}
	uid, connection, title, infoStr := resInfo[0], resInfo[1], resInfo[2], resInfo[3]
	body := fmt.Sprintf("<p>UID：%v</p> <p>联系方式：%v</p> <p>主题：%v </p> <p>信息：%v</p>", uid, connection, title, infoStr)
	message += "\r\n" + body
	return message
}
