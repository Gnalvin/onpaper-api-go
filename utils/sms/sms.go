package sms

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"onpaper-api-go/logger"
	"onpaper-api-go/settings"
)

func CreateClient(accessKeyId *string, accessKeySecret *string) (_result *dysmsapi20170525.Client, _err error) {
	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

// SendVerifyCode 发送验证码
func SendVerifyCode(phone, code string) (err error) {
	client, err := CreateClient(tea.String(settings.Conf.SMSSecretId), tea.String(settings.Conf.SMSSecretKey))
	if err != nil {
		return
	}
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		TemplateParam: tea.String(fmt.Sprintf("{\"code\":%s}", code)),
		SignName:      tea.String("onpaper"),
		TemplateCode:  tea.String(settings.Conf.SMSTemplateCode),
	}
	runtime := &util.RuntimeOptions{}
	_, err = client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		logger.ErrZapLog(err, fmt.Sprintf("send phone code: %s", phone))
		return
	}
	logger.InfoZapLog(fmt.Sprintf("send phone code: %s", phone), "")
	return
}
