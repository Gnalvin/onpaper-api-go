package SendEmail

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
)

// 发件人
const senderEmail = "no_reply@mail.onpaper.cn"
const senderName = "Onpaper"

// SMTP 密码
const password = "Onpaper2021qwL"
const host = "smtpdm.aliyun.com:80"

func SendToMail(senderEmail, senderName, password, host, subject, body, mailType, replyToAddress string, to, cc, bcc []string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", senderEmail, password, hp[0])
	var contentType string
	if mailType == "html" {
		contentType = "Content-Type: text/" + mailType + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	ccAddress := strings.Join(cc, ";")
	bccAddress := strings.Join(bcc, ";")
	toAddress := strings.Join(to, ";")
	msg := []byte("To: " + toAddress + "\r\nFrom: " + senderName + "<" + senderEmail + ">" + "\r\nSubject: =?UTF-8?B? " + subject + "?=\r\nReply-To: " +
		replyToAddress + "\r\nCc: " + ccAddress + "\r\nBcc: " + bccAddress + "\r\n" + contentType + "\r\n\r\n" + body)

	sendTo := MergeSlice(to, cc)
	sendTo = MergeSlice(sendTo, bcc)
	err := smtp.SendMail(host, auth, senderEmail, sendTo, msg)
	return err
}

func MergeSlice(s1 []string, s2 []string) []string {
	slice := make([]string, len(s1)+len(s2))
	copy(slice, s1)
	copy(slice[len(s1):], s2)
	return slice
}

// send 发送邮件主函数
func send(toEmail, subject, body, mailType string) (err error) {
	// 发送地址
	to := []string{toEmail}
	// 抄送地址
	var cc []string
	// 密送地址
	var bcc []string

	// 防止邮件头乱码
	subjectBase := base64.StdEncoding.EncodeToString([]byte(subject))
	replyToAddress := "qiuwenlang@onpaper.cn"

	err = SendToMail(senderEmail, senderName, password, host, subjectBase, body, mailType, replyToAddress, to, cc, bcc)
	if err != nil {
		return err
	}
	return
}

// SendVerifyCode 发送邮箱验证码
func SendVerifyCode(toEmail, code string) (err error) {
	// 邮件名
	subject := fmt.Sprintf("验证码:%s ", code)
	mailType := "html"

	body := AuthCodeTemp
	body = fmt.Sprintf(body, code)

	err = send(toEmail, subject, body, mailType)

	return
}
