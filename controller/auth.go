package controller

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"onpaper-api-go/cache"
	"onpaper-api-go/dao/mysql"
	m "onpaper-api-go/models"
	"onpaper-api-go/settings"
	SendEmail "onpaper-api-go/utils/email"
	"onpaper-api-go/utils/encrypt"
	"onpaper-api-go/utils/jwt"
	"onpaper-api-go/utils/oss"
	"onpaper-api-go/utils/sms"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SignIn(ctx *gin.Context) {
	// 查看之前缓存是否都设置成功
	dataCtx, _ := ctx.Get("isOk")
	isOk := dataCtx.(bool)
	if !isOk {
		ResponseError(ctx, CodeServerBusy)
		return
	}

	//获取到 用户信息
	userInfo, _ := ctx.Get("userInfo")
	userData := userInfo.(m.UserTableInfo)

	//生成token
	payload := &m.UserTokenPayload{
		Id:    userData.SnowId,
		Phone: userData.Phone,
		Email: userData.Email.String,
		MD5:   encrypt.CreatMd5(userData.Password),
	}

	refreshToken, err := jwt.CreateRefreshToken(payload)
	accessToken, err := jwt.CreateAccessToken(payload, time.Minute*15)
	if err != nil {
		err = errors.Wrap(err, "SingIn: create token fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//保存 tokenMd5
	err = cache.SetTokenMd5(*payload)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 返回数据
	ResponseSuccess(ctx, gin.H{
		"userId":       userData.SnowId,
		"userName":     userData.UserName,
		"avatar":       userData.Avatar,
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	})

	ctx.Set("userInfo", *payload)

}

// CreateToken 生成accessToken
func CreateToken(ctx *gin.Context) {
	userInfo, _ := ctx.Get("userInfo")
	userData := userInfo.(m.UserTokenPayload)

	//生成token
	accessToken, err := jwt.CreateAccessToken(&userData, time.Minute*15)
	if err != nil {
		err = errors.Wrap(err, "SingIn: create token fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 返回数据
	ResponseSuccess(ctx, gin.H{
		"accessToken": accessToken,
	})
}

// ReturnSTSData  返回sts数据
func ReturnSTSData(ctx *gin.Context) {
	// 验证 url参数 上传的是什么类型 avatars or banner xxx
	stsType, ok := ctx.GetQuery("type")
	count, ok := ctx.GetQuery("count")
	if !ok {
		// 如果没有参数
		ResponseError(ctx, CodeParamsError)
		return
	}
	fileCount, err := strconv.Atoi(count)
	if err != nil {
		// count 不是数字
		ResponseError(ctx, CodeParamsError)
		return
	}
	// 一次不能超过16个
	if fileCount >= 16 {
		ResponseError(ctx, CodeParamsError)
		return
	}

	// 验证类型 符号规范
	typeVerify := false
	typeList := []string{"avatars", "banners", "artworks", "trends", "messages", "commission"}
	for _, s := range typeList {
		if stsType == s {
			typeVerify = true
		}
	}
	// 不属于类型 返回错误
	if typeVerify != true {
		ResponseError(ctx, CodeParamsError)
		return
	}

	// 获取传递的参数 后面需要用户名生成路径
	userInfo, exi := ctx.Get("userInfo")
	if !exi {
		err = errors.New("ProfileData get ctxData fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	userData := userInfo.(m.UserTokenPayload)

	// 生成将上传的的文件ID 作为文件名
	var fileNameList []string
	for i := 0; i < fileCount; i++ {
		fileName := encrypt.CreateUUID()
		fileNameList = append(fileNameList, fileName)
	}

	//得到sts 数据
	res, err := oss.CreatSTS(userData.Id, stsType)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 构建结果
	var STSResult = m.STSResult{
		Token: m.Credentials{
			AccessKeyId:     res.Body.Credentials.AccessKeyId,
			AccessKeySecret: res.Body.Credentials.AccessKeySecret,
			SecurityToken:   res.Body.Credentials.SecurityToken,
			Expiration:      res.Body.Credentials.Expiration,
		},
		FileName: fileNameList,
	}

	ResponseSuccess(ctx, STSResult)
}

// VerifyName 验证用户名是否存在
func VerifyName(ctx *gin.Context) {

	userName, exi := ctx.GetQuery("username")
	if !exi {
		// 如果请求不存在该参数返回错误
		ResponseError(ctx, CodeParamsError)
		return
	}
	// 到数据库中查询
	err := mysql.CheckUserExistByName(userName)
	if err != nil {
		switch err {
		case mysql.ErrorUserExist:
			// 用户存在时
			ResponseError(ctx, CodeUserAlreadyExists)
		case mysql.ErrorUserNotExist:
			// 用户不存在时
			ResponseError(ctx, CodeUserDoseNotExists)
		default:
			// 其他情况数据库查询错误时
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
		}
	}
}

// VerifyEmailExist 验证邮箱是否存在
func VerifyEmailExist(ctx *gin.Context) {
	userEmail, exi := ctx.GetQuery("email")
	if !exi {
		// 如果请求不存在该参数返回错误
		ResponseError(ctx, CodeParamsError)
		return
	}
	// 到数据库中查询
	err := mysql.CheckUserExistByEmail(userEmail)
	if err != nil {
		switch err {
		case mysql.ErrorEmailExist:
			// 用户存在时
			ResponseError(ctx, CodeEmailExists)
		case mysql.ErrorEmailNotExist:
			// 用户不存在时
			ResponseError(ctx, CodeEmailNoExists)
		default:
			// 其他情况数据库查询错误时
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
		}
	}
}

// VerifyPhoneExist 验证手机是否被注册过
func VerifyPhoneExist(ctx *gin.Context) {
	//获取到 邮箱信息
	ctxData, _ := ctx.Get("phone")
	phoneData := ctxData.(m.VerifyPhone)

	isExist, err := mysql.CheckUserExistByPhone(phoneData.Phone)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"isExist": isExist,
	})
}

// SendEmailCode 发送邮箱验证码
func SendEmailCode(ctx *gin.Context) {
	//获取到 邮箱信息
	ctxData, _ := ctx.Get("email")
	email := ctxData.(string)

	ctxData, _ = ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	//邮箱验证码发送 两种情况 1。绑定邮箱 2。邮箱登陆
	//邮箱登陆时 需要验证邮箱是否注册过（避免刷接口）， 绑定邮箱者不需要
	if userInfo.Id == "" {
		err := mysql.CheckUserExistByEmail(email)
		if err != nil {
			switch err {
			case mysql.ErrorEmailExist:
			case mysql.ErrorEmailNotExist:
				// 用户不存在时
				ResponseError(ctx, CodeEmailNoExists)
				return
			default:
				// 其他情况数据库查询错误时
				ResponseErrorAndLog(ctx, CodeServerBusy, err)
				return
			}
		}
	}

	// 验证码
	code := encrypt.RandStr(6, "num")
	err := SendEmail.SendVerifyCode(email, code)
	if err != nil {
		err = errors.Wrap(err, "SendEmailCode: send email fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	//保存到 Redis
	err = cache.SetEmailVerifyCode(email, code)
	if err != nil {
		err = errors.Wrap(err, "SendEmailCode: save email code redis fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"email":  email,
		"status": "ok",
	})
}

// VerifyLoginPhone  验证登陆手机是否注册过
func VerifyLoginPhone(ctx *gin.Context) {
	phone, _ := ctx.Get("phone")
	verifyData := phone.(m.VerifyPhone)

	//1.通过手机号判断用户是否注册过 如果没注册过 没有邀请码不发送验证码
	isExist, err := mysql.CheckUserExistByPhone(verifyData.Phone)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//如果是新用户  验证邀请码的正确
	if !isExist && verifyData.Code != settings.Conf.MagicCode {
		if verifyData.Code == "" {
			ResponseError(ctx, CodeNeedInviteCode)
			return
		}
		err = mysql.CheckInviteCode(verifyData.Code)
		if err != nil {
			// 邀请码无效
			if errors.Cause(err) == mysql.ErrorInviteCodeInvalid {
				ResponseError(ctx, CodeInviteCodeInvalid)
				return
			} else {
				ResponseErrorAndLog(ctx, CodeServerBusy, err)
				return
			}
		}
	}
}

func SendPhoneCode(ctx *gin.Context) {
	phone, _ := ctx.Get("phone")
	verifyData := phone.(m.VerifyPhone)

	// 验证码
	code := encrypt.RandStr(6, "numNoZero")
	err := sms.SendVerifyCode(verifyData.Phone, code)
	if err != nil {
		err = errors.Wrap(err, "SendPhoneCode: send phone fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	//保存到 Redis
	err = cache.SetPhoneVerifyCode(verifyData.Phone, code)
	if err != nil {
		err = errors.Wrap(err, "SendPhoneCode: save phone code fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{"status": "ok"})
}

// SendSafetyPhoneCode 向绑定手机发送验证码
func SendSafetyPhoneCode(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	// 验证码
	code := encrypt.RandStr(6, "num")
	err := sms.SendVerifyCode(userInfo.Phone, code)
	if err != nil {
		err = errors.Wrap(err, "SendPhoneCode: send phone fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	//保存到 Redis
	err = cache.SetPhoneVerifyCode(userInfo.Phone, code)
	if err != nil {
		err = errors.Wrap(err, "SendSafetyPhoneCode: save phone code fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"status": "ok",
	})
}

// HandleLoginOrRegister 验证是注册还是登陆
func HandleLoginOrRegister(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("loginForm")
	loginForm, _ := ctxData.(m.LoginForm)

	// 验证步骤不能改
	// 1。验证验证码
	code, err := cache.GetPhoneVerifyCode(loginForm.Phone)
	if err != nil {
		err = errors.New("HandleLoginOrRegister redis get code fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 不一致返回错误
	if code != loginForm.VerifyCode {
		ResponseError(ctx, CodeCodeVerifyError)
		return
	}

	if loginForm.InviteCode != "" && loginForm.InviteCode != settings.Conf.MagicCode {
		// 新用户会携带邀请码
		err = mysql.CheckInviteCode(loginForm.InviteCode)
		if err != nil {
			// 邀请码无效
			if errors.Cause(err) == mysql.ErrorInviteCodeInvalid {
				ResponseError(ctx, CodeInviteCodeInvalid)
				return
			} else {
				ResponseErrorAndLog(ctx, CodeServerBusy, err)
				return
			}
		}
	}

	//2。通过手机号判断用户是否注册过
	userInfo, err := mysql.GetUserByPhone(loginForm.Phone)
	if err != nil {
		// 如果没有查询到用户 又没写密码（注册需要）或者邀请码 返回错误
		if errors.Cause(err) == sql.ErrNoRows {
			if loginForm.InviteCode == "" {
				ResponseError(ctx, CodeInviteCodeInvalid)
				return
			}
			if loginForm.Password == "" {
				ResponseError(ctx, CodeUserDoseNotExists)
				return
			}
			loginForm.IsRegister = true
		}
	}

	//3。如果是没注册过的 生成数据 到数据库注册
	if loginForm.IsRegister {
		//生成用户id
		loginForm.SnowId, err = cache.CreatUserID()
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		// 获取加密过的密码
		hash, _err := encrypt.BcryptPassword(loginForm.Password)
		if _err != nil {
			_err = errors.Wrap(_err, "HandleLoginOrRegister: BcryptPassword get fail")
			ResponseErrorAndLog(ctx, CodeServerBusy, _err)
			return
		}
		loginForm.Password = hash
		//生成用户名
		loginForm.UserName = fmt.Sprintf("纸上_%s", encrypt.RandStr(9, "all"))
		// 创建用户数据 到数据库
		_err = mysql.CreatUserInfo(&loginForm)
		if _err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, _err)
			return
		}
		userInfo.SnowId = strconv.FormatInt(loginForm.SnowId, 10)
		userInfo.UserName = loginForm.UserName
		userInfo.Phone = loginForm.Phone
		userInfo.Password = loginForm.Password
	}

	// 被封号用户
	if userInfo.Forbid == 1 {
		ResponseError(ctx, CodeUserForbidLogin)
		return
	}

	//将数据传给下一个 handle
	ctx.Set("userInfo", userInfo)
}

// HandleEmailLogin 处理邮件验证码登陆
func HandleEmailLogin(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("loginForm")
	loginForm, _ := ctxData.(m.UserEmailLogin)

	// 1。验证验证码
	code, err := cache.GetEmailVerifyCode(loginForm.Email)
	if err != nil {
		err = errors.New("HandleLoginOrRegister redis get code fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 不一致返回错误
	if code != loginForm.VerifyCode {
		ResponseError(ctx, CodeCodeVerifyError)
		return
	}

	//	查找登陆用户
	userInfo, err := mysql.GetUserByEmail(loginForm.Email)
	if err != nil {
		// 如果没有查询到用户 返回错误信息
		if errors.Cause(err) == sql.ErrNoRows {
			ResponseError(ctx, CodeUserDoseNotExists)
			return
		}
	}

	// 被封号用户
	if userInfo.Forbid == 1 {
		ResponseError(ctx, CodeUserForbidLogin)
		return
	}

	ctx.Set("userInfo", userInfo)
}

// GetBindingInfo 获取账号绑定相关信息
func GetBindingInfo(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	binding, err := mysql.GetBindingInfo(userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	binding.Phone = binding.Phone[:3] + "****" + binding.Phone[7:]

	if binding.Email != nil {
		strList := strings.Split(*binding.Email, "@")
		head := strList[0][0:2] + "****@" + strList[1]
		binding.Email = &head
	}
	if binding.Password != "" {
		binding.Password = "have"
	}

	ResponseSuccess(ctx, binding)
}

// GetAuthToken 验证安全手机验证码
func GetAuthToken(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("code")
	code := ctxData.(string)

	// 1。验证验证码
	cacheCode, err := cache.GetPhoneVerifyCode(userInfo.Phone)
	if err != nil {
		err = errors.New("GetAuthToken redis get code fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 不一致返回错误
	if cacheCode != code {
		ResponseError(ctx, CodeCodeVerifyError)
		return
	}

	token, err := jwt.CreateAccessToken(&userInfo, time.Minute*10)
	if err != nil {
		err = errors.Wrap(err, "GetAuthToken: create token fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"token": token,
	})
}

// ChangBindingEmail 修改绑定邮箱
func ChangBindingEmail(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("email")
	emailForm, _ := ctxData.(m.ChangeEmailForm)

	// 1。验证验证码
	code, err := cache.GetEmailVerifyCode(emailForm.Email)
	if err != nil {
		err = errors.New("ChangBindingEmail redis get code fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 不一致返回错误
	if code != emailForm.Code {
		ResponseError(ctx, CodeCodeVerifyError)
		return
	}

	userInfo.Email = emailForm.Email
	refreshToken, err := jwt.CreateRefreshToken(&userInfo)
	accessToken, err := jwt.CreateAccessToken(&userInfo, time.Minute*15)
	if err != nil {
		err = errors.Wrap(err, "SingIn: create token fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//保存 tokenMd5
	err = cache.SetTokenMd5(userInfo)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	err = mysql.ChangeBindingEmail(userInfo.Id, emailForm.Email)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	strList := strings.Split(emailForm.Email, "@")
	head := strList[0][0:2] + "****@" + strList[1]
	emailForm.Email = head

	ResponseSuccess(ctx, gin.H{
		"email":        emailForm.Email,
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	})

}

// ChangePassword 修改密码
func ChangePassword(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("password")
	form, _ := ctxData.(m.ChangePasswordForm)

	// 获取加密过的密码
	hash, err := encrypt.BcryptPassword(form.Password)
	if err != nil {
		err = errors.Wrap(err, "ChangePassword: BcryptPassword get fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	userInfo.MD5 = encrypt.CreatMd5(hash)
	refreshToken, err := jwt.CreateRefreshToken(&userInfo)
	accessToken, err := jwt.CreateAccessToken(&userInfo, time.Minute*15)
	if err != nil {
		err = errors.Wrap(err, "SingIn: create token fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	//保存 tokenMd5
	err = cache.SetTokenMd5(userInfo)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	err = mysql.ChangePassword(userInfo.Id, hash)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"password":     true,
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	})
}

// ChangeBindingPhone 修改绑定手机
func ChangeBindingPhone(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("phone")
	phoneForm, _ := ctxData.(m.ChangePhoneForm)

	// 1。验证验证码
	code, err := cache.GetPhoneVerifyCode(phoneForm.Phone)
	if err != nil {
		err = errors.New("ChangeBindingPhone redis get code fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 不一致返回错误
	if code != phoneForm.Code {
		ResponseError(ctx, CodeCodeVerifyError)
		return
	}

	userInfo.Phone = phoneForm.Phone
	//手机修改后 token 携带有数据 也要修改
	refreshToken, err := jwt.CreateRefreshToken(&userInfo)
	accessToken, err := jwt.CreateAccessToken(&userInfo, time.Minute*15)
	if err != nil {
		err = errors.Wrap(err, "ChangeBindingPhone: create token fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//保存 tokenMd5
	err = cache.SetTokenMd5(userInfo)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//数据库中修改
	err = mysql.ChangePhone(userInfo.Id, phoneForm.Phone)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	phone := phoneForm.Phone[:3] + "****" + phoneForm.Phone[7:]
	ResponseSuccess(ctx, gin.H{
		"phone":        phone,
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	})
}

// GetUserWxId 获取用户UnionId
func GetUserWxId(ctx *gin.Context) {
	ctxData, _ := ctx.Get("code")
	code, _ := ctxData.(string)

	url := "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
	url = fmt.Sprintf(url, settings.Conf.MiniAppID, settings.Conf.MiniAppSecret, code)

	// 发起请求
	res, _ := http.Get(url)

	// 成功后获取openId
	wxRes := m.WXIdRes{}
	json.NewDecoder(res.Body).Decode(&wxRes)
	if wxRes.ErrCode != 0 {
		ResponseError(ctx, CodeGetOpenIdFail)
		return
	}

	ResponseSuccess(ctx, gin.H{"unionId": wxRes.UnionId})
}

// WxGetUserPhone 微信获取手机号
func WxGetUserPhone(ctx *gin.Context) {
	ctxData, _ := ctx.Get("code")
	code, _ := ctxData.(string)

	token, err := cache.GetWxToken()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	var reqMap = make(map[string]interface{}, 0)
	reqMap["code"] = code
	jsonData, _ := json.Marshal(reqMap)

	url := "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s"
	url = fmt.Sprintf(url, token)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		err = errors.Wrap(err, "get WxGetUserPhone NewRequest fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	defer resp.Body.Close()

	// 成功后获取openId
	wxRes := m.WxGetUserPhoneRes{}
	err = json.NewDecoder(resp.Body).Decode(&wxRes)
	if wxRes.Errcode != 0 {
		err = errors.Wrap(err, "get wx phone fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ctx.Set("phone", wxRes.PhoneInfo.PurePhoneNumber)
}

// HandleWxLogin 微信登陆
func HandleWxLogin(ctx *gin.Context) {
	ctxData, _ := ctx.Get("phone")
	phone, _ := ctxData.(string)

	//2。通过手机号判断用户是否注册过
	userInfo, err := mysql.GetUserByPhone(phone)
	// 如果没有查询到用户 则没有注册过
	if errors.Cause(err) == sql.ErrNoRows {
		loginForm := m.LoginForm{}
		loginForm.Phone = phone
		//生成用户id
		loginForm.SnowId, err = cache.CreatUserID()
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		loginForm.Password = ""
		//生成用户名
		loginForm.UserName = fmt.Sprintf("纸上_%s", encrypt.RandStr(9, "all"))
		// 创建用户数据 到数据库
		_err := mysql.CreatUserInfo(&loginForm)
		if _err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, _err)
			return
		}
		userInfo.SnowId = strconv.FormatInt(loginForm.SnowId, 10)
		userInfo.UserName = loginForm.UserName
		userInfo.Phone = loginForm.Phone
		userInfo.Password = loginForm.Password
	}

	// 被封号用户
	if userInfo.Forbid == 1 {
		ResponseError(ctx, CodeUserForbidLogin)
		return
	}

	//将数据传给下一个 handle
	ctx.Set("userInfo", userInfo)
}

// GetWxUrl 获取微信小程序跳转url
func GetWxUrl(ctx *gin.Context) {
	token, err := cache.GetWxToken()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	url := "https://api.weixin.qq.com/wxa/generatescheme?access_token=%s"
	url = fmt.Sprintf(url, token)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		err = errors.Wrap(err, "get GetWxUrl NewRequest fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	defer resp.Body.Close()

	// 成功后获取openId
	wxRes := m.WxUrl{}
	err = json.NewDecoder(resp.Body).Decode(&wxRes)
	if wxRes.Errcode != 0 {
		err = errors.Wrap(err, "get wx GetWxUrl fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{"url": wxRes.Openlink})
}
