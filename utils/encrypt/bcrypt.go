package encrypt

import (
	cMd5 "crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"onpaper-api-go/utils/encrypt/md5"
	"sort"
)

const salt = "572d2d62c3a8aeff9389f0aee71fb2d4"

// BcryptPassword 加密密码 salt 是用户的id
func BcryptPassword(password string) (hashPassword string, err error) {
	// 在密码前后添加盐值
	saltedPassword := salt + password + salt
	// 使用 SHA512 算法对密码进行散列
	hash512 := sha512.New()
	hash512.Write([]byte(saltedPassword))

	// 传入byte密码切片，默认盐
	hash, err := bcrypt.GenerateFromPassword(hash512.Sum(nil), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	hashPassword = string(hash)
	return
}

// CompareBcrypt hash 和密码对比
func CompareBcrypt(hash string, password string) bool {
	// 在密码前后添加盐值
	saltedPassword := salt + password + salt
	// 使用 SHA512 算法对密码进行散列
	hash512 := sha512.New()
	hash512.Write([]byte(saltedPassword))

	// 成功时 err =nil  错误时 返回err
	err := bcrypt.CompareHashAndPassword([]byte(hash), hash512.Sum(nil))
	if err != nil {
		return false
	}
	return true
}

func CompareSignParams(params map[string]interface{}, filter string) string {

	// 将请求参数的key提取出来，并排好序
	newKeys := make([]string, 0)
	for k, _ := range params {
		//需要过滤的签名
		if k == filter || k == "sign" {
			continue
		}
		newKeys = append(newKeys, k)
	}
	sort.Strings(newKeys)
	var originStr string
	// 将输入进行标准化的处理
	for _, v := range newKeys {
		originStr += fmt.Sprintf("%v=%v&", v, params[v])
	}

	key := "72b729eed6dea2cc44c1dcebd0d98909a9f5b556"
	signBytes := md5.Sum([]byte(key))
	originStr += fmt.Sprintf("key=%x", signBytes)
	//fmt.Println(originStr)
	// 计算md5值
	sign2Bytes := md5.Sum([]byte(originStr))
	lastSign := fmt.Sprintf("%x", sign2Bytes)

	return lastSign
}

func CreatMd5(str string) string {
	d := []byte(str)
	m := cMd5.New()
	m.Write(d)

	return hex.EncodeToString(m.Sum(nil))
}
