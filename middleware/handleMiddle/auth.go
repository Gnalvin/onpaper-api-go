package handleMiddle

import (
	"database/sql"
	"encoding/json"
	c "onpaper-api-go/cache"
	ctl "onpaper-api-go/controller"
	"onpaper-api-go/dao/mysql"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/encrypt"
	"onpaper-api-go/utils/jwt"
	"onpaper-api-go/utils/verify"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

func VerifyJsonSign(ctx *gin.Context) {
	data, _ := ctx.GetRawData()
	var body map[string]interface{}
	err := json.Unmarshal(data, &body)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	sign := encrypt.CompareSignParams(body, "")
	if sign != body["sign"] {
		ctl.ResponseError(ctx, ctl.CodeSignFail)
		return
	}
}

func VerifyQuerySign(ctx *gin.Context) {
	queryMap := make(map[string]interface{})
	for key, value := range ctx.Request.URL.Query() {
		queryMap[key] = value[0]
	}

	_, ok := queryMap["timestamp"].(string)
	if !ok {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	timestamp, err := strconv.ParseInt(queryMap["timestamp"].(string), 10, 64)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	now := time.Now().UnixMilli()
	if now-timestamp > 30000 {
		ctl.ResponseError(ctx, ctl.CodeTimeout)
		return
	}

	sign := encrypt.CompareSignParams(queryMap, "")
	if sign != queryMap["sign"] {
		ctl.ResponseError(ctx, ctl.CodeSignFail)
		return
	}

}

// VerifyLogin 登陆验证接口
// 登陆时 验证用户名 密码是否符合格式
func VerifyLogin(ctx *gin.Context) {
	// 1.获取用户和密码
	var loginInfo m.UserLoginInfo
	err := ctx.ShouldBindJSON(&loginInfo)
	if err != nil {
		//2.判断用户名和密码不能为空,且数据类型格式正确
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	// 3.判断用户是否存在
	const emailRule = "^([A-Za-z0-9_\\-\\.])+\\@([A-Za-z0-9_\\-\\.])+\\.([A-Za-z]{2,4})$"
	const phoneRule = "^(?:\\+?86)?1(?:3\\d{3}|5[^4\\D]\\d{2}|8\\d{3}|7(?:[0-35-9]\\d{2}|4(?:0\\d|1[0-2]|9\\d))|9[0-35-9]\\d{2}|6[2567]\\d{2}|4[579]\\d{2})\\d{6}$"

	// 通过正则判断 是手机登录还是 邮箱登录
	emailVerify, err := regexp.MatchString(emailRule, loginInfo.Account)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： emailVerify fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}
	phoneVerify, err := regexp.MatchString(phoneRule, loginInfo.Account)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString：nameVerify fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	if !emailVerify && !phoneVerify {
		// 都不符合正则 说明不可能在数据库里  直接返回信息
		ctl.ResponseError(ctx, ctl.CodePasswordIsIncorrect)
		return
	}

	var userInfo m.UserTableInfo
	// 调用手机查询
	if phoneVerify {
		userInfo, err = mysql.GetUserByPhone(loginInfo.Account)
		if err != nil {
			// 如果没有查询到用户 返回错误信息
			if errors.Cause(err) == sql.ErrNoRows {
				ctl.ResponseError(ctx, ctl.CodeUserDoseNotExists)
				return
			} else {
				ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
				return
			}
		}
	}
	// 如果是 邮箱 调用邮箱查询
	if emailVerify {
		userInfo, err = mysql.GetUserByEmail(loginInfo.Account)
		if err != nil {
			// 如果没有查询到用户 返回错误信息
			if errors.Cause(err) == sql.ErrNoRows {
				ctl.ResponseError(ctx, ctl.CodeUserDoseNotExists)
				return
			} else {
				ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
				return
			}
		}
	}

	// 被封号用户
	if userInfo.Forbid == 1 {
		ctl.ResponseError(ctx, ctl.CodeUserForbidLogin)
		return
	}

	//4.判断密码是否一致
	isRight := encrypt.CompareBcrypt(userInfo.Password, loginInfo.Password)
	if !isRight {
		ctl.ResponseError(ctx, ctl.CodePasswordIsIncorrect)
	}

	//5.将数据传给下一个 handle
	ctx.Set("userInfo", userInfo)
}

// VerifyAuthMust 通过验证token 判断是否登陆，token错误 直接拒绝
func VerifyAuthMust(ctx *gin.Context) {
	// 获取请求头携带的 Authorization
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		// 如果没有携带 Authorization
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	// 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	// 如果 Authorization 不符合格式
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}
	// 解析验证token 并取出里面的payload
	userInfo, err := jwt.ParseToken(parts[1])
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}
	//如果这次携带的token 不是 accessToken 则拒绝
	if userInfo.TokenType != "AccessToken" {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	ctx.Set("userInfo", userInfo)
}

// VerifyAuth 通过验证token 查看是哪个用户登录，没有token 也不会拒绝请求
func VerifyAuth(ctx *gin.Context) {
	//初始化 payload 游客都为空
	userInfo := m.UserTokenPayload{
		Id:        "",
		Phone:     "",
		TokenType: "",
	}

	// 获取请求头携带的 Authorization
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		// 如果没有携带 Authorization 按游客处理
		ctx.Set("userInfo", userInfo)
		return
	}

	// 如果有 token 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	// 如果 Authorization 不符合格式
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		//按游客处理
		ctx.Set("userInfo", userInfo)
		return
	}

	userInfo, err := jwt.ParseToken(parts[1])
	// 带了token 验证是失败的 则返回错误
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}
	//如果这次携带的token 不是 accessToken 则拒绝
	if userInfo.TokenType != "AccessToken" {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	// token验证正确 则 将用户信息 存context
	ctx.Set("userInfo", userInfo)
}

