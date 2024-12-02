package verify

import (
	"github.com/pkg/errors"
	"onpaper-api-go/models"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// EmailRule 验证邮箱格式
func EmailRule(email string) (isPass bool, err error) {
	emailRule := "^([A-Za-z0-9_\\-\\.])+\\@([A-Za-z0-9_\\-\\.])+\\.([A-Za-z]{2,4})$"
	// 长度验证
	if utf8.RuneCountInString(email) > 50 {
		return false, errors.New("email length more than 50")
	}
	//通过正则判断 注册的用户信息是否符合格式
	isPass, err = regexp.MatchString(emailRule, email)
	if err != nil {
		return
	}
	return
}

// PhoneRule 验证手机格式
func PhoneRule(phone string) (isPass bool, err error) {
	const phoneRule = "^(?:\\+?86)?1(?:3\\d{3}|5[^4\\D]\\d{2}|8\\d{3}|7(?:[0-35-9]\\d{2}|4(?:0\\d|1[0-2]|9\\d))|9[0-35-9]\\d{2}|6[2567]\\d{2}|4[579]\\d{2})\\d{6}$"
	//通过正则判断 注册的用户信息是否符合格式
	isPass, err = regexp.MatchString(phoneRule, phone)
	return
}

// UserNameRule 验证用户名格式
func UserNameRule(name string) (isPass bool, err error) {
	nameRule := "^[\u4e00-\u9fa5A-Za-z0-9-_\uAC00-\uD7A3\u0800-\u4e00]{2,12}$"
	//通过正则判断 注册的用户信息是否符合格式
	isPass, err = regexp.MatchString(nameRule, name)
	if err != nil {
		return
	}
	return
}

// DateRule 验证 2022-02-02 日期格式
func DateRule(date string) (isPass bool, err error) {
	// 通过 time.Parse() 验证上传的日期 这样连 闰年都能判断
	_, err = time.Parse("2006-01-02", date)
	if err != nil {
		err = errors.Wrap(err, "regexp.DateRule： DateVerify fail")
		return false, err
	}
	return true, nil
}

// PasswordRule 验证密码格式
func PasswordRule(str string) bool {
	var (
		isUpper  = false
		isLower  = false
		isNumber = false
		//isSpecial = false
		minLen = 9
		maxLen = 16
	)
	length := utf8.RuneCountInString(str)
	// 验证长度 9-16之间
	if length < minLen || length > maxLen {
		return false
	}

	for _, s := range str {
		switch {
		//  验证是否有大写
		case unicode.IsUpper(s):
			isUpper = true
		//  验证是否有小写
		case unicode.IsLower(s):
			isLower = true
		//  验证是否有数字
		case unicode.IsNumber(s):
			isNumber = true
		//  验证是否有特殊符号
		//case unicode.IsPunct(s) || unicode.IsSymbol(s):
		//	isSpecial = true
		default:
		}
	}
	// 必须有字母 和 数字
	if (isLower || isUpper) && isNumber {
		return true
	}
	return false
}

// SexRule  验证性别格式
func SexRule(sex string) (isPass bool) {
	// 验证性别 是否规范
	sexVerify := false
	sexList := []string{"man", "woman", "privacy"}
	for _, s := range sexList {
		if sex == s {
			sexVerify = true
		}
	}

	return sexVerify
}

// SrtSliceListLen 验证字符串切割成数组的长度
func SrtSliceListLen(str string, count int, oneLen int, canNone bool) (isPass bool, strList []string) {
	strList = strings.Split(str, ",")
	// 验证上传的数值长度,最长只有3级
	listLen := len(strList)
	if listLen <= count {
		isPass = true
	}
	// 单个片段长度
	for _, s := range strList {
		if utf8.RuneCountInString(s) > oneLen {
			return
		}
	}

	if str == "" && canNone {
		isPass = true
	}
	return
}

// ExceptWorkType 验证期待工作类型
func ExceptWorkType(work string) (isPass bool) {
	//验证是否符合 这三个词语
	isPass = false
	workList := []string{"全职工作", "约稿创作", "项目外包", "暂不考虑"}
	for _, s := range workList {
		if work == s {
			isPass = true
		}
	}
	return
}

// SnsLinkRule 验证sns格式
func SnsLinkRule(snsLink models.SnsLinkData) (isPass bool) {
	qq := snsLink.QQ
	weibo := snsLink.Weibo
	twitter := snsLink.Twitter
	pixiv := snsLink.Pixiv
	bilibili := snsLink.Bilibili
	wechat := snsLink.WeChat

	someSnsList := []string{qq, weibo, pixiv, bilibili}
	for _, s := range someSnsList {
		textLen := utf8.RuneCountInString(s)
		//如果以上有任何一个大于长度 直接false
		if textLen > 15 {
			isPass = false
			return
		}
	}
	//单独验证 微信和推特
	textLen := utf8.RuneCountInString(wechat)
	if textLen > 20 {
		isPass = false
		return
	}

	textLen = utf8.RuneCountInString(twitter)
	if textLen > 30 {
		isPass = false
		return
	}

	return true
}

// ArtTextInfo 验证作品信息文本
func ArtTextInfo(title, intro string, tags []string) (isPass bool) {
	// 验证文本长度
	titleLen := utf8.RuneCountInString(title)
	descriptionLen := utf8.RuneCountInString(intro)
	tagsLen := len(tags)

	//标体不超过15个字 描述不超过350个字 标签个数不超过10
	if (titleLen > 15 || titleLen == 0) ||
		descriptionLen > 350 ||
		(tagsLen > 10 || tagsLen == 0) {
		return
	}

	//每一个tag 文本长度不超过20
	for _, tag := range tags {
		tagLen := utf8.RuneCountInString(tag)
		if tagLen > 20 {
			return
		}
	}

	isPass = true
	return
}

// ArtZoneText 验证作品分区文本
func ArtZoneText(text string) (isPass bool) {
	zoneList := []string{"全站", "插画", "同人", "原画", "古风", "日系", "头像", "厚涂", "虚拟主播", "Q版", "场景", "立绘", "自设/OC", "素描"}
	for _, s := range zoneList {
		if text == s {
			isPass = true
		}
	}
	return
}

// WhoSee 验证作品 浏览权限
func WhoSee(text string) (isPass bool) {
	//验证whoSee参数
	seeList := []string{"public", "onlyFans", "privacy"}
	for _, s := range seeList {
		if text == s {
			isPass = true
		}
	}
	return
}

// SixNumCode 六位数字验证码
func SixNumCode(code string) (isPass bool, err error) {
	const codeRule = "^\\d{6}$"
	//通过正则判断 注册的用户信息是否符合格式
	isPass, err = regexp.MatchString(codeRule, code)
	return
}

// InviteCode 验证邀请码
func InviteCode(code string) (isPass bool, err error) {
	const codeRule = "^[a-zA-Z0-9]{7}$"
	//通过正则判断 注册的用户信息是否符合格式
	isPass, err = regexp.MatchString(codeRule, code)
	return
}

// FileTypeList 验证图片格式数组
func FileTypeList(list []string) bool {
	validValues := map[string]struct{}{"JPG": {}, "PNG": {}, "PSD": {}, "AI": {}, "SVG": {}, "GIF": {}, "TIF": {}}
	for _, v := range list {
		if _, ok := validValues[v]; !ok {
			return false
		}
	}
	return true
}
