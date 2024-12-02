package SendEmail

import (
	"fmt"
	"os"
	"path/filepath"
)

var AuthCodeTemp = ``

// ReadAuthCodeHtml 获取邮箱验证码的html模版
func ReadAuthCodeHtml() {
	//文件的绝对路径
	absolutePath, _ := filepath.Abs("assets/html/auth_code.html")
	b, err := os.ReadFile(absolutePath) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	AuthCodeTemp = string(b)
}
