package controller

type ResCode int64

// 定义错误类型
const (
	CodeSuccess         ResCode = 200
	CodeJsonFormatError ResCode = 1000 + iota
	CodeUserAlreadyExists
	CodeUserDoseNotExists
	CodePasswordIsIncorrect
	CodeUnAuthorization
	CodeNoHaveToken
	CodeUnPermission
	CodeParamsError
	CodeEmailExists
	CodeEmailNoExists
	CodeServerBusy
	CodeUploadError
	CodeUploadTooManyFiles
	CodeUploadFilesTooLarge
	CodeUploadNotAImg
	CodeUserNoHaveArtworks
	CodeCommentNoHave
	CodeCodeVerifyError
	CodeArtworkNoExists
	CodeLongTimeNoOperate
	CodeTrendNoExists
	CodePhoneExists
	CodePhoneNoExists
	CodeNeedInviteCode
	CodeInviteCodeInvalid
	CodeOnlyHeFocusUserCanDo
	CodeUserNoHaveFocus
	CodeUserNoHaveCollects
	CodeTimeout
	CodeSignFail
	CodeUserForbidLogin
	CodeUserNoHaveTrends
	CodeUserNoAcceptPlan
	CodeNeedToken
	CodeUserStopCommission
	CodeGetOpenIdFail
	CodeUserNoHaveFans
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:              "success",
	CodeJsonFormatError:      "json_format_error",
	CodeUserAlreadyExists:    "user_already_exists",
	CodeUserDoseNotExists:    "user_dose_not_exists",
	CodePasswordIsIncorrect:  "password_is_incorrect",
	CodeUnAuthorization:      "un_authorization",
	CodeNoHaveToken:          "no_have_token",
	CodeUnPermission:         "un_permission",
	CodeParamsError:          "params_error",
	CodeEmailNoExists:        "email_no_exists",
	CodeEmailExists:          "email_exists",
	CodeServerBusy:           "server_busy",
	CodeUploadError:          "upload_error",
	CodeUploadTooManyFiles:   "upload_too_many_files",
	CodeUploadFilesTooLarge:  "upload_files_too_large",
	CodeUploadNotAImg:        "upload_not_a_img",
	CodeUserNoHaveArtworks:   "user_no_have_artworks",
	CodeCommentNoHave:        "comment_no_have",
	CodeCodeVerifyError:      "verifyCode_verify_error",
	CodeArtworkNoExists:      "artwork_no_exists",
	CodeLongTimeNoOperate:    "long_time_no_operate",
	CodeTrendNoExists:        "trend_no_exists",
	CodePhoneExists:          "phone_exists",
	CodePhoneNoExists:        "phone_no_exists",
	CodeNeedInviteCode:       "need_invite_code",
	CodeInviteCodeInvalid:    "invite_code_invalid",
	CodeOnlyHeFocusUserCanDo: "only_he_focus_user_can_do",
	CodeUserNoHaveFocus:      "user_no_have_focus",
	CodeUserNoHaveCollects:   "user_no_have_collects",
	CodeTimeout:              "timeout",
	CodeSignFail:             "sign_fail",
	CodeUserForbidLogin:      "user_forbid_login",
	CodeUserNoHaveTrends:     "user_no_have_trend",
	CodeUserNoAcceptPlan:     "user_no_have_accept_plan",
	CodeNeedToken:            "need_token",
	CodeUserStopCommission:   "user_stop_commission",
	CodeGetOpenIdFail:        "get_openid_fail",
	CodeUserNoHaveFans:       "user_no_have_fans",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
