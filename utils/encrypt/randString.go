package encrypt

import (
	"github.com/google/uuid"
	"math/rand"
	"strings"
	"time"
	"unsafe"
)

// 没有OoiIlL
const upperAndLower = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ"

const num = "1234567890"

// no 0 num
const numNoZero = "123456789"

// 没有IL
const upper = "ABCDEFGHJKMNOPQRSTUVWXYZ"

// 没有0OoiIlL
const all = "123456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

// RandStr 随机生成字符串
func RandStr(n int, codeType string) string {
	b := make([]byte, n)
	randRawStr := ""
	switch codeType {
	case "num":
		randRawStr = num
	case "upperAndLower":
		randRawStr = upperAndLower
	case "upper":
		randRawStr = upper
	case "all":
		randRawStr = all
	case "numNoZero":
		randRawStr = numNoZero
	}
	// A rand.Int63() generates 63 random bits, enough for letterIdMax UpperAndLower!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(randRawStr) {
			b[i] = randRawStr[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

// CreateUUID 生成uuid
func CreateUUID() (id string) {
	// V4 基于随机数
	u4 := uuid.New()
	return strings.Replace(u4.String(), "-", "", -1)
}
