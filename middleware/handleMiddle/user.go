package handleMiddle

import (
	ctl "onpaper-api-go/controller"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/verify"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

// VerifyAccountForm 验证参数合法性
func VerifyAccountForm(ctx *gin.Context) {
	var loginForm m.LoginForm
	// 1.绑定body json 信息到 loginForm
	err := ctx.ShouldBindJSON(&loginForm)
	if err != nil {
		ctl.ResponseErrorAndLog(ctx, ctl.CodeJsonFormatError, err)
		return
	}

	// 2.通过正则判断 注册的用户信息是否符合格式
	phoneVerify, err := verify.PhoneRule(loginForm.Phone)
	if err != nil {
		err = errors.Wrap(err, "regexp.MatchString： phoneVerify fail")
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

	// 通过函数验证密码
	passwordVerify := true
	if loginForm.Password != "" {
		passwordVerify = verify.PasswordRule(loginForm.Password)
	}

	// 验证邀请码
	inviteCodeVerify := true
	if loginForm.InviteCode != "" {
		inviteCodeVerify, err = verify.InviteCode(loginForm.InviteCode)
		if err != nil {
			err = errors.Wrap(err, "regexp.MatchString： Code fail")
			ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
			return
		}
	}

	if !phoneVerify || !codeVerify || !passwordVerify || !inviteCodeVerify {
		// 只要有一个 不符合格式不许注册
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	// 将参数传递到上下文中
	ctx.Set("loginForm", loginForm)
}

// VerifyQueryUserId 处理用户首页资料请求的数据
func VerifyQueryUserId(ctx *gin.Context) {
	var data m.VerifyUserId
	// 验证 url参数
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("userId", data.UserId)

}

// VerifyUpdateProfileData 用户上传资料时处理验证
func VerifyUpdateProfileData(ctx *gin.Context) {
	var data m.UpdateProfileData
	// 1.绑定body json 信息到 data
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	isPass := false
	switch data.ProfileType {
	case "userName":
		// 验证用户名是否规范
		isPass, _ = verify.UserNameRule(data.Profile)
	case "sex":
		// 验证用户名性别
		isPass = verify.SexRule(data.Profile)
	case "birthday":
		isPass, _ = verify.DateRule(data.Profile)
	case "workEmail":
		isPass, _ = verify.EmailRule(data.Profile)
		if data.Profile == "" {
			isPass = true
		}
	case "region":
		isPass, _ = verify.SrtSliceListLen(data.Profile, 3, 10, false)
	case "createStyle":
		isPass, _ = verify.SrtSliceListLen(data.Profile, 3, 10, true)
	case "software":
		isPass, _ = verify.SrtSliceListLen(data.Profile, 3, 10, true)
	case "exceptWork":
		isPass = verify.ExceptWorkType(data.Profile)
	case "introduce":
		textLen := utf8.RuneCountInString(data.Profile)
		if textLen <= 800 {
			isPass = true
		}
	case "snsLink":
		isPass = verify.SnsLinkRule(data.SnsData)
	default:
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 如果不通过
	if !isPass {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	// 把它传递到上下文
	ctx.Set("profile", data)
}

// VerifyUserAndPage 验证 携带查询用户和分页
func VerifyUserAndPage(ctx *gin.Context) {
	// 1. 验证参数
	var data m.VerifyUserAndPage
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("query", data)
}

func VerifyUserFollow(ctx *gin.Context) {
	// 1. 验证参数
	var data m.VerifyUserFollow
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("query", data)
}

// VerifyUserHomeArtwork 用户首页展示的作品
func VerifyUserHomeArtwork(ctx *gin.Context) {
	// 1. 验证参数
	var data m.VerifyUserHomeArtwork
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("query", data)
}

// VerifyUserFocus 处理用户关注用户请求
func VerifyUserFocus(ctx *gin.Context) {
	var data m.VerifyUserFocus
	// 1.绑定body json 信息到 data
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 把它传递到上下文
	ctx.Set("Focus", data)
}

// VerifyUserRankRequest 处理用户排名请求
func VerifyUserRankRequest(ctx *gin.Context) {
	var data m.QueryUserRank
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	// 把它传递到上下文
	ctx.Set("rankType", data)
}

// HandlePostInteract 处理上传的 对作品点赞、收藏信息
func HandlePostInteract(ctx *gin.Context) {
	var data m.PostInteractData
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 把它传递到上下文
	ctx.Set("interact", &data)
}

// HandleUserFeedback 处理反馈
func HandleUserFeedback(ctx *gin.Context) {
	var feedback m.PostFeedback
	err := ctx.ShouldBindJSON(&feedback)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	if utf8.RuneCountInString(feedback.Describe) <= 5 {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 把它传递到上下文
	ctx.Set("feedback", feedback)
}

// VerifyAllUserShow 验证全站用户展示参数
func VerifyAllUserShow(ctx *gin.Context) {
	var query m.AllUserShowQuery
	err := ctx.ShouldBindQuery(&query)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("query", query)
}