// VerifyRefreshToken 用于验证 RefreshToken
func VerifyRefreshToken(ctx *gin.Context) {
	// 获取请求头携带的 Authorization
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		// 如果没有携带 Authorization
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	// 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	// 如果 Authorization 不符合格式
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}
	// 解析验证token 并取出里面的payload
	userInfo, err := jwt.ParseToken(parts[1])
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	//如果这次携带的token 不是 RefreshToken 则拒绝
	if userInfo.TokenType != "RefreshToken" {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	// 获取token 的md5 验证邮箱性
	cacheMd5, err := c.GetTokenMd5(userInfo.Id)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeServerBusy)
		return
	}

	md5 := encrypt.CreatMd5(userInfo.MD5 + userInfo.Phone + userInfo.Email)
	if md5 != cacheMd5 {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	ctx.Set("userInfo", userInfo)
}

// VerifyEmailFormat 认证邮箱格式
func VerifyEmailFormat(ctx *gin.Context) {
	// 1.获取需要验证的邮箱
	var emailData m.VerifyEmail
	err := ctx.ShouldBindQuery(&emailData)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	//2.验证邮箱格式
	isPass, err := verify.EmailRule(emailData.Email)
	if err != nil || !isPass {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctx.Set("email", emailData.Email)
}

// VerifyPhoneFormat 认证手机格式
func VerifyPhoneFormat(ctx *gin.Context) {
	// 1.获取需要验证的手机
	var phoneData m.VerifyPhone
	err := ctx.ShouldBindQuery(&phoneData)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	//2.验证手机格式
	isPass, err := verify.PhoneRule(phoneData.Phone)
	if err != nil || !isPass {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctx.Set("phone", phoneData)
}

// VerifyEmailLogin 验证邮箱验证码登陆
func VerifyEmailLogin(ctx *gin.Context) {
	var loginForm m.UserEmailLogin
	// 1.绑定body json 信息到 loginForm
	err := ctx.ShouldBindJSON(&loginForm)
	if err != nil {
		ctl.ResponseErrorAndLog(ctx, ctl.CodeJsonFormatError, err)
		return
	}

	emailVerify, err := verify.EmailRule(loginForm.Email)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： EmailRule fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}
	// 验证码格式
	codeVerify, err := verify.SixNumCode(loginForm.VerifyCode)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： SixNumCode fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	if !codeVerify || !emailVerify {
		// 只要有一个 不符合格式不许注册
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 将参数传递到上下文中
	ctx.Set("loginForm", loginForm)
}

// VerifySafetyCode 验证手机验证码
func VerifySafetyCode(ctx *gin.Context) {
	var codeData m.VerifySafetyCode
	err := ctx.ShouldBindQuery(&codeData)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 验证码格式
	isPass, err := verify.SixNumCode(codeData.Code)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： SixNumCode fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	if !isPass {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctx.Set("code", codeData.Code)
}

// VerifyChangeEmail 验证邮箱换绑定
func VerifyChangeEmail(ctx *gin.Context) {
	var data m.ChangeEmailForm
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 验证码格式
	vCode, err := verify.SixNumCode(data.Code)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： SixNumCode fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	//2.验证邮箱格式
	vEmail, err := verify.EmailRule(data.Email)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： EmailRule fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}
	if !vEmail || !vCode {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	_, err = jwt.ParseToken(data.Token)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	ctx.Set("email", data)

}

// VerifyChangePassword 验证修改密码参数
func VerifyChangePassword(ctx *gin.Context) {
	var data m.ChangePasswordForm
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 验证码格式
	isPass := verify.PasswordRule(data.Password)
	if !isPass {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	//验证token
	_, err = jwt.ParseToken(data.Token)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}
	ctx.Set("password", data)
}

// VerifyChangePhone 验证修改手机参数
func VerifyChangePhone(ctx *gin.Context) {
	var data m.ChangePhoneForm
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 验证码格式
	vCode, err := verify.SixNumCode(data.Code)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： SixNumCode fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	//2.验证手机格式
	vPhone, err := verify.PhoneRule(data.Phone)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： PhoneRule fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	if !vPhone || !vCode {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	//验证token
	_, err = jwt.ParseToken(data.Token)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeUnAuthorization)
		return
	}

	ctx.Set("phone", data)
}

// VerifyMiniProgramCode 验证小程序code
func VerifyMiniProgramCode(ctx *gin.Context) {
	var data m.MiniProgramCode
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctx.Set("code", data.Code)
}
